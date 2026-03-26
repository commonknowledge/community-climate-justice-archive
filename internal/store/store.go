// Package store is how the rest of the code gets story data from the database.
//
// This package talks to NocoDB (where all the stories live) and gives back the
// stories and tags that the rest of the application needs. It's the bridge between
// the database and everything else.
//
// How to use it:
// 1. Call Initialize() when the app starts - this sets up the connection to NocoDB
// 2. Then use functions like GetAllStories(), GetStoriesForTheme(), etc.
// 3. The NocoDB client handles raw-record caching, reducing repeat API calls
//
// Important scope note:
// - Raw NocoDB records are cached in internal/nocodb.
// - Converted []data.Story values are cached here for reuse across store calls.
package store

import (
	"fmt"
	"log"
	"sort"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/nocodb"
)

// client is the NocoDB client that talks to the database.
// It gets set up once when Initialize() is called, then used by everything else.
var client *nocodb.Client

// storiesCache stores converted Story data so we only do raw->Story conversion once
// per cache lifecycle.
var storiesCache []data.Story

// storiesCacheLoaded tracks whether storiesCache currently holds a valid dataset.
// We need this separate flag so an empty dataset is still treated as cached.
var storiesCacheLoaded bool

// resetStoriesCache clears converted story cache.
func resetStoriesCache() {
	storiesCache = nil
	storiesCacheLoaded = false
}

// Initialize sets up the connection to NocoDB.
// Call this once when the application starts, before using any other functions.
func Initialize() error {
	var err error
	client, err = nocodb.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create NocoDB client: %w", err)
	}
	resetStoriesCache()
	log.Println("Store initialised with NocoDB client")
	return nil
}

// WarmCache fetches all stories upfront so that subsequent calls are fast.
// Call this after Initialize() - it makes the first proper fetch from NocoDB
// and stores everything in memory for quick access later.
func WarmCache() {
	log.Println("Warming cache by fetching all stories...")
	allStories := GetAllStories()
	log.Printf("Cache warmed with %d stories", len(allStories))
}

// DropCache clears any cached data, forcing a fresh fetch from NocoDB next time.
// Useful if you know the data has changed and want to see the updates.
func DropCache() error {
	if client == nil {
		return fmt.Errorf("store not initialised")
	}
	resetStoriesCache()
	client.DropCache()
	return nil
}

// ClearDiskCache removes the disk cache file.
// The disk cache is a JSON file that stores stories between runs, so you don't
// have to wait for NocoDB every time you start the app during development.
func ClearDiskCache() error {
	if client == nil {
		return fmt.Errorf("store not initialised")
	}
	return client.ClearDiskCache()
}

// SetDiskCacheMode enables/disables disk cache for debugging.
//
// When disabled (default), the app always fetches fresh data from NocoDB and
// only keeps it in memory for the current run.
func SetDiskCacheMode(enabled bool) {
	if client != nil {
		client.SetDiskCacheMode(enabled)
	}
}

// SetCacheOnlyMode tells the store to only use cached data, never hitting the API.
// Really handy for offline debugging when you've got a cached copy of the data
// and don't want to (or can't) connect to NocoDB.
func SetCacheOnlyMode(enabled bool) {
	if client != nil {
		client.SetCacheOnlyMode(enabled)
	}
}

// GetRawRecords returns the raw NocoDB API response without any processing.
// This is only used for debugging - to see exactly what NocoDB is sending us
// before we turn it into Story structs.
func GetRawRecords() ([]map[string]interface{}, error) {
	if client == nil {
		return nil, fmt.Errorf("store not initialised")
	}
	return client.GetAllRecords()
}

// -------------------------------------------------------------------
// Stories
// -------------------------------------------------------------------

// GetAllStories fetches every approved story in the archive.
//
// This is probably the most-used function - it grabs all stories from NocoDB,
// converts them once per cache lifecycle, then filters to only include approved
// stories (Approved = "Yes-Live") before returning them.
//
// The NocoDB client cache prevents repeated network calls. Store-level story
// conversion caching prevents repeated raw-record conversion across store calls.
//
// If something goes wrong talking to the database, the program stops - we can't
// really do anything useful without story data.
func GetAllStories() []data.Story {
	if client == nil {
		log.Fatal("Store not initialised - call Initialize() first")
	}

	if storiesCacheLoaded {
		return storiesCache
	}

	records, err := client.GetAllRecords()
	if err != nil {
		log.Fatalf("Failed to get all records from NocoDB: %v", err)
	}

	allStories := convertRecordsToStories(records)

	// Filter to only include approved stories
	var approvedStories []data.Story
	for _, story := range allStories {
		if story.Approved == "Yes-Live" {
			approvedStories = append(approvedStories, story)
		}
	}

	storiesCache = approvedStories
	storiesCacheLoaded = true
	return storiesCache
}

