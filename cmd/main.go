package main

import (
	"log"

	"payment-service/internal/app"
)

func main() {
	cfg := app.LoadConfig()

	server, cleanup, err := app.NewServer(cfg)
	if err != nil {
		log.Fatalf("initialize server: %v", err)
	}
	defer cleanup()

	if err := server.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
