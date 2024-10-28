package controllers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/auth"
	"github.com/HexArch/go-chat/internal/services/auth/internal/config"
	"github.com/HexArch/go-chat/internal/services/auth/internal/controllers/cache"
	"github.com/HexArch/go-chat/internal/services/auth/internal/controllers/middleware"
	"github.com/HexArch/go-chat/internal/services/auth/internal/metrics"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/proto"
)

type Server struct {
	cfg         *config.Config
	logger      *zap.Logger
	metrics     *metrics.AuthMetrics
	middleware  *middleware.AuthMiddleware
	authCtrl    *AuthController
	usersCtrl   *UsersController
	grpcServer  *grpc.Server
	httpServer  *http.Server
	tokenCache  *cache.TokenCache
	healthCheck *HealthChecker
}

func NewServer(
	cfg *config.Config,
	logger *zap.Logger,
	authCtrl *AuthController,
	usersCtrl *UsersController,
	tokenCache *cache.TokenCache,
	metrics *metrics.AuthMetrics,
	authMiddleware *middleware.AuthMiddleware,
) *Server {
	return &Server{
		cfg:         cfg,
		logger:      logger,
		metrics:     metrics,
		middleware:  authMiddleware,
		authCtrl:    authCtrl,
		usersCtrl:   usersCtrl,
		tokenCache:  tokenCache,
		healthCheck: NewHealthChecker(),
	}
}

func (s *Server) Start(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	// Start gRPC server.
	group.Go(func() error {
		return s.startGRPCServer(ctx)
	})

	// Start HTTP server.
	group.Go(func() error {
		return s.startHTTPServer(ctx)
	})

	// Start metrics server.
	group.Go(func() error {
		return s.startMetricsServer(ctx)
	})

	// Wait for any server to exit or context cancellation.
	return group.Wait()
}

func (s *Server) startGRPCServer(ctx context.Context) error {
	addr := s.cfg.Handlers.GRPC.FullAddress()
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	serverParams := keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second,
		MaxConnectionAge:      30 * time.Second,
		MaxConnectionAgeGrace: 5 * time.Second,
		Time:                  5 * time.Second,
		Timeout:               1 * time.Second,
	}

	s.grpcServer = grpc.NewServer(
		grpc.KeepaliveParams(serverParams),
		grpc.ChainUnaryInterceptor(
			s.middleware.UnaryInterceptor(),
			s.recoveryInterceptor,
			s.loggingInterceptor,
		),
		grpc.MaxConcurrentStreams(1000),
		grpc.MaxRecvMsgSize(20*1024*1024),
	)

	auth.RegisterAuthServiceServer(s.grpcServer, s.authCtrl)
	grpc_health_v1.RegisterHealthServer(s.grpcServer, s.healthCheck)

	s.logger.Info("Starting gRPC server", zap.String("address", addr))

	go func() {
		<-ctx.Done()
		s.logger.Info("Stopping gRPC server")
		s.grpcServer.GracefulStop()
	}()

	return s.grpcServer.Serve(listener)
}

func (s *Server) startHTTPServer(ctx context.Context) error {
	grpcAddr := s.cfg.Handlers.GRPC.FullAddress()
	httpAddr := s.cfg.Handlers.HTTP.FullAddress()

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),
		runtime.WithForwardResponseOption(httpResponseModifier),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	if err := auth.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register gRPC gateway: %w", err)
	}

	corsMiddleware := cors.New(cors.Options{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"Grpc-Metadata-Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	s.httpServer = &http.Server{
		Addr:         httpAddr,
		Handler:      corsMiddleware.Handler(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	s.logger.Info("Starting HTTP server", zap.String("address", httpAddr))

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Failed to shutdown HTTP server", zap.Error(err))
		}
	}()

	return s.httpServer.ListenAndServe()
}

func (s *Server) startMetricsServer(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", s.handleHealth)

	metricsServer := &http.Server{
		Addr:    s.cfg.Engines.Metrics.Address,
		Handler: mux,
	}

	s.logger.Info("Starting metrics server", zap.String("address", s.cfg.Engines.Metrics.Address))

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Failed to shutdown metrics server", zap.Error(err))
		}
	}()

	return metricsServer.ListenAndServe()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if s.healthCheck.status == grpc_health_v1.HealthCheckResponse_SERVING {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte("NOT_SERVING"))
}

func (s *Server) recoveryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("Recovered from panic",
				zap.Any("panic", r),
				zap.String("method", info.FullMethod),
			)
			s.metrics.RecordError("panic")
		}
	}()
	return handler(ctx, req)
}

func (s *Server) loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	if info.FullMethod != "/grpc.health.v1.Health/Check" {
		s.logger.Info("Handled request",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
	}

	return resp, err
}

func customHeaderMatcher(key string) (string, bool) {
	switch key {
	case "Authorization", "Content-Type":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

func httpResponseModifier(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	return nil
}

// Stop gracefully stops all servers.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Starting graceful shutdown of all servers")

	// Update health check status.
	s.healthCheck.SetStatus(grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	// Create errgroup for concurrent shutdown.
	group, ctx := errgroup.WithContext(ctx)

	// Stop gRPC server.
	group.Go(func() error {
		s.logger.Debug("Stopping gRPC server")
		done := make(chan struct{})
		go func() {
			s.grpcServer.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			s.logger.Debug("gRPC server stopped successfully")
			return nil
		case <-ctx.Done():
			s.logger.Warn("gRPC server stop timeout exceeded")
			s.grpcServer.Stop()
			return ctx.Err()
		}
	})

	// Stop HTTP server.
	group.Go(func() error {
		s.logger.Debug("Stopping HTTP server")
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("Failed to stop HTTP server", zap.Error(err))
			return fmt.Errorf("failed to stop HTTP server: %w", err)
		}
		s.logger.Debug("HTTP server stopped successfully")
		return nil
	})

	// Close token cache.
	group.Go(func() error {
		s.logger.Debug("Closing token cache")
		s.tokenCache.Close()
		s.logger.Debug("Token cache closed successfully")
		return nil
	})

	// Wait for all shutdowns to complete or timeout.
	if err := group.Wait(); err != nil {
		return fmt.Errorf("error during shutdown: %w", err)
	}

	s.logger.Info("All servers stopped successfully")
	return nil
}
