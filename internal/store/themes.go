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
			log.Fatalf("Failed to scan story in GetStoriesForTheme: %v", err)
		}

		story := dto.ToStory()
		story.URL = CreateStoryURLFromFinding(story.Finding)

		stories = append(stories, story)
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

		if Themes.Valid && Themes.String != "" {
			themeStrings := strings.Split(Themes.String, ",")

			for _, themeStr := range themeStrings {
				trimmedThemeStr := strings.TrimSpace(themeStr)
				newTheme := data.Theme{Title: trimmedThemeStr, URL: "/themes/" + util.Slugify(trimmedThemeStr) + ".html", Colour: data.TitleToHexColor(trimmedThemeStr)}
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
