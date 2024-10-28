package app

import (
	"context"
	"time"

	graceful "github.com/HexArch/go-chat/internal/pkg/graceful-shutdown"
	"github.com/HexArch/go-chat/internal/services/website/internal/clients/auth"
	"github.com/HexArch/go-chat/internal/services/website/internal/config"
	"github.com/HexArch/go-chat/internal/services/website/internal/controllers"
	"github.com/HexArch/go-chat/internal/services/website/internal/controllers/cache"
	"github.com/HexArch/go-chat/internal/services/website/internal/metrics"
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
	metrics    *metrics.WebsiteMetrics
	server     *controllers.Server
	db         *gorm.DB
}

func NewApp(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*App, error) {
	// Initialize metrics.
	metrics := metrics.NewWebsiteMetrics("website_service")

	// Initialize database.
	db, err := gorm.Open(postgres.Open(cfg.Engines.Storage.URL), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get database connection")
	}

	// Configure database connection pool
	sqlDB.SetMaxOpenConns(cfg.Engines.Storage.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Engines.Storage.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Engines.Storage.ConnMaxLifetime)

	// Initialize storage and services
	roomStorage := roomstorage.New(db)
	roomService := rooms.NewService(rooms.Deps{
		RoomStorage: roomStorage,
	})

	// Initialize Auth Client
	authClient, err := auth.NewAuthClient(
		logger.Named("auth-client"),
		cfg.AuthService.Address,
		cfg.AuthService.JWTSecret,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize auth client")
	}

	// Initialize cache
	roomCache := cache.NewRoomCache(5 * time.Minute)

	// Initialize use cases
	createRoom := createroom.New(createroom.Deps{
		RoomService: roomService,
	})
	deleteRoom := deleteroom.New(deleteroom.Deps{
		RoomService: roomService,
	})
	getOwnerRooms := getownerrooms.New(getownerrooms.Deps{
		RoomService: roomService,
	})
	getRoom := getroom.New(getroom.Deps{
		RoomService: roomService,
	})
	searchRooms := searchrooms.New(searchrooms.Deps{
		RoomService: roomService,
	})
	getAllRooms := getallrooms.New(getallrooms.Deps{
		RoomService: roomService,
	})

	// Initialize website service
	roomServiceServer := controllers.NewWebsiteServiceServer(
		logger.Named("website-service"),
		metrics,
		roomCache,
		createRoom,
		deleteRoom,
		getRoom,
		getOwnerRooms,
		searchRooms,
		getAllRooms,
	)

	// Initialize graceful shutdown
	grShutdown := graceful.NewShutdown(logger)

	// Initialize server
	server := controllers.NewServer(
		logger.Named("server"),
		cfg,
		authClient,
		roomServiceServer,
	)

	return &App{
		cfg:        cfg,
		logger:     logger,
		grShutdown: grShutdown,
		metrics:    metrics,
		server:     server,
		db:         db,
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
