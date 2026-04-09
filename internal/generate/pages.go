package generate

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/store"
)

// WriteWanderPage generates the wander page HTML file and writes it to the out/ directory.
func WriteWanderPage() error {
	log.Println("Starting wander page generation")
	themes := store.GetThemes()
	types := store.GetTypes()
	stories := getAllStories()

	// Only give the template a small initial slice of stories
	stories = limitStories(stories, initialStoriesDisplayCount)

	// Shuffle the stories
	rand.Shuffle(len(stories), func(i, j int) {
		stories[i], stories[j] = stories[j], stories[i]
	})

	// Convert stories to JSON
	storiesJSON, err := convertStoriesToJSON(stories)
	if err != nil {
		return err
	}

	page := data.Page{
		Title:          "Wander – Dudley Time Portal",
		Description:    "Wander through a random selection of stories from the Dudley Time Portal",
		Themes:         themes,
		Types:          types,
		Stories:        stories,
		RandomStoryURL: randomStoryURL(stories),
		StoriesJSON:    storiesJSON,
	}

	tmpl, err := loadTemplatesCached()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	file, err := os.Create("out/wander.html")
	if err != nil {
		return fmt.Errorf("failed to create wander.html: %w", err)
	}
	defer file.Close()

	err = tmpl.ExecuteTemplate(file, "wander.html", page)
	if err != nil {
		return fmt.Errorf("failed to execute wander template: %w", err)
	}

	log.Println("Wander page generated successfully")
	return nil
}

// WriteArchivePage generates the archive page HTML file and writes it to the out/ directory.
func WriteArchivePage() error {
	log.Println("Starting archive page generation")
	themes := store.GetThemes()
	types := store.GetTypes()
	weather := store.GetWeather()
	whatWasIsIf := store.GetWhatWasIsIfTypes()
	scalePermanence := store.GetScalePermanenceTypes()
	timePeriod := store.GetTimePeriodTypes()
	allStories := getAllStories()
	highStExperiments := collectHighStExperiments(allStories)

	// Shuffle all stories and take first 40 for initial display
	shuffledStories := make([]data.Story, len(allStories))
	copy(shuffledStories, allStories)
	rand.Shuffle(len(shuffledStories), func(i, j int) {
		shuffledStories[i], shuffledStories[j] = shuffledStories[j], shuffledStories[i]
	})

	// Take a small initial subset for first page render
	stories := limitStories(shuffledStories, initialStoriesDisplayCount)

	// Convert stories to JSON (only the 40 displayed ones)
	storiesJSON, err := convertStoriesToJSON(stories)
	if err != nil {
		return err
	}

	page := data.Page{
		Title:            "Archive – Dudley Time Portal",
		Description:      "Explore the complete Dudley Time Portal with filters for themes, types, and weather",
		HighStExperiment: highStExperiments,
		Themes:           themes,
		Types:            types,
		Weather:          weather,
		WhatWasIsIf:      whatWasIsIf,
		ScalePermanence:  scalePermanence,
		TimePeriod:       timePeriod,
		Stories:          stories, // Only 40 random stories for initial display
		RandomStoryURL:   randomStoryURL(allStories),
		StoriesJSON:      storiesJSON,
	}

	tmpl, err := loadTemplatesCached()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	file, err := os.Create("out/archive.html")
	if err != nil {
		return fmt.Errorf("failed to create archive.html: %w", err)
	}
	defer file.Close()

	err = tmpl.ExecuteTemplate(file, "archive.html", page)
	if err != nil {
		return fmt.Errorf("failed to execute archive template: %w", err)
	}

	log.Println("Archive page generated successfully")
	return nil
}

// WriteHomePage generates the homepage HTML file and writes it to the out/ directory.
func WriteHomePage() error {
	log.Println("Starting homepage generation")
	themes := store.GetThemes()
	types := store.GetTypes()
	allStories := getAllStories()

	// Get connected stories for the connections view
	connectedStories := store.GetStoriesWithConnections(connectedStoriesLimit)

	// Only give the template a small initial subset for first page render
	stories := limitStories(allStories, initialStoriesDisplayCount)

	// Shuffle the stories
	rand.Shuffle(len(stories), func(i, j int) {
		stories[i], stories[j] = stories[j], stories[i]
	})

	// Convert stories to JSON (keep existing functionality)
	storiesJSON, err := convertStoriesToJSON(stories)
	if err != nil {
		return err
	}

	page := data.Page{
		Title:            "Dudley People's School for Climate Justice – time portal",
		Description:      "The time portal for the Dudley People's School for Climate Justice",
		Themes:           themes,
		Types:            types,
		Stories:          stories,
		ConnectedStories: connectedStories,
		RandomStoryURL:   randomStoryURL(stories),
		StoriesJSON:      storiesJSON,
	}

	tmpl, err := loadTemplatesCached()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	fileName := "index.html"
	outputPath := filepath.Join("out", fileName)

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer file.Close()

	err = tmpl.ExecuteTemplate(file, "homepage.html", page)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	log.Printf("Successfully wrote homepage to %s", outputPath)

	return nil
}

// WriteAboutPage generates the about page and writes it to out/about.html.
func WriteAboutPage() error {
	log.Println("Starting about page generation")
	allStories := getAllStories()

	// Convert stories to JSON for the random-story button.
	storiesJSON, err := convertStoriesToJSON(allStories)
	if err != nil {
		return err
	}

	page := data.Page{
		Title:          "About the project – Dudley Time Portal",
		Description:    "Learn about the Dudley Time Portal, a community archive bringing together local stories of the past with observations of the present and imaginings of the future.",
		RandomStoryURL: randomStoryURL(allStories),
		StoriesJSON:    storiesJSON,
	}

	tmpl, err := loadTemplatesCached()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	file, err := os.Create("out/about.html")
	if err != nil {
		return fmt.Errorf("failed to create about.html: %w", err)
	}
	defer file.Close()

	if err := tmpl.ExecuteTemplate(file, "about.html", page); err != nil {
		return fmt.Errorf("failed to execute about template: %w", err)
	}

	log.Println("About page generated successfully")
	return nil
}
