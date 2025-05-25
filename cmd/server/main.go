package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aelhady03/twerlo-chat-app/internal/api"
	"github.com/aelhady03/twerlo-chat-app/internal/auth"
	"github.com/aelhady03/twerlo-chat-app/internal/config"
	"github.com/aelhady03/twerlo-chat-app/internal/database"
	"github.com/aelhady03/twerlo-chat-app/internal/repository"
	"github.com/aelhady03/twerlo-chat-app/internal/service"
	"github.com/aelhady03/twerlo-chat-app/internal/websocket"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, "twerlo-chat-app")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo, jwtManager)
	messageService := service.NewMessageService(messageRepo, userRepo)

	// Initialize WebSocket hub
	hub := websocket.NewHub(jwtManager, userService)
	go hub.Run()

	// Initialize router
	router := api.NewRouter(userService, messageService, jwtManager, hub, cfg)
	routes := router.SetupRoutes()

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	log.Printf("WebSocket endpoint: ws://%s/ws", addr)
	log.Printf("API documentation available at: http://%s/api", addr)

	if err := http.ListenAndServe(addr, routes); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
