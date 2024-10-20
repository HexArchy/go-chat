package main

import (
	"flag"
	"log"

	"github.com/HexArch/go-chat/internal/pkg/logger"
	"github.com/HexArch/go-chat/internal/services/auth/internal/config"
	authMigrations "github.com/HexArch/go-chat/internal/services/auth/internal/services/auth/storage/migrations"
	userMigrations "github.com/HexArch/go-chat/internal/services/auth/internal/services/user/storage/migrations"
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

	if err := userMigrations.Migrate(db); err != nil {
		logger.Fatal("Failed to migrate user tables", zap.Error(err))
	}
	logger.Info("User tables migrated successfully")

	if err := authMigrations.Migrate(db); err != nil {
		logger.Fatal("Failed to migrate auth tables", zap.Error(err))
	}
	logger.Info("Auth tables migrated successfully")
}
