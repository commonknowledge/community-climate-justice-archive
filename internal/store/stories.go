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
	db := connectToDatabase()
	defer db.Close()

	rows, err := db.Query("SELECT * FROM Stories")
	if err != nil {
		log.Fatalf("Failed to query stories: %v", err)
	}
	defer rows.Close()

	stories := []data.Story{}
	for rows.Next() {
		var dto data.StoryDTO
		err := rows.Scan(
			&dto.ID,
			&dto.CreatedTime,
			&dto.Finding,
			&dto.HighStExperiment,
			&dto.WhatWasIsIf,
			&dto.Image,
			&dto.SourceImage,
			&dto.Location,
			&dto.StartDateTime,
			&dto.EndDateTime,
			&dto.Season,
			&dto.Weather,
			&dto.StreetDetectoristClue,
			&dto.Themes,
			&dto.Experience,
			&dto.TimeSpan,
			&dto.OtherComments,
			&dto.Type,
			&dto.PersonFinder,
			&dto.MapCache,
			&dto.MapSize,
			&dto.Created,
			&dto.StreetDetectoristMapURL,
			&dto.OtherTheme,
			&dto.OtherWeather,
			&dto.ShareStatus,
			&dto.PostDate,
			&dto.TwitterText,
			&dto.CharacterCount,
			&dto.InstaText,
			&dto.InstaCount,
			&dto.InstaImage,
		)

		if err != nil {
			log.Fatalf("Failed to scan story: %v", err)
		}

		story := dto.ToStory()

		story.URL = CreateStoryURLFromFinding(story.Finding)

		stories = append(stories, story)
	}

	return stories
}
