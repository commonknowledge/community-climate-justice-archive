// Retrieves and processes weather from the database as required.
package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/util"
)

// GetStoriesForWeather retrieves all stories for a given weather from the database and returns them as a slice of Story.
func GetStoriesForWeather(weatherTitle string) []data.Story {
	log.Println("Getting stories for weather", weatherTitle)

	dbPath := "airtable-export.db"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Weather is stored as JSON array in the database like this:
	// ["Sunny", "Cloudy", "Rainy"]
	// We use LIKE to query it, as it works okay for now and we control the data, which is static.
	likePattern := fmt.Sprintf("%%%q%%", weatherTitle)

	query := `
		SELECT *
		FROM Stories
		WHERE "Weather" LIKE ?;
	`

	rows, err := db.Query(query, likePattern)

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

// GetWeather retrieves all weather from the database and returns them as a slice of Weather.
// Intended for passing to HTML templates.
func GetWeather() []data.Weather {
	log.Println("Getting weather")

	dbPath := "airtable-export.db"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT Weather FROM Stories")
	if err != nil {
		log.Fatalf("Failed to query weather: %v", err)
	}
	defer rows.Close()

	var weathers []data.Weather

	for rows.Next() {
		var (
			Weather sql.NullString
		)

		if err := rows.Scan(&Weather); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		if Weather.Valid {
			// First unmarshal into a string array since it's in format ["Sunny", "Cloudy", "Rainy"]
			var weatherStrings []string
			if err := json.Unmarshal([]byte(Weather.String), &weatherStrings); err != nil {
				log.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			for _, weatherStr := range weatherStrings {
				newWeather := data.Weather{
					Title: weatherStr,
					URL:   "/weather/" + util.Slugify(weatherStr) + ".html",
				}
				weathers = append(weathers, newWeather)
			}
		}
	}

	log.Printf("Found %d weather conditions", len(weathers))

	weathers = uniqueWeather(weathers)
	log.Printf("Found %d unique weather conditions", len(weathers))

	return weathers
}

// uniqueWeather returns a slice of unique weather conditions.
func uniqueWeather(weathers []data.Weather) []data.Weather {
	seen := make(map[string]bool)
	unique := []data.Weather{}

	// Loop over the slice and only keep first occurrence of each weather.
	for _, w := range weathers {
		if !seen[w.Title] {
			seen[w.Title] = true
			unique = append(unique, w)
		}
	}

	return unique
}
