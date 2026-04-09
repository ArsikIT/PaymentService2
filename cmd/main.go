package main

import (
	"log"

	"payment-service/internal/app"
)

func main() {
	cfg := app.LoadConfig()

	server, err := app.NewServer(cfg)
	if err != nil {
		log.Fatalf("initialize server: %v", err)
	}
	defer server.Close()

	if err := server.Run(); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
