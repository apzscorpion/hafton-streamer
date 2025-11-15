package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"hafton-movie-bot/internal/bot"
	"hafton-movie-bot/internal/config"
	"hafton-movie-bot/internal/database"
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
			log.Println("Warning: Domain not set. Streaming links will not work properly.")
			domain = "localhost:8080"
		}
	}

	// Create and start bot
	telegramBot, err := bot.New(cfg, db, storage, domain)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := telegramBot.Start(); err != nil {
			log.Fatalf("Bot error: %v", err)
		}
	}()

	log.Println("Bot is running. Press Ctrl+C to stop.")
	<-sigChan
	log.Println("Shutting down...")
}

