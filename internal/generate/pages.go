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

	// Only give the template the first 40 stories
	stories = stories[:40]

	// Shuffle the stories
	rand.Shuffle(len(stories), func(i, j int) {
		stories[i], stories[j] = stories[j], stories[i]
	})

	// Select a random story for the initial random link
	randomStory := stories[rand.Intn(len(stories))]

	// Convert stories to JSON
	storiesJSON, err := convertStoriesToJSON(stories)
	if err != nil {
		return err
	}

	page := data.Page{
		Title:          "Wander – Dudley People's School for Climate Justice",
		Description:    "Wander through a random selection of stories from the Dudley Climate Justice Archive",
		Themes:         themes,
		Types:          types,
		Stories:        stories,
		RandomStoryURL: randomStory.URL,
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
	allStories := getAllStories()

	// Select a random story for the initial random link
	randomStory := allStories[rand.Intn(len(allStories))]

	// Shuffle all stories and take first 40 for initial display
	shuffledStories := make([]data.Story, len(allStories))
	copy(shuffledStories, allStories)
	rand.Shuffle(len(shuffledStories), func(i, j int) {
		shuffledStories[i], shuffledStories[j] = shuffledStories[j], shuffledStories[i]
	})

	// Take first 40 stories for initial display
	stories := shuffledStories
	if len(stories) > 40 {
		stories = stories[:40]
	}

	// Convert stories to JSON (only the 40 displayed ones)
	storiesJSON, err := convertStoriesToJSON(stories)
	if err != nil {
		return err
	}

	page := data.Page{
		Title:          "Archive – Dudley People's School for Climate Justice",
		Description:    "Explore the complete Dudley Climate Justice Archive with filters for themes, types, and weather",
		Themes:         themes,
		Types:          types,
		Stories:        stories, // Only 40 random stories for initial display
		RandomStoryURL: randomStory.URL,
		StoriesJSON:    storiesJSON,
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

	// Get connected stories for the connections view (max 20)
	connectedStories := store.GetStoriesWithConnections(20)

	// Only give the template the first 40 stories for initial display
	stories := allStories[:40]

	// Shuffle the stories
	rand.Shuffle(len(stories), func(i, j int) {
		stories[i], stories[j] = stories[j], stories[i]
	})

	// Select a random story for the initial random link
	randomStory := stories[rand.Intn(len(stories))]

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
		RandomStoryURL:   randomStory.URL,
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
