// Package store has the functions for fetching stories from the database.
//
// Stories are what the whole archive is about - each one is something the community
// has contributed, like a photo, poem, drawing, or video.
//
// This file lets you:
// - Get all the stories
// - Get a specific story by its ID
// - Get stories filtered by tags (themes, types, weather, etc.)
// - Find stories that are connected to each other
// - Create the right URLs for story pages
//
// All the actual database stuff happens through the "adapter" (see adapter.go),
// so these functions don't need to know if we're using NocoDB or SQLite or whatever -
// they just ask the adapter for stories and the adapter worries about the details.
package store

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/util"
)

// CreateStoryURLFromFinding creates a URL for a story's page based on its title.
//
// This function generates the web URL where a story's page will be located.
// It takes the story's "Finding" (title) and converts it into a URL-safe slug.
// For example, "Climate Change Story" becomes "/stories/climate-change-story.html"
//
// Parameters:
// - finding: The story's title (the Finding field)
//
// Returns:
// A URL string like "/stories/climate-change-story.html"
//
// Note: This version doesn't include the story ID in the URL. If two stories have
// identical titles (after slugification), they would have the same URL, which could
// cause conflicts. Consider using CreateStoryURLFromFindingWithID instead.
func CreateStoryURLFromFinding(finding string) string {
	slug := util.Slugify(finding)
	fileName := fmt.Sprintf("%s.html", slug)
	return filepath.Join("/stories", fileName)
}

// CreateStoryURLFromFindingWithID creates a unique URL for a story's page.
//
// This function generates the web URL where a story's page will be located, ensuring
// uniqueness by including the story's ID. For example, if story #42 has the title
// "Climate Change Story", the URL becomes "/stories/climate-change-story-42.html"
//
// This is the preferred URL generation method because:
// - It guarantees URL uniqueness even if stories have identical titles
// - It allows stories to be identified by their ID from the URL
// - It's more resilient to title changes
//
// Parameters:
// - finding: The story's title (the Finding field)
// - id: The story's unique identifier from the database
//
// Returns:
// A URL string like "/stories/climate-change-story-42.html"
func CreateStoryURLFromFindingWithID(finding, id string) string {
	slug := util.Slugify(finding)
	fileName := fmt.Sprintf("%s-%s.html", slug, id)
	return filepath.Join("/stories", fileName)
}

// GetMoreTaggedStories gets the first 3 more tagged stories for a story.
func GetMoreTaggedStories(story data.Story, tag interface{}, count int) []data.Story {
	var stories []data.Story

	switch tag := tag.(type) {
	case data.Theme:
		stories = GetStoriesForTheme(tag.Title)

		// Randomly shuffle the stories
		rand.Shuffle(len(stories), func(i, j int) {
			stories[i], stories[j] = stories[j], stories[i]
		})

		if len(stories) < count {
			return stories
		}

		return stories[:count]
	case data.Type:
		stories = GetStoriesForType(tag.Title)

		// Randomly shuffle the stories
		rand.Shuffle(len(stories), func(i, j int) {
			stories[i], stories[j] = stories[j], stories[i]
		})

		if len(stories) < count {
			return stories
		}

		return stories[:count]
	default:
		log.Fatalf("Unsupported tag type: %T", tag)
	}

	return stories
}

// GetAllStories retrieves every story in the archive from the database.
//
// This is one of the most frequently called functions in the application. It fetches
// all stories from the configured data adapter (currently NocoDB) and returns them
// as a slice of Story structs.
//
// The function uses the data adapter's caching mechanisms, so repeated calls are
// fast and don't hit the database every time.
//
// Returns:
// A slice containing all Story structs in the archive. The stories include all their
// data: titles, descriptions, attachments, tags, relationships, etc.
//
// If the database query fails, the program terminates (this is intentional - the
// application cannot function without story data).
//
// Usage:
// This function is used when:
// - Generating the archive overview page
// - Building filter data for the JavaScript-based filtering system
// - Warming the cache at application startup
func GetAllStories() []data.Story {
	adapter := GetAdapter()
	stories, err := adapter.GetAllStories()
	if err != nil {
		log.Fatalf("Failed to get all stories: %v", err)
	}
	return stories
}

// GetStoriesWithConnections retrieves stories that have relationship connections
func GetStoriesWithConnections(limit int) []data.Story {
	adapter := GetAdapter()
	stories, err := adapter.GetStoriesWithConnections(limit)
	if err != nil {
		log.Fatalf("Failed to get connected stories: %v", err)
	}
	return stories
}
