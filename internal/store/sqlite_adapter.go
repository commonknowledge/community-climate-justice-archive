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

// GetStoryByID retrieves a single story by its ID from SQLite database
func (s *SQLiteAdapter) GetStoryByID(id string) (data.Story, error) {
	db := s.connectToDatabase()
	defer db.Close()

	query := `SELECT * FROM Stories WHERE "ID" = ?`
	rows, err := db.Query(query, id)
	if err != nil {
		return data.Story{}, fmt.Errorf("failed to query story by ID %s: %w", id, err)
	}
	defer rows.Close()

	stories, err := s.scanStories(rows)
	if err != nil {
		return data.Story{}, fmt.Errorf("failed to scan story: %w", err)
	}

	if len(stories) == 0 {
		return data.Story{}, fmt.Errorf("story with ID %s not found", id)
	}

	return stories[0], nil
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

// GetStoriesWithConnections retrieves stories that have InspiredBy or HasInspired relationships from SQLite
// and includes all their connection targets to ensure lines can be drawn
func (s *SQLiteAdapter) GetStoriesWithConnections(limit int) ([]data.Story, error) {
	allStories, err := s.GetAllStories()
	if err != nil {
		return nil, fmt.Errorf("failed to get all stories: %w", err)
	}

	// Create a map for fast lookup
	storyMap := make(map[string]data.Story)
	for _, story := range allStories {
		storyMap[story.ID] = story
	}

	// First, find stories with connections
	var primaryConnectedStories []data.Story
	for _, story := range allStories {
		if len(story.InspiredBy) > 0 || len(story.HasInspired) > 0 {
			primaryConnectedStories = append(primaryConnectedStories, story)
			if len(primaryConnectedStories) >= limit {
				break
			}
		}
	}

	// Now add all their connection targets to ensure lines can be drawn
	storySet := make(map[string]data.Story)

	// Add primary connected stories
	for _, story := range primaryConnectedStories {
		storySet[story.ID] = story
	}

	// Add all their connection targets
	for _, story := range primaryConnectedStories {
		// Add InspiredBy targets
		for _, connection := range story.InspiredBy {
			if targetStory, exists := storyMap[connection.ID]; exists {
				storySet[connection.ID] = targetStory
			}
		}
		// Add HasInspired targets
		for _, connection := range story.HasInspired {
			if targetStory, exists := storyMap[connection.ID]; exists {
				storySet[connection.ID] = targetStory
			}
		}
	}

	// Convert back to slice
	var result []data.Story
	for _, story := range storySet {
		result = append(result, story)
	}

	return result, nil
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
		story.URL = CreateStoryURLFromFindingWithID(story.Finding, story.ID)

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

// ClearDiskCache is a no-op for SQLite since it doesn't use disk caching
func (s *SQLiteAdapter) ClearDiskCache() error {
	// SQLite adapter doesn't use disk caching, so nothing to clear
	log.Println("SQLite adapter disk cache clear requested (no-op)")
	return nil
}

// SetCacheOnlyMode is a no-op for SQLite since it doesn't use external caching
func (s *SQLiteAdapter) SetCacheOnlyMode(enabled bool) {
	// SQLite adapter doesn't use external APIs or caching, so this is a no-op
	log.Println("SQLite adapter cache-only mode requested (no-op)")
}

// GetGiftedByTypes retrieves all unique gifted by values from SQLite
func (s *SQLiteAdapter) GetGiftedByTypes() ([]data.GiftedBy, error) {
	db := s.connectToDatabase()
	defer db.Close()

	query := `SELECT DISTINCT "Gifted or co-created by…" FROM Stories WHERE "Gifted or co-created by…" IS NOT NULL AND "Gifted or co-created by…" != ''`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query gifted by: %w", err)
	}
	defer rows.Close()

	giftedByMap := make(map[string]data.GiftedBy)

	for rows.Next() {
		var giftedByJSON sql.NullString
		if err := rows.Scan(&giftedByJSON); err != nil {
			continue
		}

		if !giftedByJSON.Valid || giftedByJSON.String == "" {
			continue
		}

		var giftedByStrings []string
		if err := json.Unmarshal([]byte(giftedByJSON.String), &giftedByStrings); err != nil {
			continue
		}

		for _, giftedByTitle := range giftedByStrings {
			if giftedByTitle != "" {
				giftedByMap[giftedByTitle] = data.GiftedBy{
					Title:  giftedByTitle,
					URL:    "/giftedby/" + util.Slugify(giftedByTitle) + ".html",
					Colour: data.TitleToHexColor(giftedByTitle),
				}
			}
		}
	}

	// Convert map to slice
	var giftedBy []data.GiftedBy
	for _, giftedByObj := range giftedByMap {
		giftedBy = append(giftedBy, giftedByObj)
	}

	return giftedBy, nil
}

// GetStoriesForGiftedBy retrieves all stories for a given gifted by value from SQLite
func (s *SQLiteAdapter) GetStoriesForGiftedBy(giftedByTitle string) ([]data.Story, error) {
	log.Println("Getting stories for gifted by", giftedByTitle)

	db := s.connectToDatabase()
	defer db.Close()

	// Gifted by is stored as JSON array - use LIKE for matching
	likePattern := fmt.Sprintf("%%%q%%", giftedByTitle)

	query := `
		SELECT *
		FROM Stories
		WHERE "Gifted or co-created by…" LIKE ?;
	`

	rows, err := db.Query(query, likePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query stories for gifted by: %w", err)
	}
	defer rows.Close()

	return s.scanStories(rows)
}

// GetScalePermanenceTypes retrieves all unique scale of permanence values from SQLite
func (s *SQLiteAdapter) GetScalePermanenceTypes() ([]data.ScalePermanence, error) {
	db := s.connectToDatabase()
	defer db.Close()

	query := `SELECT DISTINCT "Scale of permanence" FROM Stories WHERE "Scale of permanence" IS NOT NULL AND "Scale of permanence" != ''`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query scale of permanence: %w", err)
	}
	defer rows.Close()

	scalePermanenceMap := make(map[string]data.ScalePermanence)

	for rows.Next() {
		var scalePermanenceJSON sql.NullString
		if err := rows.Scan(&scalePermanenceJSON); err != nil {
			continue
		}

		if !scalePermanenceJSON.Valid || scalePermanenceJSON.String == "" {
			continue
		}

		var scalePermanenceStrings []string
		if err := json.Unmarshal([]byte(scalePermanenceJSON.String), &scalePermanenceStrings); err != nil {
			continue
		}

		for _, scalePermanenceTitle := range scalePermanenceStrings {
			if scalePermanenceTitle != "" {
				scalePermanenceMap[scalePermanenceTitle] = data.ScalePermanence{
					Title:  scalePermanenceTitle,
					URL:    "/scalepermanence/" + util.Slugify(scalePermanenceTitle) + ".html",
					Colour: data.TitleToHexColor(scalePermanenceTitle),
				}
			}
		}
	}

	// Convert map to slice
	var scalePermanence []data.ScalePermanence
	for _, scalePermanenceObj := range scalePermanenceMap {
		scalePermanence = append(scalePermanence, scalePermanenceObj)
	}

	return scalePermanence, nil
}

