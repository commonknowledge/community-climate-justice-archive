package store

import (
	"community-climate-justice-archive/data"
	"log"
)

// StoryIndex holds pre-computed indexes for fast story lookups by taxonomy
//
// How It Works:
// When building the archive, we need to generate separate pages for each theme,
// type, weather condition, etc. Rather than scanning through all stories each time
// we need a specific taxonomy, we build lookup tables up front.
//
// We scan through all stories ONCE and populate maps that let us instantly find
// "all stories with theme Climate Change" or "all stories with type Photo" without
// any searching.
//
// Performance:
// This approach is O(n + m) where n = number of stories to index, m = number of lookups.
// With 500 stories and 45 taxonomy pages, that's 545 operations total. The lookup
// itself is O(1) constant time - instant regardless of archive size.
//
// Memory Trade-off:
// Uses additional memory to store story references in multiple maps. With 500 stories,
// this adds approximately 2-3 MB. The speed benefit is well worth this small memory cost.
type StoryIndex struct {
	// Maps from theme title to all stories with that theme
	// e.g., "Climate Change" → [story1, story2, story5, ...]
	themeIndex map[string][]data.Story

	// Maps from type title to all stories with that type
	// e.g., "Photo" → [story1, story3, story7, ...]
	typeIndex map[string][]data.Story

	// Maps from weather title to all stories with that weather
	// e.g., "Sunny" → [story2, story4, story9, ...]
	weatherIndex map[string][]data.Story

	// Maps from "Gifted or co-created by" value to stories
	giftedByIndex map[string][]data.Story

	// Maps from "Scale of permanence" value to stories
	scalePermanenceIndex map[string][]data.Story

	// Maps from "What was/is/if" value to stories
	whatWasIsIfIndex map[string][]data.Story

	// Maps from "Time period" value to stories
	timePeriodIndex map[string][]data.Story
}

// BuildStoryIndex creates a new story index from all stories
//
// This is the function that does the "scan once" operation. It goes through
// every story and adds it to the appropriate lookup tables based on its tags.
//
// How it works:
// 1. Create empty maps for each taxonomy type
// 2. Loop through all stories once
// 3. For each story, look at its tags (themes, types, weather, etc.)
// 4. Add the story to the appropriate map entries
//
// Example:
// If story #42 has themes ["Climate Change", "Activism"], we add it to both:
//   - themeIndex["Climate Change"] = [...existing stories..., story #42]
//   - themeIndex["Activism"] = [...existing stories..., story #42]
//
// Note: Stories are not copied, just referenced. So the same story struct can
// appear in multiple index entries without duplicating memory.
func BuildStoryIndex(allStories []data.Story) *StoryIndex {
	log.Println("Building story indexes for fast taxonomy lookups...")

	// Initialise empty maps for each taxonomy type
	index := &StoryIndex{
		themeIndex:           make(map[string][]data.Story),
		typeIndex:            make(map[string][]data.Story),
		weatherIndex:         make(map[string][]data.Story),
		giftedByIndex:        make(map[string][]data.Story),
		scalePermanenceIndex: make(map[string][]data.Story),
		whatWasIsIfIndex:     make(map[string][]data.Story),
		timePeriodIndex:      make(map[string][]data.Story),
	}

	// Loop through all stories once and populate the indexes
	for _, story := range allStories {
		// Add this story to all its theme indexes
		// A story can have multiple themes, so it might appear in several lists
		for _, theme := range story.Themes {
			index.themeIndex[theme.Title] = append(index.themeIndex[theme.Title], story)
		}

		// Add this story to all its type indexes
		for _, typ := range story.Type {
			index.typeIndex[typ.Title] = append(index.typeIndex[typ.Title], story)
		}

		// Add this story to all its weather indexes
		for _, weather := range story.Weather {
			index.weatherIndex[weather.Title] = append(index.weatherIndex[weather.Title], story)
		}

		// Add this story to all its "Gifted By" indexes
		for _, giftedBy := range story.GiftedBy {
			index.giftedByIndex[giftedBy.Title] = append(index.giftedByIndex[giftedBy.Title], story)
		}

		// Add this story to all its "Scale of Permanence" indexes
		for _, scalePermanence := range story.ScalePermanence {
			index.scalePermanenceIndex[scalePermanence.Title] = append(index.scalePermanenceIndex[scalePermanence.Title], story)
		}

		// Add this story to all its "What Was/Is/If" indexes
		for _, whatWasIsIf := range story.WhatWasIsIf {
			index.whatWasIsIfIndex[whatWasIsIf.Title] = append(index.whatWasIsIfIndex[whatWasIsIf.Title], story)
		}

		// Add this story to all its "Time Period" indexes
		for _, timePeriod := range story.TimePeriod {
			index.timePeriodIndex[timePeriod.Title] = append(index.timePeriodIndex[timePeriod.Title], story)
		}
	}

	// Log some statistics so we can see what was indexed
	log.Printf("Story index built: %d themes, %d types, %d weather conditions, %d gifted by, %d scale permanence, %d what was/is/if, %d time periods",
		len(index.themeIndex),
		len(index.typeIndex),
		len(index.weatherIndex),
		len(index.giftedByIndex),
		len(index.scalePermanenceIndex),
		len(index.whatWasIsIfIndex),
		len(index.timePeriodIndex))

	return index
}

