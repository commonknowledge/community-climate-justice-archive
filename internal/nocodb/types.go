// Package nocodb provides types for handling NocoDB API responses
package nocodb

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/config"
	"community-climate-justice-archive/internal/util"
)

// NocoDBStoryDTO represents a story record from NocoDB API
type NocoDBStoryDTO struct {
	ID                      interface{} `json:"Id"`
	CreatedTime             interface{} `json:"CreatedAt"`
	Finding                 interface{} `json:"Title"`
	Title                   interface{} `json:"Title"`
	HighStExperiment        interface{} `json:"Project / Event"`
	WhatWasIsIf             interface{} `json:"What was/is/if"`
	Image                   interface{} `json:"Image"`
	SourceImage             interface{} `json:"Source image"`
	Location                interface{} `json:"Location"`
	StartDateTime           interface{} `json:"Dated created / experienced"`
	EndDateTime             interface{} `json:"Date added to the archive"`
	Season                  interface{} `json:"Season"`
	Weather                 interface{} `json:"Weather"`
	StreetDetectoristClue   interface{} `json:"If you would like to fill out a Street Detectorist map, you can download it here:"`
	Themes                  interface{} `json:"Themes"`
	Experience              interface{} `json:"Description"`
	TimeSpan                interface{} `json:"Scale of permanence"`
	InspiredBy              interface{} `json:"Inspired by"`
	HasInspired             interface{} `json:"Has inspired"`
	OtherComments           interface{} `json:"Description"`
	Type                    interface{} `json:"Type"`
	PersonFinder            interface{} `json:"Gifted or co-created by…"`
	MapCache                interface{} `json:"Map Cache"`
	MapSize                 interface{} `json:"Map Size"`
	Created                 interface{} `json:"CreatedAt"`
	StreetDetectoristMapURL interface{} `json:"StreetDetectoristMapURL"`
	OtherTheme              interface{} `json:"Other theme"`
	OtherWeather            interface{} `json:"Other weather"`
	ShareStatus             interface{} `json:"Share Status"`
	PostDate                interface{} `json:"Post date"`
	TwitterText             interface{} `json:"Twitter text"`
	CharacterCount          interface{} `json:"CharacterCount"`
	InstaText               interface{} `json:"Insta text"`
	InstaCount              interface{} `json:"InstaCount"`
	InstaImage              interface{} `json:"Insta image"`
	ReflectionLearning      interface{} `json:"Reflection / learning"`
	UpdatedAt               interface{} `json:"UpdatedAt"`
}

// ToStory converts a NocoDB record map to a Story struct
func NocoDBRecordToStory(record map[string]interface{}) (data.Story, error) {
	return NocoDBRecordToStoryWithClient(record, nil)
}

