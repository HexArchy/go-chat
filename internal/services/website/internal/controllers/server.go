package controllers

import (
	"context"
	"net"
	"net/http"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/website"
	"github.com/HexArch/go-chat/internal/services/website/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/website/internal/config"
	"github.com/HexArch/go-chat/internal/services/website/internal/controllers/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	logger      *zap.Logger
	cfg         *config.Config
	grpcServer  *grpc.Server
	httpServer  *http.Server
	authClient  *auth.AuthClient
	roomService *WebsiteServiceServer
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
		authClient:  authClient,
		roomService: roomService,
	}
}

func (s *Server) Start(ctx context.Context) error {
	// Initialize middleware.
	authMiddleware := middleware.NewAuthMiddleware(s.logger, s.authClient)

	// Create gRPC server with auth middleware.
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(authMiddleware.UnaryInterceptor()),
	)
	website.RegisterRoomServiceServer(s.grpcServer, s.roomService)

	// Start gRPC server.
	lis, err := net.Listen("tcp", s.cfg.Handlers.GRPC.FullAddress())
	if err != nil {
		return errors.Wrap(err, "failed to create listener")
	}

	go func() {
		s.logger.Info("Starting gRPC server", zap.String("address", s.cfg.Handlers.GRPC.FullAddress()))
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Configure gRPC gateway.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := website.RegisterRoomServiceHandlerFromEndpoint(
		ctx,
		mux,
		s.cfg.Handlers.GRPC.FullAddress(),
		opts,
	); err != nil {
		return errors.Wrap(err, "failed to register gRPC gateway")
	}

	// Configure CORS.
	corsHandler := cors.New(cors.Options{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Create HTTP server.
	s.httpServer = &http.Server{
		Addr:         s.cfg.Handlers.HTTP.FullAddress(),
		Handler:      corsHandler.Handler(mux),
		ReadTimeout:  s.cfg.Handlers.HTTP.ReadTimeout,
		WriteTimeout: s.cfg.Handlers.HTTP.WriteTimeout,
	}

	// Start HTTP server,
	go func() {
		s.logger.Info("Starting HTTP server", zap.String("address", s.cfg.Handlers.HTTP.FullAddress()))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to serve HTTP", zap.Error(err))
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	// Gracefully stop gRPC server.
	s.grpcServer.GracefulStop()

	// Gracefully stop HTTP server.
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to shutdown HTTP server")
	}

	// Close auth client connection.
	if err := s.authClient.Close(); err != nil {
		return errors.Wrap(err, "failed to close auth client")
	}

	return nil
}
