// Retrieves and processes themes as required.
package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"community-climate-justice-archive/data"
)

// GetStoriesForTheme retrieves all stories for a given theme from the database and returns them as a slice of Story.
func GetStoriesForTheme(themeTitle string) []data.Story {
	log.Println("Getting stories for theme", themeTitle)

	dbPath := "airtable-export.db"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Themes are stored as JSON array in the database like this:
	// ["Theme1", "Theme2", "Theme3"]
	// We use LIKE to query it, as it works okay for now and we control the data, which is static.
	likePattern := fmt.Sprintf("%%%q%%", themeTitle)

	query := `
		SELECT *
		FROM Stories
		WHERE "Themes" LIKE ?;
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

		stories = append(stories, dto.ToStory())
	}

	return stories
}

// GetThemes retrieves all themes from the database and returns them as a slice of Theme.
// Intended for passing to HTML templates.
func GetThemes() []data.Theme {
	log.Println("Getting themes")

	dbPath := "airtable-export.db"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT Themes FROM Stories")
	if err != nil {
		log.Fatalf("Failed to query themes: %v", err)
	}
	defer rows.Close()

	var themes []data.Theme

	for rows.Next() {
		var (
			Themes sql.NullString
		)

		if err := rows.Scan(&Themes); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		if Themes.Valid {
			// First unmarshal into a string array since it's in format ["Theme1", "Theme2", "Theme3"]
			var themeStrings []string
			if err := json.Unmarshal([]byte(Themes.String), &themeStrings); err != nil {
				log.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			for _, themeStr := range themeStrings {
				newTheme := data.Theme{Title: themeStr, URL: strings.ToLower(themeStr)}
				themes = append(themes, newTheme)
			}
		}
	}

	log.Printf("Found %d themes", len(themes))

	themes = uniqueThemes(themes)
	log.Printf("Found %d unique themes", len(themes))

	return themes
}

// uniqueThemes returns a slice of unique themes.
func uniqueThemes(themes []data.Theme) []data.Theme {
	seen := make(map[string]bool)
	unique := []data.Theme{}

	// Loop over the slice and only keep first occurrence of each theme.
	for _, t := range themes {
		if !seen[t.Title] {
			seen[t.Title] = true
			unique = append(unique, t)
		}
	}

	return unique
}
