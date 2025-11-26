// Package store has functions for getting weather data.
//
// Weather is one of the ways stories are tagged - was it sunny when this story
// happened? Rainy? Cloudy? It's a lovely atmospheric dimension to browse by, and
// each story can have multiple weather conditions.
//
// The weather data is retrieved from the configured data adapter (currently NocoDB),
// which handles the actual database interactions. This package provides a simple
// interface for the rest of the application to access weather data without needing
// to know about the underlying database implementation.
package store

import (
	"log"

	"community-climate-justice-archive/data"
)

// GetStoriesForWeather retrieves all stories tagged with a specific weather condition.
//
// This function finds all stories where the weather field includes the given weather
// condition. For example, calling GetStoriesForWeather("Sunny") returns all stories
// that were created or experienced during sunny weather.
//
// Parameters:
// - weatherTitle: The name of the weather condition (e.g., "Sunny", "Rainy", "Cloudy")
//
// Returns:
// A slice of Story structs matching the weather condition. If no stories match,
// returns an empty slice. The function will terminate the program if the database
// query fails (this is intentional - data access errors are considered fatal).
//
// Usage:
// sunnyStories := GetStoriesForWeather("Sunny")
func GetStoriesForWeather(weatherTitle string) []data.Story {
	adapter := GetAdapter()
	stories, err := adapter.GetStoriesForWeather(weatherTitle)
	if err != nil {
		log.Fatalf("Failed to get stories for weather: %v", err)
	}
	return stories
}

// GetWeather retrieves all unique weather conditions from the database.
//
// This function fetches all distinct weather conditions that appear in the archive.
// Each weather condition is returned with its title, URL (link to its index page),
// and a deterministic color for visual consistency in the interface.
//
// The weather conditions are collected from all stories - if multiple stories have
// the same weather condition, it only appears once in the results.
//
// Returns:
// A slice of Weather structs, where each struct contains:
// - Title: The weather condition name (e.g., "Sunny", "Rainy")
// - URL: Link to the index page for this weather condition
// - Colour: A hex colour code generated deterministically from the title
//
// This function is typically used when generating navigation menus or filter
// options, where you want to show all possible weather conditions that visitors
// can browse by.
func GetWeather() []data.Weather {
	adapter := GetAdapter()
	weather, err := adapter.GetWeather()
	if err != nil {
		log.Fatalf("Failed to get weather: %v", err)
	}
	return weather
}
