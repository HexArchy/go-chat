package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/HexArch/go-chat/internal/pkg/logger"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/app"
	"github.com/HexArch/go-chat/internal/services/frontend/internal/config"
	"go.uber.org/zap"
)

func main() {
	defer handlePanic()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configPath := flag.String("config", "configs/config.yaml", "Path to the configuration file")
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

	application, err := app.NewApp(ctx, cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize application", zap.Error(err))
	}

	go application.Start(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-c

	if err := application.Stop(ctx); err != nil {
		logger.Fatal("Failed to stop application", zap.Error(err))
	}
}

func handlePanic() {
	if r := recover(); r != nil {
		fmt.Fprintf(os.Stderr, "Application crashed with panic: %v\n", r)
		debug.PrintStack()
		log.Printf("Recovered from panic: %v\n", r)
		os.Exit(1)
	}
}
