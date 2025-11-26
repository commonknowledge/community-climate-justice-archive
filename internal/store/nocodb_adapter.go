// Package store has the NocoDB version of the adapter.
//
// This file is the "translator" that talks to NocoDB (our current database) and
// turns what it sends back into Story structs that the rest of the code can use.
//
// When something calls GetAllStories(), this adapter:
// - Makes a request to the NocoDB API
// - Gets back JSON data
// - Converts it into proper Story structs
// - Sends those back
//
// Same thing for GetStoriesForTheme(), GetStoryByID(), and all the other functions.
//
// It also handles caching (storing data temporarily so we don't have to keep asking
// NocoDB for the same thing over and over).
//
// Because this adapter implements the same interface as any other adapter would,
// we could swap it for a SQLite adapter later without changing anything else.
package store

import (
	"fmt"
	"log"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/nocodb"
)

// NocoDBAdapter implements DataAdapter for NocoDB API
//
// Performance Note:
// This adapter uses a pre-computed story index for taxonomy lookups. After fetching
// all stories once, we build an index that allows instant lookups by theme, type,
// weather, and other taxonomies. This makes builds fast and efficient, especially
// as the archive grows.
type NocoDBAdapter struct {
	client     *nocodb.Client
	storyIndex *StoryIndex // Pre-computed indexes for fast story lookups
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
//
// Performance Optimisation:
// After fetching and converting all stories, we build a story index if one doesn't
// exist yet. This index allows us to do instant lookups by taxonomy (theme, type,
// weather, etc.) rather than scanning through all stories repeatedly.
//
// The first call to this function will take slightly longer (to build the index),
// but all subsequent taxonomy lookups will be nearly instant.
func (n *NocoDBAdapter) GetAllStories() ([]data.Story, error) {
	records, err := n.client.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to get all records from NocoDB: %w", err)
	}

	stories, err := n.convertRecordsToStories(records)
	if err != nil {
		return nil, err
	}

	// Build the story index if we haven't already
	// This happens once per build and makes all subsequent taxonomy lookups fast
	if n.storyIndex == nil {
		n.storyIndex = BuildStoryIndex(stories)
	}

	return stories, nil
}

// GetStoryByID retrieves a single story by its ID from NocoDB
func (n *NocoDBAdapter) GetStoryByID(id string) (data.Story, error) {
	record, err := n.client.GetRecordByID(id)
	if err != nil {
		return data.Story{}, fmt.Errorf("failed to get record by ID %s from NocoDB: %w", id, err)
	}

	stories, err := n.convertRecordsToStories([]map[string]interface{}{record})
	if err != nil {
		return data.Story{}, fmt.Errorf("failed to convert record to story: %w", err)
	}

	if len(stories) == 0 {
		return data.Story{}, fmt.Errorf("story with ID %s not found", id)
	}

	return stories[0], nil
}

// GetRawRecords returns the raw NocoDB API response without any processing - for debugging
func (n *NocoDBAdapter) GetRawRecords() ([]map[string]interface{}, error) {
	records, err := n.client.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to get raw records from NocoDB: %w", err)
	}

	return records, nil
}

