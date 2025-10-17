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

// GetStoriesWithConnections retrieves stories that have InspiredBy or HasInspired relationships from NocoDB
// and includes all their connection targets to ensure lines can be drawn
func (n *NocoDBAdapter) GetStoriesWithConnections(limit int) ([]data.Story, error) {
	allStories, err := n.GetAllStories()
	if err != nil {
		return nil, fmt.Errorf("failed to get all stories: %w", err)
	}

	// Create a map for fast lookup
	storyMap := make(map[string]data.Story)
	for _, story := range allStories {
		storyMap[story.ID] = story
	}

	// First, find stories with connections
	var primaryConnectedStories []data.Story
	for _, story := range allStories {
		if len(story.InspiredBy) > 0 || len(story.HasInspired) > 0 {
			primaryConnectedStories = append(primaryConnectedStories, story)
			if len(primaryConnectedStories) >= limit {
				break
			}
		}
	}

	// Now add all their connection targets to ensure lines can be drawn
	storySet := make(map[string]data.Story)

	// Add primary connected stories
	for _, story := range primaryConnectedStories {
		storySet[story.ID] = story
	}

	// Add all their connection targets
	for _, story := range primaryConnectedStories {
		// Add InspiredBy targets
		for _, connection := range story.InspiredBy {
			if targetStory, exists := storyMap[connection.ID]; exists {
				storySet[connection.ID] = targetStory
			}
		}
		// Add HasInspired targets
		for _, connection := range story.HasInspired {
			if targetStory, exists := storyMap[connection.ID]; exists {
				storySet[connection.ID] = targetStory
			}
		}
	}

	// Convert back to slice
	var result []data.Story
	for _, story := range storySet {
		result = append(result, story)
	}

	return result, nil
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
		story, err := nocodb.NocoDBRecordToStoryWithClient(record, n.client)
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

// DropCache clears any cached data in the NocoDB client
func (n *NocoDBAdapter) DropCache() error {
	n.client.DropCache()
	return nil
}

// GetGiftedByTypes retrieves all unique gifted by values from NocoDB
func (n *NocoDBAdapter) GetGiftedByTypes() ([]data.GiftedBy, error) {
	allStories, err := n.GetAllStories()
	if err != nil {
		return nil, fmt.Errorf("failed to get all stories: %w", err)
	}

	giftedByMap := make(map[string]data.GiftedBy)
	for _, story := range allStories {
		for _, giftedBy := range story.GiftedBy {
			giftedByMap[giftedBy.Title] = giftedBy
		}
	}

	var giftedByTypes []data.GiftedBy
	for _, giftedBy := range giftedByMap {
		giftedByTypes = append(giftedByTypes, giftedBy)
	}

	return giftedByTypes, nil
}

// GetStoriesForGiftedBy retrieves stories filtered by gifted by from NocoDB
func (n *NocoDBAdapter) GetStoriesForGiftedBy(giftedByTitle string) ([]data.Story, error) {
	log.Println("Getting stories for gifted by", giftedByTitle)

	records, err := n.client.GetFilteredRecords("Gifted or co-created by…", giftedByTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered records for gifted by: %w", err)
	}

	return n.convertRecordsToStories(records)
}

// GetScalePermanenceTypes retrieves all unique scale of permanence values from NocoDB
func (n *NocoDBAdapter) GetScalePermanenceTypes() ([]data.ScalePermanence, error) {
	allStories, err := n.GetAllStories()
	if err != nil {
		return nil, fmt.Errorf("failed to get all stories: %w", err)
	}

	scalePermanenceMap := make(map[string]data.ScalePermanence)
	for _, story := range allStories {
		for _, scalePermanence := range story.ScalePermanence {
			scalePermanenceMap[scalePermanence.Title] = scalePermanence
		}
	}

	var scalePermanenceTypes []data.ScalePermanence
	for _, scalePermanence := range scalePermanenceMap {
		scalePermanenceTypes = append(scalePermanenceTypes, scalePermanence)
	}

	return scalePermanenceTypes, nil
}

// GetStoriesForScalePermanence retrieves stories filtered by scale of permanence from NocoDB
func (n *NocoDBAdapter) GetStoriesForScalePermanence(scalePermanenceTitle string) ([]data.Story, error) {
	log.Println("Getting stories for scale of permanence", scalePermanenceTitle)

	records, err := n.client.GetFilteredRecords("Scale of permanence", scalePermanenceTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered records for scale of permanence: %w", err)
	}

	return n.convertRecordsToStories(records)
}

// GetWhatWasIsIfTypes retrieves all unique what was/is/if values from NocoDB
func (n *NocoDBAdapter) GetWhatWasIsIfTypes() ([]data.WhatWasIsIf, error) {
	allStories, err := n.GetAllStories()
	if err != nil {
		return nil, fmt.Errorf("failed to get all stories: %w", err)
	}

	whatWasIsIfMap := make(map[string]data.WhatWasIsIf)
	for _, story := range allStories {
		for _, whatWasIsIf := range story.WhatWasIsIf {
			whatWasIsIfMap[whatWasIsIf.Title] = whatWasIsIf
		}
	}

	var whatWasIsIfTypes []data.WhatWasIsIf
	for _, whatWasIsIf := range whatWasIsIfMap {
		whatWasIsIfTypes = append(whatWasIsIfTypes, whatWasIsIf)
	}

	return whatWasIsIfTypes, nil
}

// GetStoriesForWhatWasIsIf retrieves stories filtered by what was/is/if from NocoDB
func (n *NocoDBAdapter) GetStoriesForWhatWasIsIf(whatWasIsIfTitle string) ([]data.Story, error) {
	log.Println("Getting stories for what was/is/if", whatWasIsIfTitle)

	records, err := n.client.GetFilteredRecords("What was/is/if", whatWasIsIfTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered records for what was/is/if: %w", err)
	}

	return n.convertRecordsToStories(records)
}

// GetTimePeriodTypes retrieves all unique time period values from NocoDB
func (n *NocoDBAdapter) GetTimePeriodTypes() ([]data.TimePeriod, error) {
	allStories, err := n.GetAllStories()
	if err != nil {
		return nil, fmt.Errorf("failed to get all stories: %w", err)
	}

	timePeriodMap := make(map[string]data.TimePeriod)
	for _, story := range allStories {
		for _, timePeriod := range story.TimePeriod {
			timePeriodMap[timePeriod.Title] = timePeriod
		}
	}

	var timePeriodTypes []data.TimePeriod
	for _, timePeriod := range timePeriodMap {
		timePeriodTypes = append(timePeriodTypes, timePeriod)
	}

	return timePeriodTypes, nil
}

// GetStoriesForTimePeriod retrieves stories filtered by time period from NocoDB
func (n *NocoDBAdapter) GetStoriesForTimePeriod(timePeriodTitle string) ([]data.Story, error) {
	log.Println("Getting stories for time period", timePeriodTitle)

	records, err := n.client.GetFilteredRecords("Time period", timePeriodTitle)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered records for time period: %w", err)
	}

	return n.convertRecordsToStories(records)
}
