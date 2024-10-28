package app

import (
	"context"
	"time"

	graceful "github.com/HexArch/go-chat/internal/pkg/graceful-shutdown"
	"github.com/HexArch/go-chat/internal/services/auth/internal/config"
	"github.com/HexArch/go-chat/internal/services/auth/internal/controllers"
	"github.com/HexArch/go-chat/internal/services/auth/internal/controllers/cache"
	"github.com/HexArch/go-chat/internal/services/auth/internal/controllers/middleware"
	"github.com/HexArch/go-chat/internal/services/auth/internal/metrics"
	"github.com/HexArch/go-chat/internal/services/auth/internal/services/auth"
	tokenstorage "github.com/HexArch/go-chat/internal/services/auth/internal/services/auth/storage"
	"github.com/HexArch/go-chat/internal/services/auth/internal/services/user"
	userstorage "github.com/HexArch/go-chat/internal/services/auth/internal/services/user/storage"

	createuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/create-user"
	deleteuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/delete-user"
	getuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/get-user"
	getusers "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/get-users"
	"github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/login"
	"github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/logout"
	refreshtoken "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/refresh-token"
	updateuser "github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/update-user"
	"github.com/HexArch/go-chat/internal/services/auth/internal/use-cases/validatetoken"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	cfg        *config.Config
	logger     *zap.Logger
	grShutdown *graceful.Shutdown

	server *controllers.Server
}

func NewApp(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*App, error) {
	// Initialize database.
	db, err := gorm.Open(postgres.Open(cfg.Engines.Storage.URL), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database connection")
	}

	sqlDB.SetMaxOpenConns(cfg.Engines.Storage.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Engines.Storage.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Engines.Storage.ConnMaxLifetime)

	// Initialize storages.
	userStorage := userstorage.New(db)
	tokenStorage := tokenstorage.New(db)

	// Initialize services.
	userService := user.NewService(user.Deps{UserStorage: userStorage})
	authService := auth.NewService(
		auth.Deps{
			UserStorage:  userStorage,
			TokenStorage: tokenStorage,
			Secrets: auth.TokenSecrets{
				AccessTokenSecret:  cfg.Auth.JWT.AccessSecret,
				RefreshTokenSecret: cfg.Auth.JWT.RefreshSecret,
			},
			TokenTTL: auth.TokenTTL{
				AccessTokenTTL:  cfg.Auth.JWT.AccessExpiryHours * time.Hour,
				RefreshTokenTTL: cfg.Auth.JWT.RefreshExpiryHours * time.Hour,
			},
		},
	)

	// Initialize use cases.
	createUserUC := createuser.New(createuser.Deps{UserService: userService})
	loginUC := login.New(login.Deps{AuthService: authService})
	refreshTokenUC := refreshtoken.New(refreshtoken.Deps{AuthService: authService})
	validateTokenUC := validatetoken.New(validatetoken.Deps{AuthService: authService})
	logoutUC := logout.New(logout.Deps{AuthService: authService})
	getUserUC := getuser.New(getuser.Deps{UserService: userService})
	getUsersUC := getusers.New(getusers.Deps{UserService: userService})
	updateUserUC := updateuser.New(updateuser.Deps{UserService: userService})
	deleteUserUC := deleteuser.New(deleteuser.Deps{UserService: userService})

	// Initialize metrics.
	metrics := metrics.NewAuthMetrics("auth_service")

	// Initialize cache.
	tokenCache := cache.NewTokenCache(5 * time.Minute)

	// Initialize controllers.
	authCtrl := controllers.NewAuthController(
		logger,
		metrics,
		createUserUC,
		loginUC,
		refreshTokenUC,
		validateTokenUC,
		logoutUC,
	)

	usersCtrl := controllers.NewUsersController(
		logger,
		metrics,
		getUserUC,
		getUsersUC,
		updateUserUC,
		deleteUserUC,
	)

	// Initialize middleware.
	authMiddleware := middleware.NewAuthMiddleware(
		logger,
		metrics,
		tokenCache,
		validateTokenUC,
		cfg.Auth.ServiceToken,
	)

	// Initialize server.
	server := controllers.NewServer(
		cfg,
		logger,
		authCtrl,
		usersCtrl,
		tokenCache,
		metrics,
		authMiddleware,
	)

	grShutdown := graceful.NewShutdown(logger)

	return &App{
		cfg:        cfg,
		logger:     logger,
		grShutdown: grShutdown,
		server:     server,
	}, nil
}

func (a *App) Start(ctx context.Context) {
	go func() {
		if err := a.server.Start(ctx); err != nil {
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
	a.logger.Info("Stopping application")
	if err := a.server.Stop(ctx); err != nil {
		return errors.Wrap(err, "failed to stop server")
	}
	return nil
}
