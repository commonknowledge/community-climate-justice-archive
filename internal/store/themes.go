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

// uniqueThemes returns a slice of unique themes.
func uniqueThemes(themes []data.Theme) []data.Theme {
	seen := make(map[string]bool)
	unique := []data.Theme{}

	// Loop over the slice and only keep first occurrence of each theme.
	for _, t := range themes {
		if !seen[t.Title] {
			seen[t.Title] = true
			unique = append(unique, t)
		}
	}

	return unique
}
