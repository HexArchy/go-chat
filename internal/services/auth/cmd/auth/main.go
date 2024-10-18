package main

import (
	"context"
	"log"

	"github.com/HexArch/go-chat/internal/services/auth/internal/app"
)

func main() {
	err := app.Start(context.Background())
	if err != nil {
		log.Fatalf("error running app: %v", err)
	}
}
