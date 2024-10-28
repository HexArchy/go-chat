package controllers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/website"
	"github.com/HexArch/go-chat/internal/services/website/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/website/internal/config"
	"github.com/HexArch/go-chat/internal/services/website/internal/controllers/cache"
	"github.com/HexArch/go-chat/internal/services/website/internal/controllers/middleware"
	"github.com/HexArch/go-chat/internal/services/website/internal/metrics"
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
	logger      *zap.Logger
	cfg         *config.Config
	metrics     *metrics.WebsiteMetrics
	grpcServer  *grpc.Server
	httpServer  *http.Server
	authClient  *auth.AuthClient
	roomService *WebsiteServiceServer
	roomCache   *cache.RoomCache
	healthCheck *HealthChecker
}

func NewServer(
	logger *zap.Logger,
	cfg *config.Config,
	authClient *auth.AuthClient,
	roomService *WebsiteServiceServer,
) *Server {
	return &Server{
		logger:      logger,
		cfg:         cfg,
		metrics:     metrics.NewWebsiteMetrics("website_service"),
		authClient:  authClient,
		roomService: roomService,
		roomCache:   cache.NewRoomCache(5 * time.Minute),
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

	return group.Wait()
}

func (s *Server) startGRPCServer(ctx context.Context) error {
	addr := s.cfg.Handlers.GRPC.FullAddress()
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// Configure keepalive.
	serverParams := keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second,
		MaxConnectionAge:      30 * time.Second,
		MaxConnectionAgeGrace: 5 * time.Second,
		Time:                  5 * time.Second,
		Timeout:               1 * time.Second,
	}

	// Create middleware.
	authMiddleware := middleware.NewAuthMiddleware(s.logger, s.authClient, s.metrics)

	// Create gRPC server.
	s.grpcServer = grpc.NewServer(
		grpc.KeepaliveParams(serverParams),
		grpc.ChainUnaryInterceptor(
			s.metricsInterceptor,
			authMiddleware.UnaryInterceptor(),
			s.recoveryInterceptor,
			s.loggingInterceptor,
		),
		grpc.MaxConcurrentStreams(1000),
		grpc.MaxRecvMsgSize(20*1024*1024),
	)

	// Register services.
	website.RegisterRoomServiceServer(s.grpcServer, s.roomService)
	grpc_health_v1.RegisterHealthServer(s.grpcServer, s.healthCheck)

	go func() {
		<-ctx.Done()
		s.logger.Info("Stopping gRPC server")
		s.grpcServer.GracefulStop()
	}()

	s.logger.Info("Starting gRPC server", zap.String("address", addr))
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

	if err := website.RegisterRoomServiceHandlerFromEndpoint(ctx, mux, grpcAddr, opts); err != nil {
		return fmt.Errorf("failed to register gateway: %w", err)
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
		ReadTimeout:  s.cfg.Handlers.HTTP.ReadTimeout,
		WriteTimeout: s.cfg.Handlers.HTTP.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Failed to shutdown HTTP server", zap.Error(err))
		}
	}()

	s.logger.Info("Starting HTTP server", zap.String("address", httpAddr))
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

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Failed to shutdown metrics server", zap.Error(err))
		}
	}()

	s.logger.Info("Starting metrics server", zap.String("address", s.cfg.Engines.Metrics.Address))
	return metricsServer.ListenAndServe()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if s.healthCheck.GetStatus() == grpc_health_v1.HealthCheckResponse_SERVING {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte("NOT_SERVING"))
}

func (s *Server) Stop(ctx context.Context) error {
	s.healthCheck.SetStatus(grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	s.grpcServer.GracefulStop()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	s.roomCache.Close()

	if err := s.authClient.Close(); err != nil {
		return fmt.Errorf("failed to close auth client: %w", err)
	}

	return nil
}

// Interceptors.
func (s *Server) metricsInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	s.metrics.IncActiveRequests()
	defer s.metrics.DecActiveRequests()

	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
	}

	s.metrics.RecordRequestDuration(info.FullMethod, status, duration)
	return resp, err
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
