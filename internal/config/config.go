// Package config handles loading settings for the application.
//
// The application needs to know things like "where is NocoDB?" and "what's the
// API key?" - those settings live in environment variables or a .env file.
//
// This package loads those settings when the application starts and makes them
// available to the rest of the code through AppConfig.
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all the settings the application needs.
//
// At the moment it's just NocoDB settings, but if we needed other configuration
// in the future, it would go here too.
type Config struct {
	// Where NocoDB is and how to talk to it
	NocoDBEndpoint string // The NocoDB server URL
	NocoDBAPIKey   string // The API key for authentication
	NocoDBTableID  string // Which table to read stories from
}

// AppConfig is the global config that everyone uses.
//
// It gets set when LoadConfig() runs at startup, and then the rest of the
// application just reads from AppConfig.NocoDBEndpoint, AppConfig.NocoDBAPIKey, etc.
var AppConfig *Config

// LoadConfig reads settings from environment variables and the .env file.
//
// It looks for a .env file first (for local development), then checks environment
// variables. If any required settings are missing, it stops the application -
// better to fail early than to run with wrong settings!
func LoadConfig() {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	AppConfig = &Config{
		NocoDBEndpoint: getEnvWithDefault("NOCODB_ENDPOINT", ""),
		NocoDBAPIKey:   getEnvWithDefault("NOCODB_API_KEY", ""),
		NocoDBTableID:  getEnvWithDefault("NOCODB_TABLE_ID", ""),
	}

	log.Println("Configuration loaded for NocoDB")

	// Validate NocoDB configuration
	if AppConfig.NocoDBEndpoint == "" {
		log.Fatal("NOCODB_ENDPOINT is required")
	}
	if AppConfig.NocoDBAPIKey == "" {
		log.Fatal("NOCODB_API_KEY is required")
	}
	if AppConfig.NocoDBTableID == "" {
		log.Fatal("NOCODB_TABLE_ID is required")
	}
	log.Println("NocoDB configuration validated successfully")
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
