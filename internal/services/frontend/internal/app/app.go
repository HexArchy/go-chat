package app

import (
	"context"
	"encoding/gob"
	"net/http"

	"github.com/HexArch/go-chat/internal/pkg/graceful-shutdown"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/chat"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/website"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/config"
	httpadmin "github.com/HexArch/go-chat/internal/services/frontend/internal/controllers/http"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type App struct {
	cfg        *config.Config
	logger     *zap.Logger
	grShutdown *graceful.Shutdown

	router     *mux.Router
	controller *httpadmin.Controller
}

func init() {
	gob.Register(&entities.User{})
}

func NewApp(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*App, error) {
	authClient, err := auth.NewClient(cfg.AuthService.Address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create auth client")
	}

	websiteClient, err := website.NewClient(cfg.WebsiteService.Address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create website client")
	}

	chatClient, err := chat.NewClient(cfg.ChatService.Address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chat client")
	}

	store := sessions.NewCookieStore([]byte(cfg.Session.Secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(cfg.Session.MaxAge.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	controller := httpadmin.New(
		logger,
		cfg,
		authClient,
		chatClient,
		websiteClient,
		store,
	)

	router := mux.NewRouter()
	controller.RegisterRoutes(router)

	grShutdown := graceful.NewShutdown(logger)

	return &App{
		cfg:        cfg,
		logger:     logger,
		grShutdown: grShutdown,
		router:     router,
		controller: controller,
	}, nil
}

func (a *App) Start(ctx context.Context) {
	server := &http.Server{
		Addr:         a.cfg.Handlers.HTTP.FullAddress(),
		Handler:      a.router,
		ReadTimeout:  a.cfg.Handlers.HTTP.ReadTimeout,
		WriteTimeout: a.cfg.Handlers.HTTP.WriteTimeout,
	}

	go func() {
		a.logger.Info("Starting HTTP server",
			zap.String("address", a.cfg.Handlers.HTTP.FullAddress()))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	if err := a.grShutdown.Wait(a.cfg.GracefulShutdown); err != nil {
		a.logger.Error("Error during graceful shutdown", zap.Error(err))
	} else {
		a.logger.Info("Application gracefully stopped")
	}
}

func (a *App) Stop(ctx context.Context) error {
	return nil
}
