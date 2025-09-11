// Package store provides SQLite implementation of the data adapter
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

// SQLiteAdapter implements DataAdapter for SQLite database
type SQLiteAdapter struct{}

// GetAllStories retrieves all stories from SQLite database
func (s *SQLiteAdapter) GetAllStories() ([]data.Story, error) {
	db := s.connectToDatabase()
	defer db.Close()

	query := `SELECT * FROM Stories ORDER BY "CreatedTime" DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query stories: %w", err)
	}
	defer rows.Close()

	return s.scanStories(rows)
}

// GetStoriesForTheme retrieves all stories for a given theme from SQLite
func (s *SQLiteAdapter) GetStoriesForTheme(themeTitle string) ([]data.Story, error) {
	log.Println("Getting stories for theme", themeTitle)

	db := s.connectToDatabase()
	defer db.Close()

	// Themes are stored as JSON array - use LIKE for matching
	likePattern := fmt.Sprintf("%%%q%%", themeTitle)

	query := `
		SELECT *
		FROM Stories
		WHERE "Themes" LIKE ?;
	`

	rows, err := db.Query(query, likePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query stories for theme: %w", err)
	}
	defer rows.Close()

	return s.scanStories(rows)
}

// GetStoriesForType retrieves all stories for a given type from SQLite
func (s *SQLiteAdapter) GetStoriesForType(typeTitle string) ([]data.Story, error) {
	log.Println("Getting stories for type", typeTitle)

	db := s.connectToDatabase()
	defer db.Close()

	// Types are stored as JSON array - use LIKE for matching
	likePattern := fmt.Sprintf("%%%q%%", typeTitle)

	query := `
		SELECT *
		FROM Stories
		WHERE "Type" LIKE ?;
	`

	rows, err := db.Query(query, likePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query stories for type: %w", err)
	}
	defer rows.Close()

	return s.scanStories(rows)
}

// GetStoriesForWeather retrieves all stories for a given weather from SQLite
func (s *SQLiteAdapter) GetStoriesForWeather(weatherTitle string) ([]data.Story, error) {
	log.Println("Getting stories for weather", weatherTitle)

	db := s.connectToDatabase()
	defer db.Close()

	// Weather is stored as JSON array - use LIKE for matching
	likePattern := fmt.Sprintf("%%%q%%", weatherTitle)

	query := `
		SELECT *
		FROM Stories
		WHERE "Weather" LIKE ?;
	`

	rows, err := db.Query(query, likePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query stories for weather: %w", err)
	}
	defer rows.Close()

	return s.scanStories(rows)
}

// GetThemes retrieves all unique themes from SQLite
func (s *SQLiteAdapter) GetThemes() ([]data.Theme, error) {
	db := s.connectToDatabase()
	defer db.Close()

	query := `SELECT DISTINCT "Themes" FROM Stories WHERE "Themes" IS NOT NULL AND "Themes" != ''`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query themes: %w", err)
	}
	defer rows.Close()

	themeMap := make(map[string]data.Theme)

	for rows.Next() {
		var themesJSON sql.NullString
		if err := rows.Scan(&themesJSON); err != nil {
			continue
		}

		if !themesJSON.Valid || themesJSON.String == "" {
			continue
		}

		var themeStrings []string
		if err := json.Unmarshal([]byte(themesJSON.String), &themeStrings); err != nil {
			continue
		}

		for _, themeTitle := range themeStrings {
			if themeTitle != "" {
				themeMap[themeTitle] = data.Theme{
					Title:  themeTitle,
					URL:    "/themes/" + util.Slugify(themeTitle) + ".html",
					Colour: data.TitleToHexColor(themeTitle),
				}
			}
		}
	}

	// Convert map to slice
	var themes []data.Theme
	for _, theme := range themeMap {
		themes = append(themes, theme)
	}

	return themes, nil
}

// GetTypes retrieves all unique types from SQLite
func (s *SQLiteAdapter) GetTypes() ([]data.Type, error) {
	db := s.connectToDatabase()
	defer db.Close()

	query := `SELECT DISTINCT "Type" FROM Stories WHERE "Type" IS NOT NULL AND "Type" != ''`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query types: %w", err)
	}
	defer rows.Close()

	typeMap := make(map[string]data.Type)

	for rows.Next() {
		var typesJSON sql.NullString
		if err := rows.Scan(&typesJSON); err != nil {
			continue
		}

		if !typesJSON.Valid || typesJSON.String == "" {
			continue
		}

		var typeStrings []string
		if err := json.Unmarshal([]byte(typesJSON.String), &typeStrings); err != nil {
			continue
		}

		for _, typeTitle := range typeStrings {
			if typeTitle != "" {
				typeMap[typeTitle] = data.Type{
					Title:  typeTitle,
					URL:    "/types/" + util.Slugify(typeTitle) + ".html",
					Colour: data.TitleToHexColor(typeTitle),
				}
			}
		}
	}

	// Convert map to slice
	var types []data.Type
	for _, typeObj := range typeMap {
		types = append(types, typeObj)
	}

	return types, nil
}

// GetWeather retrieves all unique weather conditions from SQLite
func (s *SQLiteAdapter) GetWeather() ([]data.Weather, error) {
	db := s.connectToDatabase()
	defer db.Close()

	query := `SELECT DISTINCT "Weather" FROM Stories WHERE "Weather" IS NOT NULL AND "Weather" != ''`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query weather: %w", err)
	}
	defer rows.Close()

	weatherMap := make(map[string]data.Weather)

	for rows.Next() {
		var weatherJSON sql.NullString
		if err := rows.Scan(&weatherJSON); err != nil {
			continue
		}

		if !weatherJSON.Valid || weatherJSON.String == "" {
			continue
		}

		var weatherStrings []string
		if err := json.Unmarshal([]byte(weatherJSON.String), &weatherStrings); err != nil {
			continue
		}

		for _, weatherTitle := range weatherStrings {
			if weatherTitle != "" {
				weatherMap[weatherTitle] = data.Weather{
					Title:  weatherTitle,
					URL:    "/weather/" + util.Slugify(weatherTitle) + ".html",
					Colour: data.TitleToHexColor(weatherTitle),
				}
			}
		}
	}

	// Convert map to slice
	var weather []data.Weather
	for _, weatherObj := range weatherMap {
		weather = append(weather, weatherObj)
	}

	return weather, nil
}

// connectToDatabase opens a connection to the SQLite database
func (s *SQLiteAdapter) connectToDatabase() *sql.DB {
	db, err := sql.Open("sqlite3", "airtable-export.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	return db
}

// scanStories scans SQL rows into Story structs
func (s *SQLiteAdapter) scanStories(rows *sql.Rows) ([]data.Story, error) {
	var stories []data.Story

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
			return nil, fmt.Errorf("failed to scan story: %w", err)
		}

		story := dto.ToStory()
		story.URL = CreateStoryURLFromFinding(story.Finding)

		stories = append(stories, story)
	}

	return stories, nil
}

// DropCache is a no-op for SQLite since it doesn't use caching
func (s *SQLiteAdapter) DropCache() error {
	// SQLite adapter doesn't use caching, so nothing to drop
	log.Println("SQLite adapter cache drop requested (no-op)")
	return nil
}
