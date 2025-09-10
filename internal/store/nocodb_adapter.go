// Package store provides NocoDB implementation of the data adapter
package store

import (
	"fmt"
	"log"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/nocodb"
)

// NocoDBAdapter implements DataAdapter for NocoDB API
type NocoDBAdapter struct {
	client *nocodb.Client
}

// NewNocoDBAdapter creates a new NocoDB adapter
func NewNocoDBAdapter() (*NocoDBAdapter, error) {
	client, err := nocodb.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create NocoDB client: %w", err)
	}

	return &NocoDBAdapter{
		client: client,
	}, nil
}

// GetAllStories retrieves all stories from NocoDB
func (n *NocoDBAdapter) GetAllStories() ([]data.Story, error) {
	records, err := n.client.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to get all records from NocoDB: %w", err)
	}

	return n.convertRecordsToStories(records)
}

// GetStoriesForTheme retrieves stories filtered by theme from NocoDB
func (n *NocoDBAdapter) GetStoriesForTheme(themeTitle string) ([]data.Story, error) {
	log.Println("Getting stories for theme", themeTitle)

	records, err := n.client.GetFilteredRecords("Themes", themeTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered records for theme: %w", err)
	}

	return n.convertRecordsToStories(records)
}

// GetStoriesForType retrieves stories filtered by type from NocoDB
func (n *NocoDBAdapter) GetStoriesForType(typeTitle string) ([]data.Story, error) {
	log.Println("Getting stories for type", typeTitle)

	records, err := n.client.GetFilteredRecords("Type", typeTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered records for type: %w", err)
	}

	return n.convertRecordsToStories(records)
}

// GetStoriesForWeather retrieves stories filtered by weather from NocoDB
func (n *NocoDBAdapter) GetStoriesForWeather(weatherTitle string) ([]data.Story, error) {
	log.Println("Getting stories for weather", weatherTitle)

	records, err := n.client.GetFilteredRecords("Weather", weatherTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered records for weather: %w", err)
	}

	return n.convertRecordsToStories(records)
}

// GetThemes retrieves all unique themes from NocoDB
func (n *NocoDBAdapter) GetThemes() ([]data.Theme, error) {
	records, err := n.client.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to get all records: %w", err)
	}

	themeMap := make(map[string]data.Theme)

	for _, record := range records {
		if themesField, exists := record["Themes"]; exists && themesField != nil {
			themes, err := nocodb.ParseThemesFromNocoDB(themesField)
			if err != nil {
				log.Printf("Warning: failed to parse themes from record: %v", err)
				continue
			}

			for _, theme := range themes {
				if theme.Title != "" {
					themeMap[theme.Title] = theme
				}
			}
		}
	}

	// Convert map to slice
	var themes []data.Theme
	for _, theme := range themeMap {
		themes = append(themes, theme)
	}

	return themes, nil
}

// GetTypes retrieves all unique types from NocoDB
func (n *NocoDBAdapter) GetTypes() ([]data.Type, error) {
	records, err := n.client.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to get all records: %w", err)
	}

	typeMap := make(map[string]data.Type)

	for _, record := range records {
		if typesField, exists := record["Type"]; exists && typesField != nil {
			types, err := nocodb.ParseTypesFromNocoDB(typesField)
			if err != nil {
				log.Printf("Warning: failed to parse types from record: %v", err)
				continue
			}

			for _, typeObj := range types {
				if typeObj.Title != "" {
					typeMap[typeObj.Title] = typeObj
				}
			}
		}
	}

	// Convert map to slice
	var types []data.Type
	for _, typeObj := range typeMap {
		types = append(types, typeObj)
	}

	return types, nil
}

// GetWeather retrieves all unique weather conditions from NocoDB
func (n *NocoDBAdapter) GetWeather() ([]data.Weather, error) {
	records, err := n.client.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to get all records: %w", err)
	}

	weatherMap := make(map[string]data.Weather)

	for _, record := range records {
		if weatherField, exists := record["Weather"]; exists && weatherField != nil {
			weather, err := nocodb.ParseWeatherFromNocoDB(weatherField)
			if err != nil {
				log.Printf("Warning: failed to parse weather from record: %v", err)
				continue
			}

			for _, weatherObj := range weather {
				if weatherObj.Title != "" {
					weatherMap[weatherObj.Title] = weatherObj
				}
			}
		}
	}

	// Convert map to slice
	var weather []data.Weather
	for _, weatherObj := range weatherMap {
		weather = append(weather, weatherObj)
	}

	return weather, nil
}

// convertRecordsToStories converts NocoDB records to Story structs
func (n *NocoDBAdapter) convertRecordsToStories(records []map[string]interface{}) ([]data.Story, error) {
	var stories []data.Story

	for _, record := range records {
		story, err := nocodb.NocoDBRecordToStory(record)
		if err != nil {
			log.Printf("Warning: failed to convert record to story: %v", err)
			continue
		}

		// Ensure URL is set
		if story.URL == "" {
			story.URL = CreateStoryURLFromFinding(story.Finding)
		}

		stories = append(stories, story)
	}

	return stories, nil
}

// Helper functions that need to be exposed from nocodb package
// We'll need to create these as exported functions in the nocodb package
