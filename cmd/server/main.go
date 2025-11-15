package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"hafton-movie-bot/internal/cleanup"
	"hafton-movie-bot/internal/config"
	"hafton-movie-bot/internal/database"
	"hafton-movie-bot/internal/server"
	"hafton-movie-bot/internal/storage"
)

func main() {
	// Try to load config file, but don't fail if it doesn't exist (Railway uses env vars)
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}
	
	cfg, err := config.Load(configPath)
	if err != nil {
		// If config file doesn't exist, create default config from environment
		log.Printf("Config file not found, using environment variables: %v", err)
		cfg = &config.Config{}
		cfg.Telegram.BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
		if cfg.Telegram.BotToken == "" {
			log.Fatalf("TELEGRAM_BOT_TOKEN environment variable is required")
		}
		cfg.Server.Port = 8080
		if port := os.Getenv("PORT"); port != "" {
			if p, err := strconv.Atoi(port); err == nil {
				cfg.Server.Port = p
			}
		}
		cfg.Server.Domain = os.Getenv("DOMAIN")
		if cfg.Server.Domain == "" {
			cfg.Server.Domain = os.Getenv("RAILWAY_PUBLIC_DOMAIN")
		}
		cfg.Server.StoragePath = os.Getenv("STORAGE_PATH")
		if cfg.Server.StoragePath == "" {
			cfg.Server.StoragePath = "./storage"
		}
		cfg.Database.Path = os.Getenv("DATABASE_PATH")
		if cfg.Database.Path == "" {
			cfg.Database.Path = "./data/bot.db"
		}
		cfg.Retention.Days = 5
		if days := os.Getenv("RETENTION_DAYS"); days != "" {
			if d, err := strconv.Atoi(days); err == nil {
				cfg.Retention.Days = d
			}
		}
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

