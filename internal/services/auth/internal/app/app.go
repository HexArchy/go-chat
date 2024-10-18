package app

import (
	"context"
	"runtime/debug"

	"github.com/HexArch/go-chat/internal/pkg/logger"
	"github.com/HexArch/go-chat/internal/services/auth/internal/config"
	"github.com/HexArch/go-chat/internal/services/auth/internal/handlers/http/api"
	"github.com/HexArch/go-chat/internal/services/auth/internal/services/auth"
	"github.com/HexArch/go-chat/internal/services/auth/internal/services/rbac"

	authStorage "github.com/HexArch/go-chat/internal/services/auth/internal/services/auth/storage"
	usecases "github.com/HexArch/go-chat/internal/services/auth/internal/use_cases"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	// Utils.
	cfg    *config.Config
	logger *zap.Logger

	// Storages.
	authStorage *authStorage.Storage

	// Services.
	authService *auth.AuthService
	rbac        *rbac.RBAC

	// Handlers.
	apiHandler *api.Handler

	// UseCases.
	useCases *usecases.UseCases
}

func Start(ctx context.Context) (err error) {
	defer func() {
		msg := recover()
		if msg != nil {
			err = errors.Errorf("panic: %v, %s", msg, string(debug.Stack()))
		}

		logger, logErr := logger.NewLogger("error")
		if logErr == nil {
			logger.Error("Panic occurred", zap.Error(err))
		}
	}()

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	logger, err := logger.NewLogger(cfg.Logging.Level)
	if err != nil {
		return err
	}

	db, err := gorm.Open(postgres.Open(cfg.Engines.Storage.URL), &gorm.Config{})
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return errors.Wrap(err, "failed to get database connection")
	}

	sqlDB.SetMaxOpenConns(cfg.Engines.Storage.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Engines.Storage.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Engines.Storage.ConnMaxLifetime)

	rbacService := rbac.NewRBAC()

	authStorage := authStorage.NewStorage(db)
	authService := auth.NewAuthService(
		authStorage,
		rbacService,
		cfg.Auth.JWT.AccessSecret,
		cfg.Auth.JWT.RefreshSecret,
		cfg.Auth.JWT.AccessExpiryHours,
		cfg.Auth.JWT.RefreshExpiryHours,
	)

	useCases := usecases.NewUseCases(authService, authStorage, rbacService)

	apiHandler, err := api.NewHandler()
	if err != nil {
		return errors.Wrap(err, "failed to start api server")
	}

	app := &App{
		cfg:         cfg,
		logger:      logger,
		authStorage: authStorage,
		authService: authService,
		rbac:        rbacService,
		apiHandler:  apiHandler,
		useCases:    useCases,
	}

	app.start(ctx)

	return nil
}
