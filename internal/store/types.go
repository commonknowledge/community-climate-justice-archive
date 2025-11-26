// Package store provides functions for retrieving story type and categorisation data.
//
// This file contains functions for working with story types and other categorisation
// systems in the archive. Story types classify the format or medium of a story,
// such as "Photo", "Poem", "Video", "Drawing", "Text", etc.
//
// Story types help visitors understand what kind of content they're about to view
// and enable filtering by preferred formats. Each story can have multiple types
// (for example, a story might be both a "Photo" and a "Poem").
//
// Additionally, this file manages other filtering categories:
// - Gifted By: Who contributed or co-created the story
// - Scale of Permanence: How long-lasting the subject matter is
// - What Was/Is/If: Temporal perspective (past, present, future)
// - Time Period: Historical era or timeframe
//
// All these categorisation systems help organise and browse the archive's stories.
package store

import (
	"log"

	"community-climate-justice-archive/data"
)

// GetStoriesForType retrieves all stories for a given type from the data source.
func GetStoriesForType(typeTitle string) []data.Story {
	adapter := GetAdapter()
	stories, err := adapter.GetStoriesForType(typeTitle)
	if err != nil {
		log.Fatalf("Failed to get stories for type: %v", err)
	}
	return stories
}

// GetTypes retrieves all types from the database and returns them as a slice of Type.
// Intended for passing to HTML templates.
func GetTypes() []data.Type {
	adapter := GetAdapter()
	types, err := adapter.GetTypes()
	if err != nil {
		log.Fatalf("Failed to get types: %v", err)
	}
	return types
}

// GetGiftedByTypes retrieves all gifted by types from the database and returns them as a slice of GiftedBy.
func GetGiftedByTypes() []data.GiftedBy {
	adapter := GetAdapter()
	giftedByTypes, err := adapter.GetGiftedByTypes()
	if err != nil {
		log.Fatalf("Failed to get gifted by types: %v", err)
	}
	return giftedByTypes
}

// GetStoriesForGiftedBy retrieves all stories for a given gifted by from the data source.
func GetStoriesForGiftedBy(giftedByTitle string) []data.Story {
	adapter := GetAdapter()
	stories, err := adapter.GetStoriesForGiftedBy(giftedByTitle)
	if err != nil {
		log.Fatalf("Failed to get stories for gifted by: %v", err)
	}
	return stories
}

// GetScalePermanenceTypes retrieves all scale of permanence types from the database and returns them as a slice of ScalePermanence.
func GetScalePermanenceTypes() []data.ScalePermanence {
	adapter := GetAdapter()
	scalePermanenceTypes, err := adapter.GetScalePermanenceTypes()
	if err != nil {
		log.Fatalf("Failed to get scale of permanence types: %v", err)
	}
	return scalePermanenceTypes
}

// GetStoriesForScalePermanence retrieves all stories for a given scale of permanence from the data source.
func GetStoriesForScalePermanence(scalePermanenceTitle string) []data.Story {
	adapter := GetAdapter()
	stories, err := adapter.GetStoriesForScalePermanence(scalePermanenceTitle)
	if err != nil {
		log.Fatalf("Failed to get stories for scale of permanence: %v", err)
	}
	return stories
}

// GetWhatWasIsIfTypes retrieves all what was/is/if types from the database and returns them as a slice of WhatWasIsIf.
func GetWhatWasIsIfTypes() []data.WhatWasIsIf {
	adapter := GetAdapter()
	whatWasIsIfTypes, err := adapter.GetWhatWasIsIfTypes()
	if err != nil {
		log.Fatalf("Failed to get what was/is/if types: %v", err)
	}
	return whatWasIsIfTypes
}

// GetStoriesForWhatWasIsIf retrieves all stories for a given what was/is/if from the data source.
func GetStoriesForWhatWasIsIf(whatWasIsIfTitle string) []data.Story {
	adapter := GetAdapter()
	stories, err := adapter.GetStoriesForWhatWasIsIf(whatWasIsIfTitle)
	if err != nil {
		log.Fatalf("Failed to get stories for what was/is/if: %v", err)
	}
	return stories
}

// GetTimePeriodTypes retrieves all time period types from the database and returns them as a slice of TimePeriod.
func GetTimePeriodTypes() []data.TimePeriod {
	adapter := GetAdapter()
	timePeriodTypes, err := adapter.GetTimePeriodTypes()
	if err != nil {
		log.Fatalf("Failed to get time period types: %v", err)
	}
	return timePeriodTypes
}

// GetStoriesForTimePeriod retrieves all stories for a given time period from the data source.
func GetStoriesForTimePeriod(timePeriodTitle string) []data.Story {
	adapter := GetAdapter()
	stories, err := adapter.GetStoriesForTimePeriod(timePeriodTitle)
	if err != nil {
		log.Fatalf("Failed to get stories for time period: %v", err)
	}
	return stories
}
