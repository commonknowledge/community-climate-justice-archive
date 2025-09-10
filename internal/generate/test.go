// Package generate provides functionality for generating the test page
package generate

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/config"
	"community-climate-justice-archive/internal/nocodb"
	"community-climate-justice-archive/internal/store"
)

// TestPageData holds the data for the test page template
type TestPageData struct {
	// Configuration info
	DataSourceType   string
	DataSourceStatus string
	UseNocoDB        bool
	NocoDBEndpoint   string
	NocoDBTableID    string

	// Data info
	Records     []map[string]interface{}
	FieldNames  []string
	RecordCount int
	Error       string

	// Sample conversion
	SampleStory data.Story

	// Metadata
	GeneratedAt string
}

// WriteTestPage generates a test page showing raw data from the current data source
// This is primarily intended for debugging NocoDB connections, so we skip generation
// when using SQLite to avoid unnecessary overhead in normal builds
func WriteTestPage() error {
	// Skip test page generation if NocoDB is not enabled
	if !config.AppConfig.UseNocoDB {
		log.Println("Skipping test page generation (NocoDB not enabled)")
		return nil
	}

	log.Println("Generating test page...")

	pageData := &TestPageData{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05 MST"),
	}

	// Set configuration info
	if config.AppConfig.UseNocoDB {
		pageData.DataSourceType = "NocoDB API"
		pageData.UseNocoDB = true
		pageData.NocoDBEndpoint = config.AppConfig.NocoDBEndpoint
		pageData.NocoDBTableID = config.AppConfig.NocoDBTableID
	} else {
		pageData.DataSourceType = "SQLite Database"
		pageData.UseNocoDB = false
	}

	// Try to fetch raw data
	if config.AppConfig.UseNocoDB {
		err := fetchNocoDBTestData(pageData)
		if err != nil {
			pageData.Error = err.Error()
			pageData.DataSourceStatus = "error"
			log.Printf("Error fetching NocoDB test data: %v", err)
		} else {
			pageData.DataSourceStatus = "success"
		}
	} else {
		// For SQLite, we'll still try to fetch through the adapter
		err := fetchAdapterTestData(pageData)
		if err != nil {
			pageData.Error = err.Error()
			pageData.DataSourceStatus = "error"
			log.Printf("Error fetching adapter test data: %v", err)
		} else {
			pageData.DataSourceStatus = "success"
		}
	}

	// Try to get a sample story for conversion testing
	if pageData.Error == "" {
		adapter := store.GetAdapter()
		stories, err := adapter.GetAllStories()
		if err != nil {
			log.Printf("Warning: failed to get sample story: %v", err)
		} else if len(stories) > 0 {
			pageData.SampleStory = stories[0] // Get first story as sample
		}
	}

	// Parse and execute template
	tmplPath := filepath.Join("templates", "test.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse test template: %w", err)
	}

	// Create output file
	outputPath := filepath.Join("out", "test.html")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create test output file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, pageData); err != nil {
		return fmt.Errorf("failed to execute test template: %w", err)
	}

	log.Printf("Test page generated successfully at %s", outputPath)
	return nil
}

// fetchNocoDBTestData fetches raw data from NocoDB for the test page
func fetchNocoDBTestData(pageData *TestPageData) error {
	client, err := nocodb.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create NocoDB client: %w", err)
	}

	records, err := client.GetAllRecords()
	if err != nil {
		return fmt.Errorf("failed to fetch records: %w", err)
	}

	pageData.Records = records
	pageData.RecordCount = len(records)

	// Extract field names from first record
	if len(records) > 0 {
		var fieldNames []string
		for fieldName := range records[0] {
			fieldNames = append(fieldNames, fieldName)
		}
		// Sort field names for consistent display
		sort.Strings(fieldNames)
		pageData.FieldNames = fieldNames
	}

	return nil
}

// fetchAdapterTestData fetches data through the adapter interface for the test page
func fetchAdapterTestData(pageData *TestPageData) error {
	adapter := store.GetAdapter()

	stories, err := adapter.GetAllStories()
	if err != nil {
		return fmt.Errorf("failed to fetch stories through adapter: %w", err)
	}

	// Convert stories back to map format for display
	var records []map[string]interface{}
	for _, story := range stories {
		record := map[string]interface{}{
			"ID":                      story.ID,
			"CreatedTime":             story.CreatedTime,
			"Finding":                 story.Finding,
			"HighStExperiment":        story.HighStExperiment,
			"WhatWasIsIf":             story.WhatWasIsIf,
			"Image":                   story.Image,
			"SourceImage":             story.SourceImage,
			"Location":                story.Location,
			"StartDateTime":           story.StartDateTime,
			"EndDateTime":             story.EndDateTime,
			"Season":                  story.Season,
			"StreetDetectoristClue":   story.StreetDetectoristClue,
			"Experience":              story.Experience,
			"TimeSpan":                story.TimeSpan,
			"OtherComments":           story.OtherComments,
			"PersonFinder":            story.PersonFinder,
			"MapCache":                story.MapCache,
			"MapSize":                 story.MapSize,
			"Created":                 story.Created,
			"StreetDetectoristMapURL": story.StreetDetectoristMapURL,
			"OtherTheme":              story.OtherTheme,
			"OtherWeather":            story.OtherWeather,
			"ShareStatus":             story.ShareStatus,
			"PostDate":                story.PostDate,
			"TwitterText":             story.TwitterText,
			"CharacterCount":          story.CharacterCount,
			"InstaText":               story.InstaText,
			"InstaCount":              story.InstaCount,
			"InstaImage":              story.InstaImage,
			"URL":                     story.URL,
		}

		// Convert theme, type, and weather arrays to display format
		var themes []string
		for _, theme := range story.Themes {
			themes = append(themes, theme.Title)
		}
		record["Themes"] = themes

		var types []string
		for _, typeObj := range story.Type {
			types = append(types, typeObj.Title)
		}
		record["Type"] = types

		var weather []string
		for _, weatherObj := range story.Weather {
			weather = append(weather, weatherObj.Title)
		}
		record["Weather"] = weather

		records = append(records, record)
	}

	pageData.Records = records
	pageData.RecordCount = len(records)

	// Extract field names
	if len(records) > 0 {
		var fieldNames []string
		for fieldName := range records[0] {
			fieldNames = append(fieldNames, fieldName)
		}
		sort.Strings(fieldNames)
		pageData.FieldNames = fieldNames
	}

	return nil
}
