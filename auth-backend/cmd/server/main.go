package main

import (
	"fmt"
	"log"

	"authbackend/internal/config"
	"authbackend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	if err := srv.Initialize(); err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	log.Printf("Starting server on %s:%d", cfg.Server.Host, cfg.Server.Port)

	if err := srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}

	fmt.Println("Server stopped")
}