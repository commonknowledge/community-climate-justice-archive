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
