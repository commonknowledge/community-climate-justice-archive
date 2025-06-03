// Retrieves and processes types from the database as required.
package store

import (
	"database/sql"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/util"
)

// GetStoriesForType retrieves all stories for a given type from the database and returns them as a slice of Story.
func GetStoriesForType(typeTitle string) []data.Story {
	log.Println("Getting stories for type", typeTitle)

	dbPath := "nocodb.sqlite"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Type is stored as a comma-separated string in the database like this:
	// Map,Drawing,Imagining
	// We search for typeTitle as a whole word in the comma-separated list.
	query := `
		SELECT *
		FROM nc_9dus___Stories
		WHERE ("Type" = ? OR "Type" LIKE ? OR "Type" LIKE ? OR "Type" LIKE ?);
	`
	arg1 := typeTitle               // Exact match: e.g., "Map"
	arg2 := typeTitle + ",%"        // Starts with: e.g., "Map,%"
	arg3 := "%," + typeTitle        // Ends with: e.g., "%,Map"
	arg4 := "%," + typeTitle + ",%" // Contains: e.g., "%,Map,%"

	rows, err := db.Query(query, arg1, arg2, arg3, arg4)

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
			log.Fatalf("Failed to scan story in GetStoriesForType: %v", err)
		}

		story := dto.ToStory()
		story.URL = CreateStoryURLFromFinding(story.Finding)

		stories = append(stories, story)
	}

	return stories
}

// GetTypes retrieves all types from the database and returns them as a slice of Type.
// Intended for passing to HTML templates.
func GetTypes() []data.Type {
	log.Println("Getting types")

	dbPath := "airtable-export.db"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT Type FROM Stories")
	if err != nil {
		log.Fatalf("Failed to query types: %v", err)
	}
	defer rows.Close()

	var types []data.Type

	for rows.Next() {
		var (
			Type sql.NullString
		)

		if err := rows.Scan(&Type); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		if Type.Valid && Type.String != "" {
			// Types are now comma-separated: "Type1,Type2,Type3"
			typeStrings := strings.Split(Type.String, ",")

			for _, typeStr := range typeStrings {
				trimmedTypeStr := strings.TrimSpace(typeStr)
				newType := data.Type{Title: trimmedTypeStr, URL: "/types/" + util.Slugify(trimmedTypeStr) + ".html", Colour: data.TitleToHexColor(trimmedTypeStr)}
				types = append(types, newType)
			}
		}
	}

	log.Printf("Found %d types", len(types))

	types = uniqueTypes(types)
	log.Printf("Found %d unique types", len(types))

	return types
}

// uniqueTypes returns a slice of unique types.
func uniqueTypes(types []data.Type) []data.Type {
	seen := make(map[string]bool)
	unique := []data.Type{}

	// Loop over the slice and only keep first occurrence of each type.
	for _, t := range types {
		if !seen[t.Title] {
			seen[t.Title] = true
			unique = append(unique, t)
		}
	}

	return unique
}
