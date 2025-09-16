// Package store provides data access abstractions
package store

import (
	"fmt"
	"log"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/config"
)

// DataAdapter defines the interface for data access operations
type DataAdapter interface {
	// GetAllStories retrieves all stories from the data source
	GetAllStories() ([]data.Story, error)

	// GetStoriesWithConnections retrieves stories that have InspiredBy or HasInspired relationships
	GetStoriesWithConnections(limit int) ([]data.Story, error)

	// GetStoriesForTheme retrieves stories filtered by theme
	GetStoriesForTheme(themeTitle string) ([]data.Story, error)

	// GetStoriesForType retrieves stories filtered by type
	GetStoriesForType(typeTitle string) ([]data.Story, error)

	// GetStoriesForWeather retrieves stories filtered by weather
	GetStoriesForWeather(weatherTitle string) ([]data.Story, error)

	// GetThemes retrieves all unique themes
	GetThemes() ([]data.Theme, error)

	// GetTypes retrieves all unique types
	GetTypes() ([]data.Type, error)

	// GetWeather retrieves all unique weather conditions
	GetWeather() ([]data.Weather, error)

	// DropCache clears any cached data to force fresh retrieval
	DropCache() error
}

// Global adapter instance
var currentAdapter DataAdapter

// InitializeAdapter sets up the data adapter based on configuration
func InitializeAdapter() error {
	if config.AppConfig.UseNocoDB {
		adapter, err := NewNocoDBAdapter()
		if err != nil {
			return fmt.Errorf("failed to initialize NocoDB adapter: %w", err)
		}
		currentAdapter = adapter
		log.Println("Initialized NocoDB adapter")
	} else {
		currentAdapter = &SQLiteAdapter{}
		log.Println("Initialized SQLite adapter")
	}
	return nil
}

// GetAdapter returns the current data adapter
func GetAdapter() DataAdapter {
	if currentAdapter == nil {
		// Fallback to SQLite if not initialized
		currentAdapter = &SQLiteAdapter{}
	}
	return currentAdapter
}

// WarmCache pre-loads all stories to warm up any caching mechanisms
// This ensures subsequent operations are fast by triggering cache population
func WarmCache() {
	log.Println("Warming cache by fetching all stories...")
	allStories := GetAllStories()
	log.Printf("Cache warmed with %d stories", len(allStories))
}
