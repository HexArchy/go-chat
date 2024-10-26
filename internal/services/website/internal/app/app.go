package app

import (
	"context"

	"github.com/HexArch/go-chat/internal/pkg/graceful-shutdown"
	"github.com/HexArch/go-chat/internal/services/website/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/website/internal/config"
	"github.com/HexArch/go-chat/internal/services/website/internal/controllers"
	"github.com/HexArch/go-chat/internal/services/website/internal/services/rooms"
	roomstorage "github.com/HexArch/go-chat/internal/services/website/internal/services/rooms/storage"
	createroom "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/create-room"
	deleteroom "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/delete-room"
	getallrooms "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/get-all-rooms"
	getownerrooms "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/get-owner-rooms"
	getroom "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/get-room"
	searchrooms "github.com/HexArch/go-chat/internal/services/website/internal/use-cases/search-rooms"
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

	// Initialize storage and services.
	roomStorage := roomstorage.New(db)
	roomService := rooms.NewService(rooms.Deps{RoomStorage: roomStorage})

	// Initialize Auth Client.
	authClient, err := auth.NewAuthClient(logger, cfg.AuthService.Address, cfg.AuthService.JWTSecret)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize auth client")
	}

	createRoom := createroom.New(createroom.Deps{RoomService: roomService})
	deleteRoom := deleteroom.New(deleteroom.Deps{RoomService: roomService})
	getOwnerRooms := getownerrooms.New(getownerrooms.Deps{RoomService: roomService})
	getRoom := getroom.New(getroom.Deps{RoomService: roomService})
	searchRooms := searchrooms.New(searchrooms.Deps{RoomService: roomService})
	getAllRooms := getallrooms.New(getallrooms.Deps{RoomService: roomService})

	roomServiceServer := controllers.NewWebsiteServiceServer(
		createRoom,
		deleteRoom,
		getRoom,
		getOwnerRooms,
		searchRooms,
		getAllRooms,
	)

	grShutdown := graceful.NewShutdown(logger)

	server := controllers.NewServer(logger, cfg, authClient, roomServiceServer)

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
