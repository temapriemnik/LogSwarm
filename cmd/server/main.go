package main

// @title Auth Backend API
// @version 1.0
// @description Authentication service API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://example.com/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

import (
	"fmt"
	"log"

	_ "authbackend/docs"
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
	log.Printf("Swagger UI available at http://%s:%d/docs", cfg.Server.Host, cfg.Server.Port)

	if err := srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}

	fmt.Println("Server stopped")
}