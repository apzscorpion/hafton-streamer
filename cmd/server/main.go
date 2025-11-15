package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hafton-movie-bot/internal/cleanup"
	"hafton-movie-bot/internal/config"
	"hafton-movie-bot/internal/database"
	"hafton-movie-bot/internal/server"
	"hafton-movie-bot/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Ensure database directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize storage
	storage, err := storage.New(cfg.Server.StoragePath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Get domain from config or environment
	domain := cfg.Server.Domain
	if domain == "" {
		domain = os.Getenv("DOMAIN")
		if domain == "" {
			log.Println("Warning: Domain not set. Using localhost:8080")
			domain = "localhost:8080"
		}
	}

	// Create and start server
	httpServer := server.New(cfg, db, storage, domain)

	// Start cleanup goroutine (runs every hour)
	cleanup := cleanup.New(db, storage, time.Hour)
	go cleanup.Start()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := httpServer.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	log.Println("HTTP server is running. Press Ctrl+C to stop.")
	<-sigChan
	log.Println("Shutting down...")
}

