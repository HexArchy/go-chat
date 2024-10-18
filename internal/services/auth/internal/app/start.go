package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (a *App) start(ctx context.Context) {
	go a.startHTTPServer(ctx)
}

func (a *App) startHTTPServer(ctx context.Context) {
	router := mux.NewRouter()

	a.apiHandler.AddRoutes(router)

	srv := &http.Server{
		Addr:         a.cfg.Handlers.HTTP.Address + ":" + a.cfg.Handlers.HTTP.Port,
		Handler:      router,
		ReadTimeout:  a.cfg.Handlers.HTTP.ReadTimeout,
		WriteTimeout: a.cfg.Handlers.HTTP.WriteTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		a.logger.Info("Starting HTTP server", zap.String("address", a.cfg.Handlers.HTTP.Address+":"+a.cfg.Handlers.HTTP.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	<-stop
	a.logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		a.logger.Error("Error during server shutdown", zap.Error(err))
	}

	a.logger.Info("Server gracefully stopped")
}
