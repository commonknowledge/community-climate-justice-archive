// Package store is how the rest of the code gets story data from the database.
//
// This package talks to NocoDB (where all the stories live) and gives back the
// stories and tags that the rest of the application needs. It's the bridge between
// the database and everything else.
//
// How to use it:
// 1. Call Initialize() when the app starts - this sets up the connection to NocoDB
// 2. Then use functions like GetAllStories(), GetStoriesForTheme(), etc.
// 3. The NocoDB client handles caching, so calling things multiple times is fast
package store

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"sort"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/nocodb"
	"community-climate-justice-archive/internal/util"
)

// client is the NocoDB client that talks to the database.
// It gets set up once when Initialize() is called, then used by everything else.
var client *nocodb.Client

// Initialize sets up the connection to NocoDB.
// Call this once when the application starts, before using any other functions.
func Initialize() error {
	var err error
	client, err = nocodb.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create NocoDB client: %w", err)
	}
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

// GetAllStories fetches every story in the archive.
//
// This is probably the most-used function - it grabs all stories from NocoDB
// and returns them as a list. The NocoDB client handles caching, so calling
// this multiple times doesn't hammer the database.
//
// If something goes wrong talking to the database, the program stops - we can't
// really do anything useful without story data.
func GetAllStories() []data.Story {
	if client == nil {
		log.Fatal("Store not initialised - call Initialize() first")
	}

	records, err := client.GetAllRecords()
	if err != nil {
		log.Fatalf("Failed to get all records from NocoDB: %v", err)
	}

	stories := convertRecordsToStories(records)
	return stories
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

		stories = append(stories, story)
	}

	return stories
}

// -------------------------------------------------------------------
// Themes
// -------------------------------------------------------------------

// GetStoriesForTheme finds all stories tagged with a particular theme.
//
// Themes are things like "Climate Change", "Community", "Nature" - the big topics
// that stories can be about. This loops through all stories and returns the ones
// that have the given theme in their tags.
func GetStoriesForTheme(themeTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, theme := range story.Themes {
			if theme.Title == themeTitle {
				result = append(result, story)
				break // Found it, no need to check other themes on this story
			}
		}
	}

	return result
}

// GetThemes collects all the unique themes from across the archive.
//
// This looks at every story and gathers up all the themes that appear. Each theme
// only shows up once in the results, even if dozens of stories have it.
func GetThemes() []data.Theme {
	allStories := GetAllStories()

	// Use a map to collect unique themes
	themeMap := make(map[string]data.Theme)

	for _, story := range allStories {
		for _, theme := range story.Themes {
			if theme.Title != "" {
				themeMap[theme.Title] = theme
			}
		}
	}

	// Turn the map into a list
	var themes []data.Theme
	for _, theme := range themeMap {
		themes = append(themes, theme)
	}

	// Sort alphabetically by title
	sort.Slice(themes, func(i, j int) bool {
		return themes[i].Title < themes[j].Title
	})

	return themes
}

// -------------------------------------------------------------------
// Types
// -------------------------------------------------------------------

// GetStoriesForType finds all stories of a particular type.
//
// Types describe what form the story takes - "Photo", "Poem", "Video", "Drawing",
// "Text", and so on. A story can have multiple types (like a photo with a poem).
func GetStoriesForType(typeTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, typ := range story.Type {
			if typ.Title == typeTitle {
				result = append(result, story)
				break
			}
		}
	}

	return result
}

// GetTypes collects all the unique types from across the archive.
func GetTypes() []data.Type {
	allStories := GetAllStories()

	typeMap := make(map[string]data.Type)

	for _, story := range allStories {
		for _, typ := range story.Type {
			if typ.Title != "" {
				typeMap[typ.Title] = typ
			}
		}
	}

	var types []data.Type
	for _, typ := range typeMap {
		types = append(types, typ)
	}

	// Sort alphabetically by title
	sort.Slice(types, func(i, j int) bool {
		return types[i].Title < types[j].Title
	})

	return types
}

// -------------------------------------------------------------------
// Weather
// -------------------------------------------------------------------

// GetStoriesForWeather finds all stories tagged with a particular weather condition.
//
// Weather is a lovely way to browse the archive - was it sunny when this story
// happened? Rainy? Foggy? It adds an atmospheric dimension to exploring.
func GetStoriesForWeather(weatherTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, weather := range story.Weather {
			if weather.Title == weatherTitle {
				result = append(result, story)
				break
			}
		}
	}

	return result
}

