package main

import (
	"flag"
	"log"

	"github.com/HexArch/go-chat/internal/pkg/logger"
	"github.com/HexArch/go-chat/internal/services/website/internal/config"
	roomMigrations "github.com/HexArch/go-chat/internal/services/website/internal/services/rooms/storage/migrations"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	configPath := flag.String("config", "../../configs/config.prod.yaml", "Path to the configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := logger.NewLogger(cfg.Logging.Level)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Database URL", zap.String("url", cfg.Engines.Storage.URL))

	db, err := gorm.Open(postgres.Open(cfg.Engines.Storage.URL), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error; err != nil {
		logger.Fatal("Failed to create uuid-ossp extension", zap.Error(err))
	}

	if err := roomMigrations.Migrate(db); err != nil {
		logger.Fatal("Failed to migrate room tables", zap.Error(err))
	}
	logger.Info("Room tables migrated successfully")
}
