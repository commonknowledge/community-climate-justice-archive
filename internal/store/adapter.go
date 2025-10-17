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

	// GetGiftedByTypes retrieves all unique gifted by values
	GetGiftedByTypes() ([]data.GiftedBy, error)

	// GetStoriesForGiftedBy retrieves stories filtered by gifted by
	GetStoriesForGiftedBy(giftedByTitle string) ([]data.Story, error)

	// GetScalePermanenceTypes retrieves all unique scale of permanence values
	GetScalePermanenceTypes() ([]data.ScalePermanence, error)

	// GetStoriesForScalePermanence retrieves stories filtered by scale of permanence
	GetStoriesForScalePermanence(scalePermanenceTitle string) ([]data.Story, error)

	// GetWhatWasIsIfTypes retrieves all unique what was/is/if values
	GetWhatWasIsIfTypes() ([]data.WhatWasIsIf, error)

	// GetStoriesForWhatWasIsIf retrieves stories filtered by what was/is/if
	GetStoriesForWhatWasIsIf(whatWasIsIfTitle string) ([]data.Story, error)

	// GetTimePeriodTypes retrieves all unique time period values
	GetTimePeriodTypes() ([]data.TimePeriod, error)

	// GetStoriesForTimePeriod retrieves stories filtered by time period
	GetStoriesForTimePeriod(timePeriodTitle string) ([]data.Story, error)

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
