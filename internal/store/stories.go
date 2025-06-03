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
	db, err := sql.Open("sqlite3", "nocodb.sqlite")
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

func StoriesTable() string {
	return "nc_9dus___Stories"
}

func GetAllStories() []data.Story {
	db := connectToDatabase()
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", StoriesTable()))
	if err != nil {
		log.Fatalf("Failed to query stories: %v", err)
	}
	defer rows.Close()

	stories := []data.Story{}
	for rows.Next() {
		var dto data.StoryDTO
		err := rows.Scan(
			&dto.ID,
			&dto.CreatedAt,
			&dto.UpdatedAt,
			&dto.CreatedBy,
			&dto.UpdatedBy,
			&dto.NCOrder,
			&dto.NCRecordID,
			&dto.NCRecordHash,
			&dto.Finding,
			&dto.Location,
			&dto.StartDateTime,
			&dto.EndDateTime,
			&dto.Weather,
			&dto.MapCache,
			&dto.MapSize,
			&dto.Type,
			&dto.Image,
			&dto.SourceImage,
			&dto.StreetDetectoristClue,
			&dto.Season,
			&dto.Themes,
			&dto.HighStExperiment,
			&dto.Experience,
			&dto.PersonFinderImaginerStreetDetectorist,
			&dto.IfYouWouldLikeToFillOutAStreetDetectorist,
			&dto.TimeSpan,
			&dto.OtherTheme,
			&dto.OtherWeather,
			&dto.OtherCommentsSources,
			&dto.WhatWasIsIf,
			&dto.ShareStatus,
			&dto.PostDate,
			&dto.TwitterText,
			&dto.InstaText,
			&dto.InstaImage,
			&dto.Created,
		)

		if err != nil {
			log.Fatalf("Failed to scan story in GetAllStories: %v", err)
		}

		story := dto.ToStory()

		story.URL = CreateStoryURLFromFinding(story.Finding)

		stories = append(stories, story)
	}

	return stories
}
