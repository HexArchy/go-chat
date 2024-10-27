package app

import (
	"context"

	"github.com/HexArch/go-chat/internal/pkg/graceful-shutdown"
	"github.com/HexArch/go-chat/internal/services/chat/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/chat/internal/clients/website"
	"github.com/HexArch/go-chat/internal/services/chat/internal/config"
	"github.com/HexArch/go-chat/internal/services/chat/internal/controllers"
	"github.com/HexArch/go-chat/internal/services/chat/internal/services/chat"
	chatstorage "github.com/HexArch/go-chat/internal/services/chat/internal/services/chat/storage"
	connectuc "github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/connect"
	disconnectuc "github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/disconnect"
	getmessages "github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/get-messages"
	sendmessageuc "github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/send-message"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	cfg        *config.Config
	logger     *zap.Logger
	grShutdown *graceful.Shutdown
	server     *controllers.Server
}

func NewApp(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*App, error) {
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

	authClient, err := auth.NewClient(logger, cfg.AuthService.Address, cfg.AuthService.ServiceToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create auth client")
	}

	websiteClient, err := website.NewClient(logger, cfg.WebsiteService.Address, cfg.WebsiteService.ServiceToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create website client")
	}

	messageStorage := chatstorage.NewStorage(db)
	chatService := chat.NewService(chat.Deps{
		Storage: messageStorage,
	}, logger)

	connectUC := connectuc.New(connectuc.Deps{
		WebsiteService: websiteClient,
		ChatService:    chatService,
	})

	disconnectUC := disconnectuc.New(disconnectuc.Deps{
		ChatService: chatService,
	})

	sendMessageUC := sendmessageuc.New(sendmessageuc.Deps{
		ChatService: chatService,
	})

	getMessagesUC := getmessages.New(getmessages.Deps{
		ChatService: chatService,
	})

	wsHandler := controllers.NewWebSocketHandler(
		logger,
		connectUC,
		disconnectUC,
		sendMessageUC,
		getMessagesUC,
		authClient,
	)

	grShutdown := graceful.NewShutdown(logger)

	server := controllers.NewServer(
		logger,
		cfg,
		wsHandler,
	)

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
	return a.server.Stop(ctx)
}