// GetStoriesForScalePermanence retrieves all stories for a given scale of permanence value from SQLite
func (s *SQLiteAdapter) GetStoriesForScalePermanence(scalePermanenceTitle string) ([]data.Story, error) {
	log.Println("Getting stories for scale of permanence", scalePermanenceTitle)

	db := s.connectToDatabase()
	defer db.Close()

	// Scale of permanence is stored as JSON array - use LIKE for matching
	likePattern := fmt.Sprintf("%%%q%%", scalePermanenceTitle)

	query := `
		SELECT *
		FROM Stories
		WHERE "Scale of permanence" LIKE ?;
	`

	rows, err := db.Query(query, likePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query stories for scale of permanence: %w", err)
	}
	defer rows.Close()

	return s.scanStories(rows)
}

// GetWhatWasIsIfTypes retrieves all unique what was/is/if values from SQLite
func (s *SQLiteAdapter) GetWhatWasIsIfTypes() ([]data.WhatWasIsIf, error) {
	db := s.connectToDatabase()
	defer db.Close()

	query := `SELECT DISTINCT "What was/is/if" FROM Stories WHERE "What was/is/if" IS NOT NULL AND "What was/is/if" != ''`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query what was/is/if: %w", err)
	}
	defer rows.Close()

	whatWasIsIfMap := make(map[string]data.WhatWasIsIf)

	for rows.Next() {
		var whatWasIsIfJSON sql.NullString
		if err := rows.Scan(&whatWasIsIfJSON); err != nil {
			continue
		}

		if !whatWasIsIfJSON.Valid || whatWasIsIfJSON.String == "" {
			continue
		}

		var whatWasIsIfStrings []string
		if err := json.Unmarshal([]byte(whatWasIsIfJSON.String), &whatWasIsIfStrings); err != nil {
			continue
		}

		for _, whatWasIsIfTitle := range whatWasIsIfStrings {
			if whatWasIsIfTitle != "" {
				whatWasIsIfMap[whatWasIsIfTitle] = data.WhatWasIsIf{
					Title:  whatWasIsIfTitle,
					URL:    "/whatwasisif/" + util.Slugify(whatWasIsIfTitle) + ".html",
					Colour: data.TitleToHexColor(whatWasIsIfTitle),
				}
			}
		}
	}

	// Convert map to slice
	var whatWasIsIf []data.WhatWasIsIf
	for _, whatWasIsIfObj := range whatWasIsIfMap {
		whatWasIsIf = append(whatWasIsIf, whatWasIsIfObj)
	}

	return whatWasIsIf, nil
}

