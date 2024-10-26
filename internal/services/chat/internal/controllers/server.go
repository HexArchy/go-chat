package controllers

import (
	"context"
	"net"
	"net/http"

	"github.com/HexArch/go-chat/internal/api/generated/go-chat/api/proto/chat"
	"github.com/HexArch/go-chat/internal/services/chat/internal/config"
	"github.com/HexArch/go-chat/internal/services/chat/internal/controllers/middleware"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	logger     *zap.Logger
	cfg        *config.Config
	grpcServer *grpc.Server
	httpServer *http.Server
	chatSvc    *ChatServiceServer
	authMW     *middleware.AuthMiddleware
	upgrader   websocket.Upgrader
}

func NewServer(
	logger *zap.Logger,
	cfg *config.Config,
	chatSvc *ChatServiceServer,
	authMW *middleware.AuthMiddleware,
) *Server {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// TODO: Implement proper origin check
			return true
		},
	}

	return &Server{
		logger:   logger,
		cfg:      cfg,
		chatSvc:  chatSvc,
		authMW:   authMW,
		upgrader: upgrader,
	}
}

func (s *Server) Start(ctx context.Context) error {
	// Setup gRPC server.
	grpcAddress := s.cfg.Handlers.GRPC.FullAddress()
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		return err
	}

	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.authMW.UnaryInterceptor()),
	)
	chat.RegisterChatServiceServer(s.grpcServer, s.chatSvc)

	go func() {
		s.logger.Info("Starting gRPC server", zap.String("address", grpcAddress))
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Setup HTTP server with gRPC-Gateway.
	grpcMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := chat.RegisterChatServiceHandlerFromEndpoint(
		ctx,
		grpcMux,
		grpcAddress,
		opts,
	); err != nil {
		return err
	}

	// Setup router with WebSocket endpoint.
	router := mux.NewRouter()
	router.HandleFunc("/ws/chat/{room_id}", s.handleWebSocket)
	router.PathPrefix("/api/").Handler(grpcMux)

	// Setup CORS.
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"localhost:8000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	// Create HTTP server.
	s.httpServer = &http.Server{
		Addr:         s.cfg.Handlers.HTTP.FullAddress(),
		Handler:      corsHandler.Handler(router),
		ReadTimeout:  s.cfg.Handlers.HTTP.ReadTimeout,
		WriteTimeout: s.cfg.Handlers.HTTP.WriteTimeout,
	}

	// Start HTTP server.
	go func() {
		s.logger.Info("Starting HTTP/WebSocket server", zap.String("address", s.cfg.Handlers.HTTP.FullAddress()))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to serve HTTP", zap.Error(err))
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	// Gracefully stop accepting new connections.
	s.logger.Info("Stopping server...")

	// Stop gRPC server.
	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	// Wait for gRPC server to stop or context to be canceled.
	select {
	case <-stopped:
		s.logger.Info("gRPC server stopped")
	case <-ctx.Done():
		s.logger.Warn("Deadline exceeded while stopping gRPC server")
		s.grpcServer.Stop()
	}

	// Shutdown HTTP server.
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Error shutting down HTTP server", zap.Error(err))
		return err
	}
	s.logger.Info("HTTP server stopped")

	return nil
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract room ID from URL.
	vars := mux.Vars(r)
	roomID := vars["room_id"]
	if roomID == "" {
		s.logger.Error("Room ID not provided in WebSocket URL")
		http.Error(w, "Room ID is required", http.StatusBadRequest)
		return
	}

	// Validate authorization header.
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		s.logger.Error("Authorization header missing in WebSocket request")
		http.Error(w, "Authorization required", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket.
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("Failed to upgrade WebSocket connection",
			zap.Error(err),
			zap.String("room_id", roomID),
		)
		return
	}

	// Create context with timeout for the WebSocket connection.
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Handle WebSocket connection.
	go func() {
		defer conn.Close()
		s.chatSvc.HandleWebSocket(ctx, conn)
	}()
}

// Health check endpoint.
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) metrics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement metrics collection.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Metrics Not Implemented"))
}
