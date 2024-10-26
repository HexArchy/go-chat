package app

import (
	"context"
	"encoding/gob"
	"net/http"
	"time"

	"github.com/HexArch/go-chat/internal/pkg/graceful-shutdown"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/chat"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/shared"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/clients/website"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/config"
	httpadmin "github.com/HexArch/go-chat/internal/services/frontend/internal/controllers/http"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/entities"
	tokenmanager "github.com/HexArch/go-chat/internal/services/frontend/internal/services/token-manager"
	authuc "github.com/HexArch/go-chat/internal/services/frontend/internal/use-cases/auth"
	profileuc "github.com/HexArch/go-chat/internal/services/frontend/internal/use-cases/profile"
	roomsuc "github.com/HexArch/go-chat/internal/services/frontend/internal/use-cases/rooms"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	sessionName = "chat_session"
	tokenKey    = "token"
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
	gob.Register(&tokenmanager.SessionData{})
}

func NewApp(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*App, error) {
	store := sessions.NewCookieStore([]byte(cfg.Session.Secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(cfg.Session.MaxAge.Seconds()),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Domain:   "localhost",
	}

	authClient, err := auth.NewClient(logger, cfg.AuthService.Address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create auth client")
	}

	tokenManager := tokenmanager.NewTokenManager(authClient, logger, store, sessionName)

	authInterceptor := shared.NewAuthInterceptor(logger, tokenManager, store, sessionName)

	websiteClient, err := website.NewClient(logger, cfg.WebsiteService.Address, authInterceptor, &shared.RetryConfig{
		MaxAttempts:     2,
		InitialInterval: 20 * time.Millisecond,
		MaxInterval:     1 * time.Second,
		Multiplier:      0.24,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create website client")
	}

	chatClient, err := chat.NewClient(logger, cfg.ChatService.Address, authInterceptor)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chat client")
	}

	controller := httpadmin.NewController(
		logger, cfg,
		authuc.NewLoginUseCase(authClient, logger),
		authuc.NewRegisterUseCase(authClient, logger),
		authuc.NewLogoutUseCase(authClient, logger),
		profileuc.NewGetProfileUseCase(authClient, logger),
		profileuc.NewEditProfileUseCase(authClient, logger),
		roomsuc.NewCreateRoomUseCase(websiteClient, logger),
		roomsuc.NewDeleteRoomUseCase(websiteClient, logger),
		roomsuc.NewListRoomsUseCase(websiteClient, logger),
		roomsuc.NewSearchRoomsUseCase(websiteClient, logger),
		roomsuc.NewViewRoomUseCase(websiteClient, chatClient, logger),
		roomsuc.NewManageWebSocketUseCase(chatClient, logger),
		tokenManager, store, sessionName, tokenKey,
	)

	router := controller.SetupRoutes()

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