// GetWeather collects all the unique weather conditions from across the archive.
func GetWeather() []data.Weather {
	allStories := GetAllStories()

	weatherMap := make(map[string]data.Weather)

	for _, story := range allStories {
		for _, weather := range story.Weather {
			if weather.Title != "" {
				weatherMap[weather.Title] = weather
			}
		}
	}

	var weather []data.Weather
	for _, w := range weatherMap {
		weather = append(weather, w)
	}

	// Sort alphabetically by title
	sort.Slice(weather, func(i, j int) bool {
		return weather[i].Title < weather[j].Title
	})

	return weather
}

// -------------------------------------------------------------------
// GiftedBy (Contributors)
// -------------------------------------------------------------------

// GetStoriesForGiftedBy finds all stories from a particular contributor.
//
// "Gifted by" tracks who shared or co-created each story - local schools,
// community groups, individuals. It's a nice way to celebrate everyone
// who's contributed to the archive.
func GetStoriesForGiftedBy(giftedByTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, giftedBy := range story.GiftedBy {
			if giftedBy.Title == giftedByTitle {
				result = append(result, story)
				break
			}
		}
	}

	return result
}

// GetGiftedByTypes collects all the unique contributors from across the archive.
func GetGiftedByTypes() []data.GiftedBy {
	allStories := GetAllStories()

	giftedByMap := make(map[string]data.GiftedBy)

	for _, story := range allStories {
		for _, giftedBy := range story.GiftedBy {
			if giftedBy.Title != "" {
				giftedByMap[giftedBy.Title] = giftedBy
			}
		}
	}

	var giftedByTypes []data.GiftedBy
	for _, giftedBy := range giftedByMap {
		giftedByTypes = append(giftedByTypes, giftedBy)
	}

	// Sort alphabetically by title
	sort.Slice(giftedByTypes, func(i, j int) bool {
		return giftedByTypes[i].Title < giftedByTypes[j].Title
	})

	return giftedByTypes
}

// -------------------------------------------------------------------
// Scale of Permanence
// -------------------------------------------------------------------

// GetStoriesForScalePermanence finds all stories with a particular permanence level.
//
// Scale of permanence comes from permaculture - it's about how long-lasting things
// are, from temporary to permanent. It's an interesting lens for thinking about
// the stories in the archive.
func GetStoriesForScalePermanence(scalePermanenceTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, sp := range story.ScalePermanence {
			if sp.Title == scalePermanenceTitle {
				result = append(result, story)
				break
			}
		}
	}

	return result
}

// GetScalePermanenceTypes collects all the unique permanence levels from the archive.
func GetScalePermanenceTypes() []data.ScalePermanence {
	allStories := GetAllStories()

	spMap := make(map[string]data.ScalePermanence)

	for _, story := range allStories {
		for _, sp := range story.ScalePermanence {
			if sp.Title != "" {
				spMap[sp.Title] = sp
			}
		}
	}

	var scalePermanenceTypes []data.ScalePermanence
	for _, sp := range spMap {
		scalePermanenceTypes = append(scalePermanenceTypes, sp)
	}

	// Sort alphabetically by title
	sort.Slice(scalePermanenceTypes, func(i, j int) bool {
		return scalePermanenceTypes[i].Title < scalePermanenceTypes[j].Title
	})

	return scalePermanenceTypes
}

// -------------------------------------------------------------------
// What Was/Is/If (Temporal Perspective)
// -------------------------------------------------------------------

// GetStoriesForWhatWasIsIf finds all stories with a particular temporal perspective.
//
// "What Was" is about the past, "What Is" about the present, "What If" about
// imagined futures. It's a lovely way to think about how stories relate to time.
func GetStoriesForWhatWasIsIf(whatWasIsIfTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, wwii := range story.WhatWasIsIf {
			if wwii.Title == whatWasIsIfTitle {
				result = append(result, story)
				break
			}
		}
	}

	return result
}

