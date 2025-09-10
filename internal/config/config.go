// Package config provides configuration management for the application
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration values for the application
type Config struct {
	// NocoDB configuration
	NocoDBEndpoint string
	NocoDBAPIKey   string
	NocoDBTableID  string

	// Data source selection
	UseNocoDB bool
}

// Global configuration instance
var AppConfig *Config

// LoadConfig loads configuration from environment variables and .env file
func LoadConfig() {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	AppConfig = &Config{
		NocoDBEndpoint: getEnvWithDefault("NOCODB_ENDPOINT", ""),
		NocoDBAPIKey:   getEnvWithDefault("NOCODB_API_KEY", ""),
		NocoDBTableID:  getEnvWithDefault("NOCODB_TABLE_ID", ""),
		UseNocoDB:      getEnvBool("USE_NOCODB", false),
	}

	log.Printf("Configuration loaded - UseNocoDB: %t", AppConfig.UseNocoDB)

	// Validate NocoDB configuration if enabled
	if AppConfig.UseNocoDB {
		if AppConfig.NocoDBEndpoint == "" {
			log.Fatal("NOCODB_ENDPOINT is required when USE_NOCODB=true")
		}
		if AppConfig.NocoDBAPIKey == "" {
			log.Fatal("NOCODB_API_KEY is required when USE_NOCODB=true")
		}
		if AppConfig.NocoDBTableID == "" {
			log.Fatal("NOCODB_TABLE_ID is required when USE_NOCODB=true")
		}
		log.Println("NocoDB configuration validated successfully")
	}
}

// getEnvWithDefault returns the environment variable value or a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns the environment variable as a boolean or a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