// GetStoriesForTheme returns all stories with a given theme
//
// This is a simple map lookup - O(1) constant time, regardless of how many
// stories exist. Compare this to the old approach which scanned every story.
//
// If the theme doesn't exist, returns an empty slice (not nil, to avoid
// nil pointer errors in templates).
func (idx *StoryIndex) GetStoriesForTheme(themeTitle string) []data.Story {
	stories, exists := idx.themeIndex[themeTitle]
	if !exists {
		// Return empty slice rather than nil to avoid template errors
		return []data.Story{}
	}
	return stories
}

// GetStoriesForType returns all stories with a given type
//
// Same O(1) instant lookup as GetStoriesForTheme, just for the type taxonomy.
func (idx *StoryIndex) GetStoriesForType(typeTitle string) []data.Story {
	stories, exists := idx.typeIndex[typeTitle]
	if !exists {
		return []data.Story{}
	}
	return stories
}

// GetStoriesForWeather returns all stories with a given weather condition
//
// Same O(1) instant lookup as above, for weather taxonomy.
func (idx *StoryIndex) GetStoriesForWeather(weatherTitle string) []data.Story {
	stories, exists := idx.weatherIndex[weatherTitle]
	if !exists {
		return []data.Story{}
	}
	return stories
}

// GetStoriesForGiftedBy returns all stories with a given "Gifted or co-created by" value
//
// Same O(1) instant lookup as above, for the "Gifted By" taxonomy.
func (idx *StoryIndex) GetStoriesForGiftedBy(giftedByTitle string) []data.Story {
	stories, exists := idx.giftedByIndex[giftedByTitle]
	if !exists {
		return []data.Story{}
	}
	return stories
}

// GetStoriesForScalePermanence returns all stories with a given "Scale of permanence" value
//
// Same O(1) instant lookup as above, for the "Scale of Permanence" taxonomy.
func (idx *StoryIndex) GetStoriesForScalePermanence(scalePermanenceTitle string) []data.Story {
	stories, exists := idx.scalePermanenceIndex[scalePermanenceTitle]
	if !exists {
		return []data.Story{}
	}
	return stories
}

// GetStoriesForWhatWasIsIf returns all stories with a given "What was/is/if" value
//
// Same O(1) instant lookup as above, for the "What Was/Is/If" taxonomy.
func (idx *StoryIndex) GetStoriesForWhatWasIsIf(whatWasIsIfTitle string) []data.Story {
	stories, exists := idx.whatWasIsIfIndex[whatWasIsIfTitle]
	if !exists {
		return []data.Story{}
	}
	return stories
}

// GetStoriesForTimePeriod returns all stories with a given time period
//
// Same O(1) instant lookup as above, for the "Time Period" taxonomy.
func (idx *StoryIndex) GetStoriesForTimePeriod(timePeriodTitle string) []data.Story {
	stories, exists := idx.timePeriodIndex[timePeriodTitle]
	if !exists {
		return []data.Story{}
	}
	return stories
}

