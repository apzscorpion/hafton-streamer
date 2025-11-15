package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Telegram struct {
		BotToken string `yaml:"bot_token"`
	} `yaml:"telegram"`
	Server struct {
		Port        int    `yaml:"port"`
		Domain      string `yaml:"domain"`
		StoragePath string `yaml:"storage_path"`
	} `yaml:"server"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
	Retention struct {
		Days int `yaml:"days"`
	} `yaml:"retention"`
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate required fields
	if config.Telegram.BotToken == "" {
		return nil, fmt.Errorf("telegram.bot_token is required")
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	if config.Server.StoragePath == "" {
		config.Server.StoragePath = "./storage"
	}
	if config.Database.Path == "" {
		config.Database.Path = "./data/bot.db"
	}
	if config.Retention.Days == 0 {
		config.Retention.Days = 5
	}

	return &config, nil
}