// GetWhatWasIsIfTypes collects all the unique temporal perspectives from the archive.
func GetWhatWasIsIfTypes() []data.WhatWasIsIf {
	allStories := GetAllStories()

	wwiiMap := make(map[string]data.WhatWasIsIf)

	for _, story := range allStories {
		for _, wwii := range story.WhatWasIsIf {
			if wwii.Title != "" {
				wwiiMap[wwii.Title] = wwii
			}
		}
	}

	var whatWasIsIfTypes []data.WhatWasIsIf
	for _, wwii := range wwiiMap {
		whatWasIsIfTypes = append(whatWasIsIfTypes, wwii)
	}

	// Sort alphabetically by title
	sort.Slice(whatWasIsIfTypes, func(i, j int) bool {
		return whatWasIsIfTypes[i].Title < whatWasIsIfTypes[j].Title
	})

	return whatWasIsIfTypes
}

// -------------------------------------------------------------------
// Time Period
// -------------------------------------------------------------------

// GetStoriesForTimePeriod finds all stories from a particular era.
//
// Time periods might be things like "1960s", "Victorian Era", "Present Day" -
// whenever the story relates to. It helps people explore stories from particular
// moments in history.
func GetStoriesForTimePeriod(timePeriodTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, tp := range story.TimePeriod {
			if tp.Title == timePeriodTitle {
				result = append(result, story)
				break
			}
		}
	}

	return result
}

// GetTimePeriodTypes collects all the unique time periods from the archive.
func GetTimePeriodTypes() []data.TimePeriod {
	allStories := GetAllStories()

	tpMap := make(map[string]data.TimePeriod)

	for _, story := range allStories {
		for _, tp := range story.TimePeriod {
			if tp.Title != "" {
				tpMap[tp.Title] = tp
			}
		}
	}

	var timePeriodTypes []data.TimePeriod
	for _, tp := range tpMap {
		timePeriodTypes = append(timePeriodTypes, tp)
	}

	// Sort alphabetically by title
	sort.Slice(timePeriodTypes, func(i, j int) bool {
		return timePeriodTypes[i].Title < timePeriodTypes[j].Title
	})

	return timePeriodTypes
}

// -------------------------------------------------------------------
// URL Helpers
// -------------------------------------------------------------------

// CreateStoryURLFromFinding makes a URL for a story page based on its title.
//
// It turns the title into a URL-safe slug. So "Climate Change Story" becomes
// "/stories/climate-change-story.html"
//
// Note: If two stories happen to have identical titles, they'd get the same URL,
// which would be a problem. Use CreateStoryURLFromFindingWithID instead to be safe.
func CreateStoryURLFromFinding(finding string) string {
	slug := util.Slugify(finding)
	fileName := fmt.Sprintf("%s.html", slug)
	return filepath.Join("/stories", fileName)
}

// CreateStoryURLFromFindingWithID makes a unique URL by including the story's ID.
//
// So story #42 with title "Climate Change" becomes "/stories/climate-change-42.html"
//
// This is the safer option because IDs are always unique, even if two stories
// happen to have the same title.
func CreateStoryURLFromFindingWithID(finding, id string) string {
	slug := util.Slugify(finding)
	fileName := fmt.Sprintf("%s-%s.html", slug, id)
	return filepath.Join("/stories", fileName)
}

// -------------------------------------------------------------------
// Helpers for Templates
// -------------------------------------------------------------------

// GetMoreTaggedStories finds other stories with the same tag, for "More like this" sections.
//
// Given a story and one of its tags (theme or type), this returns a random selection
// of other stories that share that tag. Useful for showing related content at the
// bottom of story pages.
func GetMoreTaggedStories(story data.Story, tag interface{}, count int) []data.Story {
	var stories []data.Story

	switch t := tag.(type) {
	case data.Theme:
		stories = GetStoriesForTheme(t.Title)
	case data.Type:
		stories = GetStoriesForType(t.Title)
	default:
		log.Printf("Warning: GetMoreTaggedStories called with unsupported tag type: %T", tag)
		return []data.Story{}
	}

	// Remove the current story from the list (we don't want to suggest itself)
	var filteredStories []data.Story
	for _, s := range stories {
		if s.ID != story.ID {
			filteredStories = append(filteredStories, s)
		}
	}

	// Shuffle randomly so you get different suggestions each time
	rand.Shuffle(len(filteredStories), func(i, j int) {
		filteredStories[i], filteredStories[j] = filteredStories[j], filteredStories[i]
	})

	// Return up to 'count' stories
	if len(filteredStories) <= count {
		return filteredStories
	}
	return filteredStories[:count]
}
