package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/HexArch/go-chat/internal/services/chat/internal/config"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Server struct {
	logger     *zap.Logger
	config     *config.Config
	httpServer *http.Server
	wsHandler  *WebSocketHandler
}

func NewServer(
	logger *zap.Logger,
	config *config.Config,
	wsHandler *WebSocketHandler,
) *Server {
	return &Server{
		logger:    logger,
		config:    config,
		wsHandler: wsHandler,
	}
}

func (s *Server) Start(ctx context.Context) error {
	router := mux.NewRouter()
	router.HandleFunc("/ws/chat/{roomID}", s.wsHandler.ServeWS)

	s.httpServer = &http.Server{
		Addr:         s.config.Handlers.HTTP.FullAddress(),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