// GetStoriesForWhatWasIsIf retrieves all stories for a given what was/is/if value from SQLite
func (s *SQLiteAdapter) GetStoriesForWhatWasIsIf(whatWasIsIfTitle string) ([]data.Story, error) {
	log.Println("Getting stories for what was/is/if", whatWasIsIfTitle)

	db := s.connectToDatabase()
	defer db.Close()

	// What was/is/if is stored as JSON array - use LIKE for matching
	likePattern := fmt.Sprintf("%%%q%%", whatWasIsIfTitle)

	query := `
		SELECT *
		FROM Stories
		WHERE "What was/is/if" LIKE ?;
	`

	rows, err := db.Query(query, likePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query stories for what was/is/if: %w", err)
	}
	defer rows.Close()

	return s.scanStories(rows)
}

// GetTimePeriodTypes retrieves all unique time period values from SQLite
func (s *SQLiteAdapter) GetTimePeriodTypes() ([]data.TimePeriod, error) {
	db := s.connectToDatabase()
	defer db.Close()

	query := `SELECT DISTINCT "Time period" FROM Stories WHERE "Time period" IS NOT NULL AND "Time period" != ''`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query time period: %w", err)
	}
	defer rows.Close()

	timePeriodMap := make(map[string]data.TimePeriod)

	for rows.Next() {
		var timePeriodJSON sql.NullString
		if err := rows.Scan(&timePeriodJSON); err != nil {
			continue
		}

		if !timePeriodJSON.Valid || timePeriodJSON.String == "" {
			continue
		}

		var timePeriodStrings []string
		if err := json.Unmarshal([]byte(timePeriodJSON.String), &timePeriodStrings); err != nil {
			continue
		}

		for _, timePeriodTitle := range timePeriodStrings {
			if timePeriodTitle != "" {
				timePeriodMap[timePeriodTitle] = data.TimePeriod{
					Title:  timePeriodTitle,
					URL:    "/timeperiod/" + util.Slugify(timePeriodTitle) + ".html",
					Colour: data.TitleToHexColor(timePeriodTitle),
				}
			}
		}
	}

	// Convert map to slice
	var timePeriod []data.TimePeriod
	for _, timePeriodObj := range timePeriodMap {
		timePeriod = append(timePeriod, timePeriodObj)
	}

	return timePeriod, nil
}

// GetStoriesForTimePeriod retrieves all stories for a given time period value from SQLite
func (s *SQLiteAdapter) GetStoriesForTimePeriod(timePeriodTitle string) ([]data.Story, error) {
	log.Println("Getting stories for time period", timePeriodTitle)

	db := s.connectToDatabase()
	defer db.Close()

	// Time period is stored as JSON array - use LIKE for matching
	likePattern := fmt.Sprintf("%%%q%%", timePeriodTitle)

	query := `
		SELECT *
		FROM Stories
		WHERE "Time period" LIKE ?;
	`

	rows, err := db.Query(query, likePattern)
	if err != nil {
		return nil, fmt.Errorf("failed to query stories for time period: %w", err)
	}
	defer rows.Close()

	return s.scanStories(rows)
}
