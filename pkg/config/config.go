package config

import (
	"errors"
	"os"
)

// Config holds the configuration values from environment variables
type Config struct {
	APIKey string
	APPKey string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	apiKey := os.Getenv("DD_API_KEY")
	if apiKey == "" {
		return nil, errors.New("DD_API_KEY environment variable is required")
	}
	appKey := os.Getenv("DD_APP_KEY")
	if appKey == "" {
		return nil, errors.New("DD_APP_KEY environment variable is required")
	}

	return &Config{
		APIKey: apiKey,
		APPKey: appKey,
	}, nil
}