// NocoDBRecordToStoryWithClient converts a NocoDB record to a Story struct with client for relationship resolution
func NocoDBRecordToStoryWithClient(record map[string]interface{}, client *Client) (data.Story, error) {
	// Convert map to our DTO struct for easier handling
	dto, err := mapToNocoDBStoryDTO(record)
	if err != nil {
		return data.Story{}, fmt.Errorf("failed to convert record to DTO: %w", err)
	}

	// Convert themes
	themes, err := ParseThemesFromNocoDB(dto.Themes)
	if err != nil {
		log.Printf("Warning: failed to parse themes: %v", err)
		themes = []data.Theme{}
	}

	// Convert types
	types, err := ParseTypesFromNocoDB(dto.Type)
	if err != nil {
		log.Printf("Warning: failed to parse types: %v", err)
		types = []data.Type{}
	}

	// Convert weather
	weather, err := ParseWeatherFromNocoDB(dto.Weather)
	if err != nil {
		log.Printf("Warning: failed to parse weather: %v", err)
		weather = []data.Weather{}
	}

	story := data.Story{
		ID:                      toString(dto.ID),
		CreatedTime:             toString(dto.CreatedTime),
		Finding:                 toString(dto.Finding),
		Title:                   toString(dto.Title),
		HighStExperiment:        toString(dto.HighStExperiment),
		WhatWasIsIf:             toString(dto.WhatWasIsIf),
		Image:                   func() string { img, _ := ParseImagesFromNocoDB(dto.Image); return img }(),
		SourceImage:             func() string { img, _ := ParseImagesFromNocoDB(dto.SourceImage); return img }(),
		Location:                toString(dto.Location),
		StartDateTime:           toString(dto.StartDateTime),
		EndDateTime:             toString(dto.EndDateTime),
		Season:                  toString(dto.Season),
		StreetDetectoristClue:   toString(dto.StreetDetectoristClue),
		Themes:                  themes,
		Experience:              toString(dto.Experience),
		TimeSpan:                toString(dto.TimeSpan),
		InspiredBy:              fetchStoryConnectionsDirect(toString(dto.ID), "ccsugv6du8wnisr", client),
		HasInspired:             fetchStoryConnectionsDirect(toString(dto.ID), "cilfzk65ypiw6o4", client),
		OtherComments:           toString(dto.OtherComments),
		Type:                    types,
		Weather:                 weather,
		PersonFinder:            toString(dto.PersonFinder),
		MapCache:                toString(dto.MapCache),
		MapSize:                 toString(dto.MapSize),
		Created:                 toString(dto.Created),
		StreetDetectoristMapURL: toString(dto.StreetDetectoristMapURL),
		OtherTheme:              toString(dto.OtherTheme),
		OtherWeather:            toString(dto.OtherWeather),
		ShareStatus:             toString(dto.ShareStatus),
		PostDate:                toString(dto.PostDate),
		TwitterText:             toString(dto.TwitterText),
		CharacterCount:          toString(dto.CharacterCount),
		InstaText:               toString(dto.InstaText),
		InstaCount:              toString(dto.InstaCount),
		InstaImage:              toString(dto.InstaImage),
		ReflectionLearning:      toString(dto.ReflectionLearning),
		UpdatedAt:               toString(dto.UpdatedAt),
	}

	// Set URL based on finding (same logic as SQLite version)
	story.URL = createStoryURLFromFinding(story.Finding)

	return story, nil
}

// mapToNocoDBStoryDTO converts a generic map to our typed DTO
func mapToNocoDBStoryDTO(record map[string]interface{}) (*NocoDBStoryDTO, error) {
	// Convert map to JSON and then to our struct
	jsonData, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %w", err)
	}

	var dto NocoDBStoryDTO
	if err := json.Unmarshal(jsonData, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to DTO: %w", err)
	}

	return &dto, nil
}

// ParseThemesFromNocoDB parses themes from NocoDB field (could be JSON array or string)
func ParseThemesFromNocoDB(themesField interface{}) ([]data.Theme, error) {
	if themesField == nil {
		return []data.Theme{}, nil
	}

	var themeStrings []string

	switch v := themesField.(type) {
	case string:
		if v == "" {
			return []data.Theme{}, nil
		}
		// NocoDB returns themes as comma-separated string: "Tiny Things,Care,Control"
		// First try to parse as JSON array (for backwards compatibility)
		if err := json.Unmarshal([]byte(v), &themeStrings); err != nil {
			// If not JSON, treat as comma-separated string (NocoDB format)
			themeStrings = strings.Split(v, ",")
			// Trim whitespace from each theme
			for i, theme := range themeStrings {
				themeStrings[i] = strings.TrimSpace(theme)
			}
		}
	case []interface{}:
		// Convert interface slice to string slice
		for _, item := range v {
			if str := toString(item); str != "" {
				themeStrings = append(themeStrings, str)
			}
		}
	case []string:
		themeStrings = v
	default:
		// Try to convert to string and parse
		str := toString(v)
		if str != "" {
			if err := json.Unmarshal([]byte(str), &themeStrings); err != nil {
				// If not JSON, treat as comma-separated string
				themeStrings = strings.Split(str, ",")
				for i, theme := range themeStrings {
					themeStrings[i] = strings.TrimSpace(theme)
				}
			}
		}
	}

	// Convert to Theme structs
	var themes []data.Theme
	for _, themeTitle := range themeStrings {
		if themeTitle != "" {
			themes = append(themes, data.Theme{
				Title:  themeTitle,
				URL:    "/themes/" + util.Slugify(themeTitle) + ".html",
				Colour: data.TitleToHexColor(themeTitle),
			})
		}
	}

	return themes, nil
}

