// Retrieves and processes weather from the database as required.
package store

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"community-climate-justice-archive/data"
)

// GetStoriesForWeather retrieves all stories for a given weather from the data source.
func GetStoriesForWeather(weatherTitle string) []data.Story {
	adapter := GetAdapter()
	stories, err := adapter.GetStoriesForWeather(weatherTitle)
	if err != nil {
		log.Fatalf("Failed to get stories for weather: %v", err)
	}
	return stories
}

// Legacy function content moved to adapter
func GetStoriesForWeatherLegacy(weatherTitle string) []data.Story {
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
			log.Fatalf("Failed to scan story: %v", err)
		}

		story := dto.ToStory()
		story.URL = CreateStoryURLFromFindingWithID(story.Finding, story.ID)

		stories = append(stories, story)
	}

	return stories
}

// GetWeather retrieves all weather from the database and returns them as a slice of Weather.
// Intended for passing to HTML templates.
func GetWeather() []data.Weather {
	adapter := GetAdapter()
	weather, err := adapter.GetWeather()
	if err != nil {
		log.Fatalf("Failed to get weather: %v", err)
	}
	return weather
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
