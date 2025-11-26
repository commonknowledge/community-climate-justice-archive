// Retrieves and processes themes as required.
package store

import (
	"log"

	"community-climate-justice-archive/data"
)

// GetStoriesForTheme retrieves all stories for a given theme from the data source.
func GetStoriesForTheme(themeTitle string) []data.Story {
	adapter := GetAdapter()
	stories, err := adapter.GetStoriesForTheme(themeTitle)
	if err != nil {
		log.Fatalf("Failed to get stories for theme: %v", err)
	}
	return stories
}

// GetThemes retrieves all themes from the database and returns them as a slice of Theme.
// Intended for passing to HTML templates.
func GetThemes() []data.Theme {
	adapter := GetAdapter()
	themes, err := adapter.GetThemes()
	if err != nil {
		log.Fatalf("Failed to get themes: %v", err)
	}
	return themes
}
