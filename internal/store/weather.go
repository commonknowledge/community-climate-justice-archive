// Retrieves and processes weather from the database as required.
package store

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

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
			log.Fatalf("Failed to scan story in GetStoriesForWeather: %v", err)
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

		if Weather.Valid && Weather.String != "" {
			// Weather conditions are now comma-separated: "Condition1,Condition2,Condition3"
			weatherStrings := strings.Split(Weather.String, ",")

			for _, weatherStr := range weatherStrings {
				trimmedWeatherStr := strings.TrimSpace(weatherStr)
				newWeather := data.Weather{
					Title:  trimmedWeatherStr,
					URL:    "/weather/" + util.Slugify(trimmedWeatherStr) + ".html",
					Colour: data.TitleToHexColor(trimmedWeatherStr),
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
