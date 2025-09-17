// Package store provides functions to store and retrieve stories from the database.
package store

import (
	"database/sql"
	"log"
	"math/rand"

	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/util"
)

// CreateStoryURLFromFinding creates a URL to the output file for a story page.
func CreateStoryURLFromFinding(finding string) string {
	slug := util.Slugify(finding)
	fileName := fmt.Sprintf("%s.html", slug)
	return filepath.Join("/stories", fileName)
}

func connectToDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "airtable-export.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	return db
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
