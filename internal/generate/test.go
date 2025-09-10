// Package generate provides functionality for generating the test page
package generate

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"time"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/config"
	"community-climate-justice-archive/internal/store"
)

// NocoDBTestPageData holds the data for the NocoDB test page template
type NocoDBTestPageData struct {
	// Configuration info
	NocoDBEndpoint string
	NocoDBTableID  string

	// Data info
	Stories    []data.Story
	StoryCount int

	// Metadata
	GeneratedAt string
}

// WriteNocoDBTestPage generates a test page showing data retrieved from NocoDB.
// This is intended for debugging NocoDB connections and testing the main story retrieval functionality.
func WriteNocoDBTestPage() error {
	// Skip test page generation if NocoDB is not enabled
	if !config.AppConfig.UseNocoDB {
		log.Println("Skipping NocoDB test page generation (NocoDB not enabled)")
		return nil
	}

	log.Println("Generating NocoDB test page...")

	pageData := &NocoDBTestPageData{
		GeneratedAt:    time.Now().Format("2006-01-02 15:04:05 MST"),
		NocoDBEndpoint: config.AppConfig.NocoDBEndpoint,
		NocoDBTableID:  config.AppConfig.NocoDBTableID,
	}

	// Test the main story retrieval functionality that we're trying to verify works
	log.Println("Testing story retrieval through GetAllStories()...")
	stories := store.GetAllStories()

	log.Printf("Successfully retrieved %d stories from NocoDB", len(stories))

	pageData.Stories = stories
	pageData.StoryCount = len(stories)

	// Parse and execute template
	tmplPath := filepath.Join("templates", "nocodb-test.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse NocoDB test template: %w", err)
	}

	// Create output file
	outputPath := filepath.Join("out", "nocodb-test.html")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create NocoDB test output file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, pageData); err != nil {
		return fmt.Errorf("failed to execute NocoDB test template: %w", err)
	}

	log.Printf("NocoDB test page generated successfully at %s", outputPath)
	return nil
}
