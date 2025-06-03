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

	db := ConnectToDatabase() // Use the centralized connection function
	if db == nil {
		log.Fatalf("Failed to connect to the database")
	}
	defer db.Close()

	// Weather is stored as a comma-separated string in the database.
	// We search for weatherTitle as a whole word in the comma-separated list.
	query := `
		SELECT *
		FROM %s
		WHERE ("Weather" = ? OR "Weather" LIKE ? OR "Weather" LIKE ? OR "Weather" LIKE ?);
	`
	formattedQuery := fmt.Sprintf(query, StoriesTable()) // Use StoriesTable()

	arg1 := weatherTitle               // Exact match
	arg2 := weatherTitle + ",%"        // Starts with
	arg3 := "%," + weatherTitle        // Ends with
	arg4 := "%," + weatherTitle + ",%" // Contains

	rows, err := db.Query(formattedQuery, arg1, arg2, arg3, arg4)
	if err != nil {
		log.Fatalf("Failed to query stories for weather %s: %v", weatherTitle, err)
	}
	defer rows.Close()

	stories := []data.Story{}
	for rows.Next() {
		var dto data.StoryDTO
		// Ensure this Scan call matches the StoryDTO structure and order exactly (36 fields)
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
			log.Fatalf("Failed to scan story in GetStoriesForWeather: %v. Check column count and order.", err)
		}

		story := dto.ToStory()
		story.URL = CreateStoryURLFromFinding(story.Finding)

		stories = append(stories, story)
	}

	return stories
}

// GetWeather retrieves all unique weather conditions from the database.
func GetWeather() []data.Weather {
	log.Println("Getting weather")

	db := ConnectToDatabase()
	if db == nil {
		log.Fatalf("Failed to connect to the database for GetWeather")
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT Weather FROM %s", StoriesTable())
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query weather: %v", err)
	}
	defer rows.Close()

	var weathers []data.Weather
	seen := make(map[string]bool) // For deduplication

	for rows.Next() {
		var weatherStr sql.NullString
		if err := rows.Scan(&weatherStr); err != nil {
			log.Fatalf("Failed to scan weather string: %v", err)
		}

		if weatherStr.Valid && weatherStr.String != "" {
			weatherItems := strings.Split(weatherStr.String, ",")
			for _, weatherTitle := range weatherItems {
				trimmedTitle := strings.TrimSpace(weatherTitle)
				if trimmedTitle == "" {
					continue
				}
				if !seen[trimmedTitle] {
					seen[trimmedTitle] = true
					newWeather := data.Weather{
						Title:  trimmedTitle,
						URL:    "/weather/" + util.Slugify(trimmedTitle) + ".html",
						Colour: data.TitleToHexColor(trimmedTitle),
					}
					weathers = append(weathers, newWeather)
				}
			}
		}
	}

	if err = rows.Err(); err != nil { // Check for errors during iteration
		log.Fatalf("Error iterating weather rows: %v", err)
	}

	log.Printf("Found %d unique weather conditions", len(weathers))
	return weathers
}
