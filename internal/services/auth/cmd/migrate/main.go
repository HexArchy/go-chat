package main

import (
	"log"

	"github.com/HexArch/go-chat/internal/services/auth/internal/config"
	"github.com/HexArch/go-chat/internal/services/auth/internal/services/auth/storage/migrations"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.Engines.Storage.URL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	schemaName := "go_chat_schema"

	if err := migrations.SetupSchema(db, schemaName); err != nil {
		log.Fatalf("Failed to setup schema: %v", err)
	}

	if err := migrations.CreateInitialSchema(db); err != nil {
		log.Fatalf("Failed to create initial schema: %v", err)
	}

	log.Println("Migrations completed successfully")
}
