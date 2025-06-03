// Retrieves and processes themes as required.
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

// GetStoriesForTheme retrieves all stories for a given theme from the database and returns them as a slice of Story.
func GetStoriesForTheme(themeTitle string) []data.Story {
	log.Println("Getting stories for theme", themeTitle)

	db := ConnectToDatabase() // Use the centralized connection function
	if db == nil {
		log.Fatalf("Failed to connect to the database")
	}
	defer db.Close()

	// Themes are stored as a comma-separated string in the database.
	// We search for themeTitle as a whole word in the comma-separated list.
	query := `
		SELECT *
		FROM %s
		WHERE ("Themes" = ? OR "Themes" LIKE ? OR "Themes" LIKE ? OR "Themes" LIKE ?);
	`
	// Use StoriesTable() to get the correct table name
	formattedQuery := fmt.Sprintf(query, StoriesTable())

	arg1 := themeTitle               // Exact match
	arg2 := themeTitle + ",%"        // Starts with
	arg3 := "%," + themeTitle        // Ends with
	arg4 := "%," + themeTitle + ",%" // Contains

	rows, err := db.Query(formattedQuery, arg1, arg2, arg3, arg4)
	if err != nil {
		log.Fatalf("Failed to query stories for theme %s: %v", themeTitle, err)
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
			log.Fatalf("Failed to scan story in GetStoriesForTheme: %v. Check column count and order.", err)
		}

		story := dto.ToStory()
		story.URL = CreateStoryURLFromFinding(story.Finding) // CreateStoryURLFromFinding is in stories.go

		stories = append(stories, story)
	}

	return stories
}

// GetThemes retrieves all unique themes from the database.
func GetThemes() []data.Theme {
	log.Println("Getting themes")

	db := ConnectToDatabase()
	if db == nil {
		log.Fatalf("Failed to connect to the database for GetThemes")
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT Themes FROM %s", StoriesTable())
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query themes: %v", err)
	}
	defer rows.Close()

	var themes []data.Theme
	seen := make(map[string]bool) // For deduplication

	for rows.Next() {
		var themesStr sql.NullString
		if err := rows.Scan(&themesStr); err != nil {
			log.Fatalf("Failed to scan themes string: %v", err)
		}

		if themesStr.Valid && themesStr.String != "" {
			themeItems := strings.Split(themesStr.String, ",")
			for _, themeTitle := range themeItems {
				trimmedTitle := strings.TrimSpace(themeTitle)
				if trimmedTitle == "" {
					continue
				}
				if !seen[trimmedTitle] {
					seen[trimmedTitle] = true
					newTheme := data.Theme{
						Title:  trimmedTitle,
						URL:    "/themes/" + util.Slugify(trimmedTitle) + ".html",
						Colour: data.TitleToHexColor(trimmedTitle),
					}
					themes = append(themes, newTheme)
				}
			}
		}
	}

	if err = rows.Err(); err != nil { // Check for errors during iteration
		log.Fatalf("Error iterating theme rows: %v", err)
	}

	log.Printf("Found %d unique themes", len(themes))
	return themes
}