// GetStoriesForTheme retrieves stories filtered by theme
//
// Performance:
// Uses the pre-computed story index for instant O(1) lookup. The index is built
// automatically when GetAllStories() is first called.
//
// If the index hasn't been built yet, we ensure it exists by calling GetAllStories first.
func (n *NocoDBAdapter) GetStoriesForTheme(themeTitle string) ([]data.Story, error) {
	// Ensure the index exists (it should already be built by GetAllStories)
	if n.storyIndex == nil {
		_, err := n.GetAllStories()
		if err != nil {
			return nil, fmt.Errorf("failed to build story index: %w", err)
		}
	}

	// Instant lookup from the index
	return n.storyIndex.GetStoriesForTheme(themeTitle), nil
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

// GetStoriesForType retrieves stories filtered by type
//
// Uses the pre-computed story index for instant O(1) lookup.
func (n *NocoDBAdapter) GetStoriesForType(typeTitle string) ([]data.Story, error) {
	// Ensure the index exists
	if n.storyIndex == nil {
		_, err := n.GetAllStories()
		if err != nil {
			return nil, fmt.Errorf("failed to build story index: %w", err)
		}
	}

	// Instant lookup from the index
	return n.storyIndex.GetStoriesForType(typeTitle), nil
}

// GetStoriesForWeather retrieves stories filtered by weather
//
// Uses the pre-computed story index for instant O(1) lookup.
func (n *NocoDBAdapter) GetStoriesForWeather(weatherTitle string) ([]data.Story, error) {
	// Ensure the index exists
	if n.storyIndex == nil {
		_, err := n.GetAllStories()
		if err != nil {
			return nil, fmt.Errorf("failed to build story index: %w", err)
		}
	}

	// Instant lookup from the index
	return n.storyIndex.GetStoriesForWeather(weatherTitle), nil
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
			story.URL = CreateStoryURLFromFindingWithID(story.Finding, story.ID)
		}

		stories = append(stories, story)
	}

	return stories, nil
}

// DropCache clears any cached data in the NocoDB client
//
// This also clears the story index, which will be rebuilt automatically on the next
// call to GetAllStories().
func (n *NocoDBAdapter) DropCache() error {
	n.client.DropCache()
	n.storyIndex = nil // Clear the index so it gets rebuilt
	return nil
}

// ClearDiskCache clears the disk cache file
func (n *NocoDBAdapter) ClearDiskCache() error {
	return n.client.ClearDiskCache()
}

// SetCacheOnlyMode enables cache-only mode for offline debugging
func (n *NocoDBAdapter) SetCacheOnlyMode(enabled bool) {
	n.client.SetCacheOnlyMode(enabled)
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

// GetStoriesForGiftedBy retrieves stories filtered by gifted by
//
// Uses the pre-computed story index for instant O(1) lookup.
func (n *NocoDBAdapter) GetStoriesForGiftedBy(giftedByTitle string) ([]data.Story, error) {
	// Ensure the index exists
	if n.storyIndex == nil {
		_, err := n.GetAllStories()
		if err != nil {
			return nil, fmt.Errorf("failed to build story index: %w", err)
		}
	}

	// Instant lookup from the index
	return n.storyIndex.GetStoriesForGiftedBy(giftedByTitle), nil
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

// GetStoriesForScalePermanence retrieves stories filtered by scale of permanence
//
// Uses the pre-computed story index for instant O(1) lookup.
func (n *NocoDBAdapter) GetStoriesForScalePermanence(scalePermanenceTitle string) ([]data.Story, error) {
	// Ensure the index exists
	if n.storyIndex == nil {
		_, err := n.GetAllStories()
		if err != nil {
			return nil, fmt.Errorf("failed to build story index: %w", err)
		}
	}

	// Instant lookup from the index
	return n.storyIndex.GetStoriesForScalePermanence(scalePermanenceTitle), nil
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

// GetStoriesForWhatWasIsIf retrieves stories filtered by what was/is/if
//
// Uses the pre-computed story index for instant O(1) lookup.
func (n *NocoDBAdapter) GetStoriesForWhatWasIsIf(whatWasIsIfTitle string) ([]data.Story, error) {
	// Ensure the index exists
	if n.storyIndex == nil {
		_, err := n.GetAllStories()
		if err != nil {
			return nil, fmt.Errorf("failed to build story index: %w", err)
		}
	}

	// Instant lookup from the index
	return n.storyIndex.GetStoriesForWhatWasIsIf(whatWasIsIfTitle), nil
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

// GetStoriesForTimePeriod retrieves stories filtered by time period
//
// Uses the pre-computed story index for instant O(1) lookup.
func (n *NocoDBAdapter) GetStoriesForTimePeriod(timePeriodTitle string) ([]data.Story, error) {
	// Ensure the index exists
	if n.storyIndex == nil {
		_, err := n.GetAllStories()
		if err != nil {
			return nil, fmt.Errorf("failed to build story index: %w", err)
		}
	}

	// Instant lookup from the index
	return n.storyIndex.GetStoriesForTimePeriod(timePeriodTitle), nil
}
