// Package store has functions for getting theme data.
//
// Themes are one of the main ways people browse stories - things like "Climate Change",
// "Community", "Nature". Each story can have multiple themes, so browsing by theme
// is a nice way to find related stories.
//
// The functions here just ask the adapter for theme data, and the adapter handles
// talking to the actual database (NocoDB at the moment).
package store

import (
	"log"

	"community-climate-justice-archive/data"
)

// GetStoriesForTheme retrieves all stories tagged with a specific theme.
//
// This function finds all stories where the themes field includes the given theme.
// For example, calling GetStoriesForTheme("Climate Change") returns all stories
// tagged with the "Climate Change" theme.
//
// Themes are one of the primary ways stories are categorized. A single story can
// have multiple themes, so the same story might appear when querying for different
// themes.
//
// Parameters:
// - themeTitle: The name of the theme (e.g., "Climate Change", "Community", "Nature")
//
// Returns:
// A slice of Story structs matching the theme. If no stories match, returns an
// empty slice. The function will terminate the program if the database query fails.
func GetStoriesForTheme(themeTitle string) []data.Story {
	adapter := GetAdapter()
	stories, err := adapter.GetStoriesForTheme(themeTitle)
	if err != nil {
		log.Fatalf("Failed to get stories for theme: %v", err)
	}
	return stories
}

// GetThemes retrieves all unique themes from the database.
//
// This function fetches all distinct themes that appear in the archive. Each theme
// is returned with its title, URL (link to its index page), and a deterministic
// color for visual consistency in the interface.
//
// Themes are collected from all stories - if multiple stories share the same theme,
// it only appears once in the results.
//
// Returns:
// A slice of Theme structs, where each struct contains:
// - Title: The theme name (e.g., "Climate Change", "Community")
// - URL: Link to the index page for this theme
// - Colour: A hex colour code generated deterministically from the title
//
// This function is typically used when generating navigation menus, filter options,
// or theme index pages, where you want to show all available themes for browsing.
func GetThemes() []data.Theme {
	adapter := GetAdapter()
	themes, err := adapter.GetThemes()
	if err != nil {
		log.Fatalf("Failed to get themes: %v", err)
	}
	return themes
}