// ParseTypesFromNocoDB parses types from NocoDB field
func ParseTypesFromNocoDB(typesField interface{}) ([]data.Type, error) {
	if typesField == nil {
		return []data.Type{}, nil
	}

	var typeStrings []string

	switch v := typesField.(type) {
	case string:
		if v == "" {
			return []data.Type{}, nil
		}
		// NocoDB returns types as comma-separated string
		if err := json.Unmarshal([]byte(v), &typeStrings); err != nil {
			// If not JSON, treat as comma-separated string (NocoDB format)
			typeStrings = strings.Split(v, ",")
			for i, typeStr := range typeStrings {
				typeStrings[i] = strings.TrimSpace(typeStr)
			}
		}
	case []interface{}:
		for _, item := range v {
			if str := toString(item); str != "" {
				typeStrings = append(typeStrings, str)
			}
		}
	case []string:
		typeStrings = v
	default:
		str := toString(v)
		if str != "" {
			if err := json.Unmarshal([]byte(str), &typeStrings); err != nil {
				// If not JSON, treat as comma-separated string
				typeStrings = strings.Split(str, ",")
				for i, typeStr := range typeStrings {
					typeStrings[i] = strings.TrimSpace(typeStr)
				}
			}
		}
	}

	// Convert to Type structs
	var types []data.Type
	for _, typeTitle := range typeStrings {
		if typeTitle != "" {
			types = append(types, data.Type{
				Title:  typeTitle,
				URL:    "/types/" + util.Slugify(typeTitle) + ".html",
				Colour: data.TitleToHexColor(typeTitle),
			})
		}
	}

	return types, nil
}

// ParseWeatherFromNocoDB parses weather from NocoDB field
func ParseWeatherFromNocoDB(weatherField interface{}) ([]data.Weather, error) {
	if weatherField == nil {
		return []data.Weather{}, nil
	}

	var weatherStrings []string

	switch v := weatherField.(type) {
	case string:
		if v == "" {
			return []data.Weather{}, nil
		}
		// NocoDB returns weather as comma-separated string
		if err := json.Unmarshal([]byte(v), &weatherStrings); err != nil {
			// If not JSON, treat as comma-separated string (NocoDB format)
			weatherStrings = strings.Split(v, ",")
			for i, weatherStr := range weatherStrings {
				weatherStrings[i] = strings.TrimSpace(weatherStr)
			}
		}
	case []interface{}:
		for _, item := range v {
			if str := toString(item); str != "" {
				weatherStrings = append(weatherStrings, str)
			}
		}
	case []string:
		weatherStrings = v
	default:
		str := toString(v)
		if str != "" {
			if err := json.Unmarshal([]byte(str), &weatherStrings); err != nil {
				// If not JSON, treat as comma-separated string
				weatherStrings = strings.Split(str, ",")
				for i, weatherStr := range weatherStrings {
					weatherStrings[i] = strings.TrimSpace(weatherStr)
				}
			}
		}
	}

	// Convert to Weather structs
	var weather []data.Weather
	for _, weatherTitle := range weatherStrings {
		if weatherTitle != "" {
			weather = append(weather, data.Weather{
				Title:  weatherTitle,
				URL:    "/weather/" + util.Slugify(weatherTitle) + ".html",
				Colour: data.TitleToHexColor(weatherTitle),
			})
		}
	}

	return weather, nil
}

// ParseImagesFromNocoDB parses images from NocoDB field and converts to expected JSON format
// Downloads missing images from NocoDB if they don't exist locally
func ParseImagesFromNocoDB(imageField interface{}) (string, error) {
	if imageField == nil {
		return "", nil
	}

	switch v := imageField.(type) {
	case string:
		// If it's already a JSON string (from SQLite), return as-is
		if strings.TrimSpace(v) == "" {
			return "", nil
		}
		if strings.HasPrefix(strings.TrimSpace(v), "[") {
			return v, nil
		}
		// Single filename string - convert to expected JSON format
		return convertSingleFilenameToJSON(v), nil
	case []interface{}:
		// NocoDB returns images as array of objects with metadata
		// Convert to the JSON format expected by GetStoryImages()
		var storyImages []map[string]interface{}

		for _, item := range v {
			if imageObj, ok := item.(map[string]interface{}); ok {
				// Extract filename and download path from NocoDB object
				var filename string
				var downloadPath string

				if title, exists := imageObj["title"]; exists {
					filename = toString(title)
				}
				if path, exists := imageObj["path"]; exists {
					downloadPath = toString(path)
				}

				if filename != "" {
					// Use original filename since we're downloading images now
					// No need to clean NocoDB suffixes anymore

					// Check if we need to download the image
					localImagePath := filepath.Join("images", filename)
					if config.AppConfig.UseNocoDB && downloadPath != "" && !fileExists(localImagePath) {
						err := downloadImageFromNocoDB(downloadPath, localImagePath)
						if err != nil {
							log.Printf("Warning: failed to download image %s: %v", filename, err)
							// Continue anyway, image might be available elsewhere
						}
					}

					// Create StoryImage-compatible object
					storyImage := map[string]interface{}{
						"filename": filename,
						"url":      "",           // Will be set by GetStoryImages()
						"type":     "image/jpeg", // Default, will be corrected later
						"size":     0,
						"width":    0,
						"height":   0,
					}
					storyImages = append(storyImages, storyImage)
				}
			}
		}

		if len(storyImages) > 0 {
			jsonBytes, err := json.Marshal(storyImages)
			if err != nil {
				return "", err
			}
			return string(jsonBytes), nil
		}
		return "", nil
	default:
		// Try to convert to string and handle
		str := toString(v)
		if str == "" {
			return "", nil
		}
		return convertSingleFilenameToJSON(str), nil
	}
}

