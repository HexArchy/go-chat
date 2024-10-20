package controllers

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/auth"
	"github.com/HexArch/go-chat/internal/services/auth/internal/config"
	"github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/validatetoken"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	grpcServer      *grpc.Server
	httpServer      *http.Server
	logger          *zap.Logger
	cfg             *config.Config
	authSvc         *AuthServiceServer
	validateTokenUC *validatetoken.UseCase
	jwtSecret       []byte
}

func NewServer(logger *zap.Logger, cfg *config.Config, authSvc *AuthServiceServer, jwtSecret []byte, validateTokenUC *validatetoken.UseCase) *Server {
	return &Server{
		logger:          logger,
		cfg:             cfg,
		authSvc:         authSvc,
		jwtSecret:       jwtSecret,
		validateTokenUC: validateTokenUC,
	}
}

func (s *Server) Start(ctx context.Context) error {
	grpcAddress := s.cfg.Handlers.GRPC.FullAddress()
	httpAddress := s.cfg.Handlers.HTTP.FullAddress()

	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		return err
	}

	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(AuthInterceptor(s.jwtSecret, s.authSvc.validateTokenUC)),
	)
	auth.RegisterAuthServiceServer(s.grpcServer, s.authSvc)

	go func() {
		s.logger.Info("Starting gRPC server", zap.String("address", grpcAddress))
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	mux := runtime.NewServeMux()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8081"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := auth.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcAddress, opts); err != nil {
		return err
	}

	s.httpServer = &http.Server{
		Addr:    httpAddress,
		Handler: handler,
	}

	go func() {
		s.logger.Info("Starting HTTP server", zap.String("address", httpAddress))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to serve HTTP", zap.Error(err))
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.grpcServer.GracefulStop()
	ctxShutdown, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(ctxShutdown); err != nil {
		return err
	}
	return nil
}
