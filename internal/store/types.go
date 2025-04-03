// Retrieves and processes types as required.
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

// StoryDTO is a data transfer object that handles NULL values from the database
type StoryDTO struct {
	ID                      sql.NullString
	CreatedTime             sql.NullString
	Finding                 sql.NullString
	HighStExperiment        sql.NullString
	WhatWasIsIf             sql.NullString
	Image                   sql.NullString
	SourceImage             sql.NullString
	Location                sql.NullString
	StartDateTime           sql.NullString
	EndDateTime             sql.NullString
	Season                  sql.NullString
	Weather                 sql.NullString
	StreetDetectoristClue   sql.NullString
	Themes                  sql.NullString
	Experience              sql.NullString
	TimeSpan                sql.NullString
	OtherComments           sql.NullString
	Type                    sql.NullString
	PersonFinder            sql.NullString
	MapCache                sql.NullString
	MapSize                 sql.NullString
	Created                 sql.NullString
	StreetDetectoristMapURL sql.NullString
	OtherTheme              sql.NullString
	OtherWeather            sql.NullString
	ShareStatus             sql.NullString
	PostDate                sql.NullString
	TwitterText             sql.NullString
	CharacterCount          sql.NullString
	InstaText               sql.NullString
	InstaCount              sql.NullString
	InstaImage              sql.NullString
	ImageData               []byte
}

// ToStory converts the DTO to a domain model Story
func (dto *StoryDTO) ToStory() data.Story {
	return data.Story{
		ID:                      dto.ID.String,
		CreatedTime:             dto.CreatedTime.String,
		Finding:                 dto.Finding.String,
		HighStExperiment:        dto.HighStExperiment.String,
		WhatWasIsIf:             dto.WhatWasIsIf.String,
		Image:                   dto.Image.String,
		SourceImage:             dto.SourceImage.String,
		Location:                dto.Location.String,
		StartDateTime:           dto.StartDateTime.String,
		EndDateTime:             dto.EndDateTime.String,
		Season:                  dto.Season.String,
		Weather:                 dto.Weather.String,
		StreetDetectoristClue:   dto.StreetDetectoristClue.String,
		Themes:                  dto.Themes.String,
		Experience:              dto.Experience.String,
		TimeSpan:                dto.TimeSpan.String,
		OtherComments:           dto.OtherComments.String,
		Type:                    dto.Type.String,
		PersonFinder:            dto.PersonFinder.String,
		MapCache:                dto.MapCache.String,
		MapSize:                 dto.MapSize.String,
		Created:                 dto.Created.String,
		StreetDetectoristMapURL: dto.StreetDetectoristMapURL.String,
		OtherTheme:              dto.OtherTheme.String,
		OtherWeather:            dto.OtherWeather.String,
		ShareStatus:             dto.ShareStatus.String,
		PostDate:                dto.PostDate.String,
		TwitterText:             dto.TwitterText.String,
		CharacterCount:          dto.CharacterCount.String,
		InstaText:               dto.InstaText.String,
		InstaCount:              dto.InstaCount.String,
		InstaImage:              dto.InstaImage.String,
		ImageData:               dto.ImageData,
	}
}

// GetStoriesForType retrieves all stories for a given type from the database and returns them as a slice of Story.
func GetStoriesForType(typeTitle string) []data.Story {
	log.Println("Getting stories for type", typeTitle)

	dbPath := "airtable-export.db"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Type is stored as JSON array in the database like this:
	// ["Map", "Drawing", "Imagining"]
	// We use LIKE to query it, as it works okay for now and we control the data, which is static.
	likePattern := fmt.Sprintf("%%%q%%", typeTitle)

	query := `
		SELECT *
		FROM Stories
		WHERE "Type" LIKE ?;
	`

	rows, err := db.Query(query, likePattern)

	if err != nil {
		log.Fatalf("Failed to query stories: %v", err)
	}
	defer rows.Close()

	stories := []data.Story{}
	for rows.Next() {
		var dto StoryDTO
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
			&dto.ImageData,
		)

		if err != nil {
			log.Fatalf("Failed to scan story: %v", err)
		}

		stories = append(stories, dto.ToStory())
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

		if Type.Valid {
			// First unmarshal into a string array since it's in format ["Map", "Drawing", "Imagining"]
			var typeStrings []string
			if err := json.Unmarshal([]byte(Type.String), &typeStrings); err != nil {
				log.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			for _, typeStr := range typeStrings {
				newType := data.Type{Title: typeStr, URL: strings.ToLower(typeStr)}
				types = append(types, newType)
			}
		}
	}

	log.Printf("Found %d types", len(types))

	types = uniqueTypes(types)
	log.Printf("Found %d unique types", len(types))

	return types
}

// uniqueTypes returns a slice of unique types
func uniqueTypes(types []data.Type) []data.Type {
	seen := make(map[string]bool)
	unique := []data.Type{}

	// Loop over the slice and only keep first occurrence
	for _, t := range types {
		if !seen[t.Title] {
			seen[t.Title] = true
			unique = append(unique, t)
		}
	}

	return unique
}
