// Retrieves and processes types from the database as required.
package store

import (
	"database/sql"
	// "encoding/json" // No longer needed for GetTypes parsing
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/util"
)

// GetStoriesForType retrieves all stories for a given type from the database and returns them as a slice of Story.
func GetStoriesForType(typeTitle string) []data.Story {
	log.Println("Getting stories for type", typeTitle)

	db := ConnectToDatabase() // Use the centralized connection function
	if db == nil {
		log.Fatalf("Failed to connect to the database")
	}
	defer db.Close()

	// Type is stored as a comma-separated string in the database like this:
	// Map,Drawing,Imagining
	// We search for typeTitle as a whole word in the comma-separated list.
	query := `
		SELECT *
		FROM %s
		WHERE ("Type" = ? OR "Type" LIKE ? OR "Type" LIKE ? OR "Type" LIKE ?);
	`
	formattedQuery := fmt.Sprintf(query, StoriesTable()) // Use StoriesTable()

	arg1 := typeTitle               // Exact match: e.g., "Map"
	arg2 := typeTitle + ",%"        // Starts with: e.g., "Map,%"
	arg3 := "%," + typeTitle        // Ends with: e.g., "%,Map"
	arg4 := "%," + typeTitle + ",%" // Contains: e.g., "%,Map,%"

	rows, err := db.Query(formattedQuery, arg1, arg2, arg3, arg4)
	if err != nil {
		log.Fatalf("Failed to query stories for type %s: %v", typeTitle, err)
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
			log.Fatalf("Failed to scan story in GetStoriesForType: %v. Check column count and order.", err)
		}

		story := dto.ToStory()
		story.URL = CreateStoryURLFromFinding(story.Finding)

		stories = append(stories, story)
	}

	return stories
}

// GetTypes retrieves all unique types from the database.
func GetTypes() []data.Type {
	log.Println("Getting types")

	db := ConnectToDatabase()
	if db == nil {
		log.Fatalf("Failed to connect to the database for GetTypes")
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT Type FROM %s", StoriesTable())
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query types: %v", err)
	}
	defer rows.Close()

	var types []data.Type
	seen := make(map[string]bool) // For deduplication

	for rows.Next() {
		var typesStr sql.NullString
		if err := rows.Scan(&typesStr); err != nil {
			log.Fatalf("Failed to scan types string: %v", err)
		}

		if typesStr.Valid && typesStr.String != "" {
			typeItems := strings.Split(typesStr.String, ",")
			for _, typeTitle := range typeItems {
				trimmedTitle := strings.TrimSpace(typeTitle)
				if trimmedTitle == "" {
					continue
				}
				if !seen[trimmedTitle] {
					seen[trimmedTitle] = true
					newType := data.Type{
						Title:  trimmedTitle,
						URL:    "/types/" + util.Slugify(trimmedTitle) + ".html",
						Colour: data.TitleToHexColor(trimmedTitle),
					}
					types = append(types, newType)
				}
			}
		}
	}

	if err = rows.Err(); err != nil { // Check for errors during iteration
		log.Fatalf("Error iterating type rows: %v", err)
	}

	log.Printf("Found %d unique types", len(types))
	return types
}