// GetStoryByID finds a single story by its ID.
//
// This searches through all stories to find the one with the matching ID.
// Returns an empty Story if nothing matches.
func GetStoryByID(id string) (data.Story, error) {
	if client == nil {
		return data.Story{}, fmt.Errorf("store not initialised")
	}

	record, err := client.GetRecordByID(id)
	if err != nil {
		return data.Story{}, fmt.Errorf("failed to get record by ID %s: %w", id, err)
	}

	story, err := nocodb.NocoDBRecordToStoryWithClient(record, client)
	if err != nil {
		return data.Story{}, fmt.Errorf("failed to convert record to story: %w", err)
	}

	// Make sure the URL is set
	if story.URL == "" {
		story.URL = CreateStoryURLFromFindingWithID(story.Finding, story.ID)
	}

	// Sort tags alphabetically for consistent display
	sortStoryTags(&story)

	return story, nil
}

// GetStoriesWithConnections finds stories that have "inspired by" or "has inspired"
// relationships with other stories.
//
// It also includes all the stories they're connected to - so if story A inspired
// story B, you get both A and B back. This is useful for drawing connection lines
// on the wander page.
func GetStoriesWithConnections(limit int) []data.Story {
	allStories := GetAllStories()

	// Build a map for quick lookups by ID
	storyMap := make(map[string]data.Story)
	for _, story := range allStories {
		storyMap[story.ID] = story
	}

	// Find stories that have connections
	var connectedStories []data.Story
	for _, story := range allStories {
		if len(story.InspiredBy) > 0 || len(story.HasInspired) > 0 {
			connectedStories = append(connectedStories, story)
			if len(connectedStories) >= limit {
				break
			}
		}
	}

	// Also include the stories they're connected to
	resultMap := make(map[string]data.Story)

	// Add the primary connected stories
	for _, story := range connectedStories {
		resultMap[story.ID] = story
	}

	// Add their connection targets
	for _, story := range connectedStories {
		for _, connection := range story.InspiredBy {
			if targetStory, exists := storyMap[connection.ID]; exists {
				resultMap[connection.ID] = targetStory
			}
		}
		for _, connection := range story.HasInspired {
			if targetStory, exists := storyMap[connection.ID]; exists {
				resultMap[connection.ID] = targetStory
			}
		}
	}

	// Turn the map back into a list
	var result []data.Story
	for _, story := range resultMap {
		result = append(result, story)
	}

	return result
}

// convertRecordsToStories turns raw NocoDB records into proper Story structs.
func convertRecordsToStories(records []map[string]interface{}) []data.Story {
	var stories []data.Story

	for _, record := range records {
		story, err := nocodb.NocoDBRecordToStoryWithClient(record, client)
		if err != nil {
			log.Printf("Warning: failed to convert record to story: %v", err)
			continue
		}

		// Make sure URL is set
		if story.URL == "" {
			story.URL = CreateStoryURLFromFindingWithID(story.Finding, story.ID)
		}

		// Sort tags alphabetically for consistent display
		sortStoryTags(&story)

		stories = append(stories, story)
	}

	return stories
}

// sortStoryTags sorts all tag arrays on a story alphabetically by title.
// This ensures consistent ordering when tags are displayed.
func sortStoryTags(story *data.Story) {
	sort.Slice(story.Themes, func(i, j int) bool {
		return story.Themes[i].Title < story.Themes[j].Title
	})
	sort.Slice(story.Weather, func(i, j int) bool {
		return story.Weather[i].Title < story.Weather[j].Title
	})
	sort.Slice(story.Contributors, func(i, j int) bool {
		return story.Contributors[i].Name < story.Contributors[j].Name
	})
	sort.Slice(story.PublicContributors, func(i, j int) bool {
		return story.PublicContributors[i].Name < story.PublicContributors[j].Name
	})
	sort.Slice(story.GiftedBy, func(i, j int) bool {
		return story.GiftedBy[i].Title < story.GiftedBy[j].Title
	})
	sort.Slice(story.Type, func(i, j int) bool {
		return story.Type[i].Title < story.Type[j].Title
	})
	sort.Slice(story.ScalePermanence, func(i, j int) bool {
		return story.ScalePermanence[i].Title < story.ScalePermanence[j].Title
	})
	sort.Slice(story.WhatWasIsIf, func(i, j int) bool {
		return story.WhatWasIsIf[i].Title < story.WhatWasIsIf[j].Title
	})
	sort.Slice(story.TimePeriod, func(i, j int) bool {
		return story.TimePeriod[i].Title < story.TimePeriod[j].Title
	})
}
