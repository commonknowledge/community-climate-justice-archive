// Retrieves and processes weather from the database as required.
package store

import (
	"log"

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
