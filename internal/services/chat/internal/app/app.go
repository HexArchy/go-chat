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
	getmessagesuc "github.com/HexArch/go-chat/internal/services/chat/internal/use-cases/get-messages"
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

	authClient, err := auth.NewClient(cfg.AuthService.Address, cfg.AuthService.ServiceToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create auth client")
	}

	websiteClient, err := website.NewClient(cfg.WebsiteService.Address, cfg.WebsiteService.ServiceToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create website client")
	}

	messageStorage := chatstorage.New(db)
	chatService := chat.New(chat.Deps{
		MessageStorage: messageStorage,
		WebsiteClient:  websiteClient,
	}, logger)

	connectUC := connectuc.New(connectuc.Deps{
		WebsiteService: websiteClient,
		ChatService:    chatService,
		AuthService:    authClient,
	})

	disconnectUC := disconnectuc.New(disconnectuc.Deps{
		ChatService: chatService,
		AuthService: authClient,
	})

	sendMessageUC := sendmessageuc.New(sendmessageuc.Deps{
		ChatService: chatService,
		AuthService: authClient,
	})

	getMessagesUC := getmessagesuc.New(getmessagesuc.Deps{
		ChatService:    chatService,
		WebsiteService: websiteClient,
		AuthService:    authClient,
	})

	chatServiceServer := controllers.NewChatServiceServer(
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
		chatServiceServer,
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
