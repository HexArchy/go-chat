package app

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (a *App) start(ctx context.Context) {
	go a.startHTTPServer(ctx)

	if err := a.grShutdown.Wait(a.cfg.GracefulShutdown); err != nil {
		a.logger.Error("Error during graceful shutdown", zap.Error(err))
	} else {
		a.logger.Info("Application gracefully stopped")
	}
}

func (a *App) startHTTPServer(ctx context.Context) {
	router := mux.NewRouter()

	a.setupRoutes(router)

	srv := &http.Server{
		Addr:         a.cfg.Handlers.HTTP.Address + ":" + a.cfg.Handlers.HTTP.Port,
		Handler:      router,
		ReadTimeout:  a.cfg.Handlers.HTTP.ReadTimeout,
		WriteTimeout: a.cfg.Handlers.HTTP.WriteTimeout,
	}

	a.grShutdown.Add(func(ctx context.Context) error {
		return srv.Shutdown(ctx)
	})

	go func() {
		a.logger.Info("Starting HTTP server",
			zap.String("address", srv.Addr),
			zap.String("readTimeout", a.cfg.Handlers.HTTP.ReadTimeout.String()),
			zap.String("writeTimeout", a.cfg.Handlers.HTTP.WriteTimeout.String()))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()
}

func (a *App) setupRoutes(router *mux.Router) {
	a.apiHandler.AddRoutes(router)

	a.logger.Info("Routes set up successfully")
}
