// Package nocodb provides types for handling NocoDB API responses
package nocodb

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/util"
)

// NocoDBStoryDTO represents a story record from NocoDB API
type NocoDBStoryDTO struct {
	ID                      interface{} `json:"Id"`
	CreatedTime             interface{} `json:"CreatedAt"`
	Finding                 interface{} `json:"Title"`
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
}

// ToStory converts a NocoDB record map to a Story struct
func NocoDBRecordToStory(record map[string]interface{}) (data.Story, error) {
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
		InspiredBy:              toString(dto.InspiredBy),
		HasInspired:             toString(dto.HasInspired),
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

// ParseImagesFromNocoDB parses images from NocoDB field (complex JSON objects)
func ParseImagesFromNocoDB(imageField interface{}) (string, error) {
	if imageField == nil {
		return "", nil
	}

	switch v := imageField.(type) {
	case string:
		// Already a string, return as-is
		return v, nil
	case []interface{}:
		// NocoDB returns images as array of objects with metadata
		// For now, we'll extract the first image's path or signedPath
		if len(v) > 0 {
			if imageObj, ok := v[0].(map[string]interface{}); ok {
				// Try to get the path or signedPath
				if path, exists := imageObj["path"]; exists {
					return toString(path), nil
				}
				if signedPath, exists := imageObj["signedPath"]; exists {
					return toString(signedPath), nil
				}
				if title, exists := imageObj["title"]; exists {
					return toString(title), nil
				}
			}
		}
		// If we can't extract individual image, return JSON representation
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", nil
		}
		return string(jsonBytes), nil
	default:
		// Try to convert to string
		return toString(v), nil
	}
}

// toString safely converts an interface{} to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case *string:
		if val == nil {
			return ""
		}
		return *val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// createStoryURLFromFinding creates a URL from the story finding (same logic as store package)
func createStoryURLFromFinding(finding string) string {
	if finding == "" {
		return ""
	}
	slug := util.Slugify(finding)
	return fmt.Sprintf("/stories/%s.html", slug)
}