// Helper function to convert a single filename to the expected JSON format
func convertSingleFilenameToJSON(filename string) string {
	if filename == "" {
		return ""
	}

	// Use original filename since we're downloading images now
	// No need to clean NocoDB suffixes anymore

	storyImage := map[string]interface{}{
		"filename": filename,
		"url":      "",           // Will be set by GetStoryImages()
		"type":     "image/jpeg", // Default, will be corrected later
		"size":     0,
		"width":    0,
		"height":   0,
	}

	jsonBytes, _ := json.Marshal([]map[string]interface{}{storyImage})
	return string(jsonBytes)
}

// Helper function to extract filename from a file path
func extractFilenameFromPath(path string) string {
	if path == "" {
		return ""
	}

	// Extract filename from path like "download/2025/06/02/abc123/filename.jpg"
	parts := strings.Split(path, "/")
	filename := path
	if len(parts) > 0 {
		filename = parts[len(parts)-1]
	}

	// Return original filename since we're downloading images now
	// No need to clean NocoDB suffixes anymore
	return filename
}

// createStoryURLFromFinding creates a URL from the story finding (same logic as store package)
func createStoryURLFromFinding(finding string) string {
	if finding == "" {
		return ""
	}
	slug := util.Slugify(finding)
	return fmt.Sprintf("/stories/%s.html", slug)
}

// Helper function to check if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// downloadImageFromNocoDB downloads an image from NocoDB to the local images directory
func downloadImageFromNocoDB(downloadPath, localPath string) error {
	// We need access to a NocoDB client to download
	// For now, create a temporary client - in a real implementation you might pass this as a parameter
	client, err := NewClient()
	if err != nil {
		return fmt.Errorf("failed to create NocoDB client: %w", err)
	}

	return client.DownloadAttachment(downloadPath, localPath)
}

// fetchStoryConnectionsDirect reads relationship data from cache instead of making API calls
func fetchStoryConnectionsDirect(recordID, fieldID string, client *Client) []data.StoryConnection {
	if client == nil || recordID == "" || fieldID == "" {
		return []data.StoryConnection{}
	}

	// Get all cached records and find the one matching our recordID
	allRecords, err := client.GetAllRecords()
	if err != nil {
		log.Printf("Warning: Failed to get cached records for record %s: %v", recordID, err)
		return []data.StoryConnection{}
	}

	// Find the record with matching ID
	for _, record := range allRecords {
		if toString(record["Id"]) == recordID {
			// Determine which cached field to read based on fieldID
			var cacheKey string
			if fieldID == "ccsugv6du8wnisr" { // Inspired by
				cacheKey = "__cached_inspired_by"
			} else if fieldID == "cilfzk65ypiw6o4" { // Has inspired
				cacheKey = "__cached_has_inspired"
			} else {
				log.Printf("Warning: Unknown fieldID %s for record %s", fieldID, recordID)
				return []data.StoryConnection{}
			}

			// Get the cached relationships
			if cachedData, exists := record[cacheKey]; exists {
				if connections, ok := cachedData.([]data.StoryConnection); ok {
					return connections
				} else {
					log.Printf("Warning: Cached relationship data has wrong type for record %s", recordID)
				}
			}

			// If no cached data found, return empty slice (this is normal for records with no relationships)
			return []data.StoryConnection{}
		}
	}

	log.Printf("Warning: Record %s not found in cache", recordID)
	return []data.StoryConnection{}
}
