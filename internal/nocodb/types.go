// Package nocodb translates between NocoDB's format and our Go structs.
//
// NocoDB sends data back to us in JSON, but that JSON doesn't quite match the
// Story struct we use everywhere else. This file does the translation.
//
// The tricky bits:
// - NocoDB field names have spaces in them (like "Image / video / sound")
// - Tags come as arrays of strings (like ["Climate Change", "Community"])
// - Attachments have lots of metadata about file sizes, dimensions, etc.
// - Related stories need to be fetched separately
//
// What this file does:
// - NocoDBStoryDTO: A struct that matches NocoDB's field names exactly
// - Parse functions: Convert NocoDB's arrays into proper Theme, Type, Weather structs
// - Conversion logic: Turn NocoDBStoryDTO into the Story struct everyone else uses
//
// So when NocoDB sends ["Climate Change", "Community"], we turn that into a
// proper []Theme list with URLs and colours. When it sends attachment JSON, we
// parse that into StoryAttachment structs. And so on!
package nocodb

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/util"
)

const (
	// NocoDB field IDs for self-referential story connection fields.
	relationshipFieldInspiredByID  = "ccsugv6du8wnisr"
	relationshipFieldHasInspiredID = "cilfzk65ypiw6o4"

	// Cache keys used by client.fetchAndCacheRelationships.
	cachedInspiredByKey         = "__cached_inspired_by"
	cachedHasInspiredKey        = "__cached_has_inspired"
	cachedContributorsKey       = "__cached_contributors"
	cachedPublicContributorsKey = "__cached_public_contributors"
)

// NocoDBStoryDTO is what NocoDB sends us when we ask for a story.
//
// The field names match NocoDB's JSON exactly (notice they have spaces and slashes).
// The json:"..." tags tell Go which JSON field maps to which struct field.
//
// We use interface{} for most fields because NocoDB might send null, a string,
// a number, or even an array depending on the field. We handle that uncertainty
// in the conversion functions.
type NocoDBStoryDTO struct {
	ID                      interface{} `json:"Id"`
	CreatedTime             interface{} `json:"CreatedAt"`
	Finding                 interface{} `json:"Title"`
	HighStExperiment        interface{} `json:"Project / Event"`
	WhatWasIsIf             interface{} `json:"What was/is/if"`
	ImageVideoSound         interface{} `json:"Image / video / sound"`
	Location                interface{} `json:"Location"`
	StartDateTime           interface{} `json:"Dated created / experienced"`
	EndDateTime             interface{} `json:"Date added to the archive"`
	Season                  interface{} `json:"Season"`
	Weather                 interface{} `json:"Weather"`
	StreetDetectoristClue   interface{} `json:"If you would like to fill out a Street Detectorist map, you can download it here:"`
	Themes                  interface{} `json:"Themes"`
	Experience              interface{} `json:"Description"`
	TimeSpan                interface{} `json:"-"`
	InspiredBy              interface{} `json:"Inspired by"`
	HasInspired             interface{} `json:"Has inspired"`
	OtherComments           interface{} `json:"Description"`
	Type                    interface{} `json:"Type"`
	GiftedBy                interface{} `json:"-"` // ellipsis (U+2026) is not valid in Go JSON tags; read directly from record map
	Contributors            interface{} `json:"Contributors"`
	PublicContributors      interface{} `json:"Public Contributors"`
	ScalePermanence         interface{} `json:"Scale of permanence"`
	TimePeriod              interface{} `json:"Time period"`
	PersonFinder            interface{} `json:"-"`
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
	Approved                interface{} `json:"Approved"`
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

	// Convert gifted by — read directly from record because the NocoDB column name
	// "Gifted or co-created by…" contains U+2026 (ellipsis), which is not a valid
	// character in Go JSON struct tags, so dto.GiftedBy is always nil.
	giftedBy, err := ParseGiftedByFromNocoDB(record["Gifted or co-created by\u2026"])
	if err != nil {
		log.Printf("Warning: failed to parse gifted by: %v", err)
		giftedBy = []data.GiftedBy{}
	}

	// Convert contributors
	contributors, err := ParseContributorsFromNocoDB(dto.Contributors)
	if err != nil {
		log.Printf("Warning: failed to parse contributors: %v", err)
		contributors = []data.Contributor{}
	}
	if cached := fetchContributorsDirect(toString(dto.ID), cachedContributorsKey, client); len(cached) > 0 {
		contributors = cached
	}

	// Convert public contributors
	publicContributors, err := ParseContributorsFromNocoDB(dto.PublicContributors)
	if err != nil {
		log.Printf("Warning: failed to parse public contributors: %v", err)
		publicContributors = []data.Contributor{}
	}
	if cached := fetchContributorsDirect(toString(dto.ID), cachedPublicContributorsKey, client); len(cached) > 0 {
		publicContributors = cached
	}

	// Convert scale of permanence
	scalePermanence, err := ParseScalePermanenceFromNocoDB(dto.ScalePermanence)
	if err != nil {
		log.Printf("Warning: failed to parse scale of permanence: %v", err)
		scalePermanence = []data.ScalePermanence{}
	}

	// Convert what was/is/if
	whatWasIsIf, err := ParseWhatWasIsIfFromNocoDB(dto.WhatWasIsIf)
	if err != nil {
		log.Printf("Warning: failed to parse what was/is/if: %v", err)
		whatWasIsIf = []data.WhatWasIsIf{}
	}

	// Convert time period
	timePeriod, err := ParseTimePeriodFromNocoDB(dto.TimePeriod)
	if err != nil {
		log.Printf("Warning: failed to parse time period: %v", err)
		timePeriod = []data.TimePeriod{}
	}

	// Ensure any newly uploaded NocoDB attachments exist locally before the
	// image-processing step runs later in the build.
	if err := syncAttachmentsFromNocoDB(dto.ImageVideoSound); err != nil {
		log.Printf("Warning: failed to sync attachments for story %s: %v", toString(dto.ID), err)
	}

	story := data.Story{
		ID:               toString(dto.ID),
		CreatedTime:      toString(dto.CreatedTime),
		Finding:          toString(dto.Finding),
		HighStExperiment: toString(dto.HighStExperiment),
		ImageVideoSound: func() string {
			// Keep original NocoDB structure for rich metadata (mimetype, size, dimensions)
			if dto.ImageVideoSound != nil {
				if jsonBytes, err := json.Marshal(dto.ImageVideoSound); err == nil {
					return string(jsonBytes)
				}
			}
			return ""
		}(),
		Location:                toString(dto.Location),
		StartDateTime:           toString(dto.StartDateTime),
		EndDateTime:             toString(dto.EndDateTime),
		Season:                  toString(dto.Season),
		StreetDetectoristClue:   toString(dto.StreetDetectoristClue),
		Themes:                  themes,
		Experience:              toString(dto.Experience),
		TimeSpan:                toString(dto.TimeSpan),
		InspiredBy:              fetchStoryConnectionsDirect(toString(dto.ID), relationshipFieldInspiredByID, client),
		HasInspired:             fetchStoryConnectionsDirect(toString(dto.ID), relationshipFieldHasInspiredID, client),
		OtherComments:           toString(dto.OtherComments),
		Type:                    types,
		Weather:                 weather,
		Contributors:            contributors,
		PublicContributors:      publicContributors,
		GiftedBy:                giftedBy,
		ScalePermanence:         scalePermanence,
		WhatWasIsIf:             whatWasIsIf,
		TimePeriod:              timePeriod,
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
		Approved:                toString(dto.Approved),
	}

	// Set URL based on finding with ID suffix
	story.URL = createStoryURLFromFindingWithID(story.Finding, story.ID)

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

// ParseGiftedByFromNocoDB parses gifted by from NocoDB field
func ParseGiftedByFromNocoDB(giftedByField interface{}) ([]data.GiftedBy, error) {
	if giftedByField == nil {
		return []data.GiftedBy{}, nil
	}

	var giftedByStrings []string

	switch v := giftedByField.(type) {
	case string:
		if v == "" {
			return []data.GiftedBy{}, nil
		}
		// NocoDB returns gifted by as comma-separated string
		if err := json.Unmarshal([]byte(v), &giftedByStrings); err != nil {
			// If not JSON, treat as comma-separated string (NocoDB format)
			giftedByStrings = strings.Split(v, ",")
			for i, giftedByStr := range giftedByStrings {
				giftedByStrings[i] = strings.TrimSpace(giftedByStr)
			}
		}
	case []interface{}:
		for _, item := range v {
			if str := toString(item); str != "" {
				giftedByStrings = append(giftedByStrings, str)
			}
		}
	case []string:
		giftedByStrings = v
	default:
		str := toString(v)
		if str != "" {
			if err := json.Unmarshal([]byte(str), &giftedByStrings); err != nil {
				// If not JSON, treat as comma-separated string
				giftedByStrings = strings.Split(str, ",")
				for i, giftedByStr := range giftedByStrings {
					giftedByStrings[i] = strings.TrimSpace(giftedByStr)
				}
			}
		}
	}

	// Convert to GiftedBy structs
	var giftedBy []data.GiftedBy
	for _, giftedByTitle := range giftedByStrings {
		if giftedByTitle != "" {
			giftedBy = append(giftedBy, data.GiftedBy{
				Title:  giftedByTitle,
				URL:    "/giftedby/" + util.Slugify(giftedByTitle) + ".html",
				Colour: data.TitleToHexColor(giftedByTitle),
			})
		}
	}

	return giftedBy, nil
}

// ParseContributorsFromNocoDB parses linked contributor records from NocoDB fields.
func ParseContributorsFromNocoDB(contributorsField interface{}) ([]data.Contributor, error) {
	if contributorsField == nil {
		return []data.Contributor{}, nil
	}

	isLikelyCount := func(value string) bool {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return false
		}
		_, err := strconv.Atoi(trimmed)
		return err == nil
	}

	toContributor := func(raw map[string]interface{}) data.Contributor {
		name := strings.TrimSpace(toString(raw["name"]))
		if name == "" {
			name = strings.TrimSpace(toString(raw["Name"]))
		}
		if name == "" {
			name = strings.TrimSpace(toString(raw["title"]))
		}
		if name == "" {
			name = strings.TrimSpace(toString(raw["Title"]))
		}

		email := strings.TrimSpace(toString(raw["email"]))
		if email == "" {
			email = strings.TrimSpace(toString(raw["Email"]))
		}

		approved := strings.TrimSpace(toString(raw["approved"]))
		if approved == "" {
			approved = strings.TrimSpace(toString(raw["Approved"]))
		}

		return data.Contributor{
			Name:     name,
			Email:    email,
			Approved: approved,
		}
	}

	var contributors []data.Contributor

	switch v := contributorsField.(type) {
	case string:
		if v == "" {
			return []data.Contributor{}, nil
		}

		var objectArray []map[string]interface{}
		if err := json.Unmarshal([]byte(v), &objectArray); err == nil {
			for _, item := range objectArray {
				contributor := toContributor(item)
				if contributor.Name != "" {
					contributors = append(contributors, contributor)
				}
			}
			return contributors, nil
		}

		var stringArray []string
		if err := json.Unmarshal([]byte(v), &stringArray); err == nil {
			for _, name := range stringArray {
				name = strings.TrimSpace(name)
				if name != "" && !isLikelyCount(name) {
					contributors = append(contributors, data.Contributor{Name: name})
				}
			}
			return contributors, nil
		}

		for _, name := range strings.Split(v, ",") {
			name = strings.TrimSpace(name)
			if name != "" && !isLikelyCount(name) {
				contributors = append(contributors, data.Contributor{Name: name})
			}
		}

	case []interface{}:
		for _, item := range v {
			switch parsed := item.(type) {
			case map[string]interface{}:
				contributor := toContributor(parsed)
				if contributor.Name != "" {
					contributors = append(contributors, contributor)
				}
			case string:
				name := strings.TrimSpace(parsed)
				if name != "" && !isLikelyCount(name) {
					contributors = append(contributors, data.Contributor{Name: name})
				}
			}
		}

	case []string:
		for _, name := range v {
			name = strings.TrimSpace(name)
			if name != "" && !isLikelyCount(name) {
				contributors = append(contributors, data.Contributor{Name: name})
			}
		}

	default:
		str := strings.TrimSpace(toString(v))
		if str != "" && !isLikelyCount(str) {
			contributors = append(contributors, data.Contributor{Name: str})
		}
	}

	return contributors, nil
}

// ParseScalePermanenceFromNocoDB parses scale of permanence from NocoDB field
func ParseScalePermanenceFromNocoDB(scalePermanenceField interface{}) ([]data.ScalePermanence, error) {
	if scalePermanenceField == nil {
		return []data.ScalePermanence{}, nil
	}

	var scalePermanenceStrings []string

	switch v := scalePermanenceField.(type) {
	case string:
		if v == "" {
			return []data.ScalePermanence{}, nil
		}
		// NocoDB returns scale of permanence as comma-separated string
		if err := json.Unmarshal([]byte(v), &scalePermanenceStrings); err != nil {
			// If not JSON, treat as comma-separated string (NocoDB format)
			scalePermanenceStrings = strings.Split(v, ",")
			for i, scalePermanenceStr := range scalePermanenceStrings {
				scalePermanenceStrings[i] = strings.TrimSpace(scalePermanenceStr)
			}
		}
	case []interface{}:
		for _, item := range v {
			if str := toString(item); str != "" {
				scalePermanenceStrings = append(scalePermanenceStrings, str)
			}
		}
	case []string:
		scalePermanenceStrings = v
	default:
		str := toString(v)
		if str != "" {
			if err := json.Unmarshal([]byte(str), &scalePermanenceStrings); err != nil {
				// If not JSON, treat as comma-separated string
				scalePermanenceStrings = strings.Split(str, ",")
				for i, scalePermanenceStr := range scalePermanenceStrings {
					scalePermanenceStrings[i] = strings.TrimSpace(scalePermanenceStr)
				}
			}
		}
	}

	// Convert to ScalePermanence structs
	var scalePermanence []data.ScalePermanence
	for _, scalePermanenceTitle := range scalePermanenceStrings {
		if scalePermanenceTitle != "" {
			scalePermanence = append(scalePermanence, data.ScalePermanence{
				Title:  scalePermanenceTitle,
				URL:    "/scalepermanence/" + util.Slugify(scalePermanenceTitle) + ".html",
				Colour: data.TitleToHexColor(scalePermanenceTitle),
			})
		}
	}

	return scalePermanence, nil
}

// ParseWhatWasIsIfFromNocoDB parses what was/is/if from NocoDB field
func ParseWhatWasIsIfFromNocoDB(whatWasIsIfField interface{}) ([]data.WhatWasIsIf, error) {
	if whatWasIsIfField == nil {
		return []data.WhatWasIsIf{}, nil
	}

	var whatWasIsIfStrings []string

	switch v := whatWasIsIfField.(type) {
	case string:
		if v == "" {
			return []data.WhatWasIsIf{}, nil
		}
		// NocoDB returns what was/is/if as comma-separated string
		if err := json.Unmarshal([]byte(v), &whatWasIsIfStrings); err != nil {
			// If not JSON, treat as comma-separated string (NocoDB format)
			whatWasIsIfStrings = strings.Split(v, ",")
			for i, whatWasIsIfStr := range whatWasIsIfStrings {
				whatWasIsIfStrings[i] = strings.TrimSpace(whatWasIsIfStr)
			}
		}
	case []interface{}:
		for _, item := range v {
			if str := toString(item); str != "" {
				whatWasIsIfStrings = append(whatWasIsIfStrings, str)
			}
		}
	case []string:
		whatWasIsIfStrings = v
	default:
		str := toString(v)
		if str != "" {
			if err := json.Unmarshal([]byte(str), &whatWasIsIfStrings); err != nil {
				// If not JSON, treat as comma-separated string
				whatWasIsIfStrings = strings.Split(str, ",")
				for i, whatWasIsIfStr := range whatWasIsIfStrings {
					whatWasIsIfStrings[i] = strings.TrimSpace(whatWasIsIfStr)
				}
			}
		}
	}

	// Convert to WhatWasIsIf structs
	var whatWasIsIf []data.WhatWasIsIf
	for _, whatWasIsIfTitle := range whatWasIsIfStrings {
		if whatWasIsIfTitle != "" {
			whatWasIsIf = append(whatWasIsIf, data.WhatWasIsIf{
				Title:  whatWasIsIfTitle,
				URL:    "/whatwasisif/" + util.Slugify(whatWasIsIfTitle) + ".html",
				Colour: data.TitleToHexColor(whatWasIsIfTitle),
			})
		}
	}

	return whatWasIsIf, nil
}

// ParseTimePeriodFromNocoDB parses time period from NocoDB field
func ParseTimePeriodFromNocoDB(timePeriodField interface{}) ([]data.TimePeriod, error) {
	if timePeriodField == nil {
		return []data.TimePeriod{}, nil
	}

	var timePeriodStrings []string

	switch v := timePeriodField.(type) {
	case string:
		if v == "" {
			return []data.TimePeriod{}, nil
		}
		// NocoDB returns time period as comma-separated string
		if err := json.Unmarshal([]byte(v), &timePeriodStrings); err != nil {
			// If not JSON, treat as comma-separated string (NocoDB format)
			timePeriodStrings = strings.Split(v, ",")
			for i, timePeriodStr := range timePeriodStrings {
				timePeriodStrings[i] = strings.TrimSpace(timePeriodStr)
			}
		}
	case []interface{}:
		for _, item := range v {
			if str := toString(item); str != "" {
				timePeriodStrings = append(timePeriodStrings, str)
			}
		}
	case []string:
		timePeriodStrings = v
	default:
		str := toString(v)
		if str != "" {
			if err := json.Unmarshal([]byte(str), &timePeriodStrings); err != nil {
				// If not JSON, treat as comma-separated string
				timePeriodStrings = strings.Split(str, ",")
				for i, timePeriodStr := range timePeriodStrings {
					timePeriodStrings[i] = strings.TrimSpace(timePeriodStr)
				}
			}
		}
	}

	// Convert to TimePeriod structs
	var timePeriod []data.TimePeriod
	for _, timePeriodTitle := range timePeriodStrings {
		if timePeriodTitle != "" {
			timePeriod = append(timePeriod, data.TimePeriod{
				Title:  timePeriodTitle,
				URL:    "/timeperiod/" + util.Slugify(timePeriodTitle) + ".html",
				Colour: data.TitleToHexColor(timePeriodTitle),
			})
		}
	}

	return timePeriod, nil
}

// ParseAttachmentsFromNocoDB parses attachments from NocoDB field and converts to expected JSON format
// Downloads missing files from NocoDB if they don't exist locally
func ParseAttachmentsFromNocoDB(attachmentField interface{}) (string, error) {
	if attachmentField == nil {
		return "", nil
	}

	switch v := attachmentField.(type) {
	case string:
		// If it's already a JSON string, return as-is
		if strings.TrimSpace(v) == "" {
			return "", nil
		}
		if strings.HasPrefix(strings.TrimSpace(v), "[") {
			return v, nil
		}
		// Single filename string - convert to expected JSON format
		return convertSingleFilenameToJSON(v), nil
	case []interface{}:
		// NocoDB returns attachments as array of objects with metadata
		// Convert to the JSON format expected by GetStoryAttachments()
		var storyAttachments []map[string]interface{}

		for _, item := range v {
			if attachmentObj, ok := item.(map[string]interface{}); ok {
				filename := toString(firstNonNil(attachmentObj["title"], attachmentObj["filename"]))
				downloadPath := toString(firstNonNil(attachmentObj["path"], attachmentObj["url"]))
				mimetype := toString(firstNonNil(attachmentObj["mimetype"], attachmentObj["type"]))

				if filename != "" {
					localFilePath := localAttachmentPath(filename, mimetype)
					if downloadPath != "" && localFilePath != "" && !fileExists(localFilePath) {
						err := downloadFileFromNocoDB(downloadPath, localFilePath)
						if err != nil {
							log.Printf("Warning: failed to download file %s: %v", filename, err)
							// Continue anyway, file might be available elsewhere
						}
					}

					storyAttachment := map[string]interface{}{
						"title":    filename,
						"filename": filename,
						"url":      "", // Will be set by GetStoryAttachments()
						"path":     downloadPath,
						"mimetype": mimetype,
						"type":     mimetype,
						"size":     firstNonNil(attachmentObj["size"], 0),
						"width":    firstNonNil(attachmentObj["width"], 0),
						"height":   firstNonNil(attachmentObj["height"], 0),
					}
					storyAttachments = append(storyAttachments, storyAttachment)
				}
			}
		}

		if len(storyAttachments) > 0 {
			jsonBytes, err := json.Marshal(storyAttachments)
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

// syncAttachmentsFromNocoDB downloads any missing NocoDB-hosted attachments into
// the local asset folders so later build steps can process and copy them.
func syncAttachmentsFromNocoDB(attachmentField interface{}) error {
	attachments, err := attachmentObjectsFromField(attachmentField)
	if err != nil {
		return err
	}

	for _, attachmentObj := range attachments {
		filename := toString(firstNonNil(attachmentObj["title"], attachmentObj["filename"]))
		downloadPath := toString(firstNonNil(attachmentObj["path"], attachmentObj["url"]))
		mimetype := toString(firstNonNil(attachmentObj["mimetype"], attachmentObj["type"]))

		if filename == "" || downloadPath == "" {
			continue
		}

		localFilePath := localAttachmentPath(filename, mimetype)
		if localFilePath == "" || fileExists(localFilePath) {
			continue
		}

		if err := downloadFileFromNocoDB(downloadPath, localFilePath); err != nil {
			return fmt.Errorf("failed to download %s to %s: %w", filename, localFilePath, err)
		}
	}

	return nil
}

func attachmentObjectsFromField(attachmentField interface{}) ([]map[string]interface{}, error) {
	if attachmentField == nil {
		return nil, nil
	}

	switch v := attachmentField.(type) {
	case []map[string]interface{}:
		return v, nil
	case []interface{}:
		attachments := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			if attachmentObj, ok := item.(map[string]interface{}); ok {
				attachments = append(attachments, attachmentObj)
			}
		}
		return attachments, nil
	case string:
		if strings.TrimSpace(v) == "" {
			return nil, nil
		}

		var attachments []map[string]interface{}
		if err := json.Unmarshal([]byte(v), &attachments); err != nil {
			return nil, fmt.Errorf("failed to parse attachment JSON: %w", err)
		}
		return attachments, nil
	default:
		return nil, nil
	}
}

func localAttachmentPath(filename, mimetype string) string {
	// Keep downloaded source attachments together in images/ so the existing build
	// pipeline can process/copy mixed media from one place.
	switch attachmentCategoryFromMimeType(mimetype) {
	case "image", "audio", "video", "document":
		return filepath.Join("images", filename)
	default:
		return filepath.Join("images", filename)
	}
}

func attachmentCategoryFromMimeType(mimetype string) string {
	switch {
	case strings.HasPrefix(mimetype, "image/"):
		return "image"
	case strings.HasPrefix(mimetype, "audio/"):
		return "audio"
	case strings.HasPrefix(mimetype, "video/"):
		return "video"
	default:
		return "document"
	}
}

func firstNonNil(values ...interface{}) interface{} {
	for _, value := range values {
		if value == nil {
			continue
		}
		if str, ok := value.(string); ok && strings.TrimSpace(str) == "" {
			continue
		}
		return value
	}
	return nil
}

// Helper function to convert a single filename to the expected JSON format
func convertSingleFilenameToJSON(filename string) string {
	if filename == "" {
		return ""
	}

	// Use original filename since we're downloading images now
	// No need to clean NocoDB suffixes anymore

	storyImage := map[string]interface{}{
		"title":    filename,
		"filename": filename,
		"url":      "",           // Will be set by GetStoryAttachments()
		"mimetype": "image/jpeg", // Default, will be corrected later
		"type":     "image/jpeg",
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

// createStoryURLFromFindingWithID creates a URL from the story finding with ID suffix
func createStoryURLFromFindingWithID(finding, id string) string {
	if finding == "" {
		return ""
	}
	slug := util.Slugify(finding)
	return fmt.Sprintf("/stories/%s-%s.html", slug, id)
}

// Helper function to check if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// downloadFileFromNocoDB downloads a file from NocoDB to the local directory
func downloadFileFromNocoDB(downloadPath, localPath string) error {
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

	// Ensure cache is loaded once before direct lookups.
	if !client.cacheLoaded {
		_, err := client.GetAllRecords()
		if err != nil {
			log.Printf("Warning: Failed to get cached records for record %s: %v", recordID, err)
			return []data.StoryConnection{}
		}
	}

	record, found := client.getCachedRecordByID(recordID)
	if !found {
		log.Printf("Warning: Record %s not found in cache", recordID)
		return []data.StoryConnection{}
	}

	// Determine which cached field to read based on fieldID
	var cacheKey string
	if fieldID == relationshipFieldInspiredByID {
		cacheKey = cachedInspiredByKey
	} else if fieldID == relationshipFieldHasInspiredID {
		cacheKey = cachedHasInspiredKey
	} else {
		log.Printf("Warning: Unknown fieldID %s for record %s", fieldID, recordID)
		return []data.StoryConnection{}
	}

	// Get the cached relationships
	if cachedData, exists := record[cacheKey]; exists {
		// Handle nil case (no relationships)
		if cachedData == nil {
			return []data.StoryConnection{}
		}

		// Handle both fresh cache ([]data.StoryConnection) and disk-loaded cache ([]interface{})
		if connections, ok := cachedData.([]data.StoryConnection); ok {
			// Fresh cache data - use directly
			return connections
		} else if interfaceSlice, ok := cachedData.([]interface{}); ok {
			// Disk-loaded cache data - convert from generic interfaces
			return convertInterfaceSliceToConnections(interfaceSlice)
		} else {
			log.Printf("Warning: Cached relationship data has unexpected type %T for record %s", cachedData, recordID)
		}
	}

	// If no cached data found, return empty slice (this is normal for records with no relationships)
	return []data.StoryConnection{}
}

// convertInterfaceSliceToConnections converts a slice of generic interfaces (from JSON deserialization)
// back to StoryConnection objects
func convertInterfaceSliceToConnections(interfaceSlice []interface{}) []data.StoryConnection {
	var connections []data.StoryConnection

	for _, item := range interfaceSlice {
		if connMap, ok := item.(map[string]interface{}); ok {
			connection := data.StoryConnection{
				ID:                 toString(connMap["id"]),
				Title:              toString(connMap["title"]),
				Finding:            toString(connMap["finding"]),
				Image:              toString(connMap["image"]),
				ThumbURL:           toString(connMap["thumbUrl"]),
				URL:                toString(connMap["url"]),
				AttachmentType:     toString(connMap["attachmentType"]),
				AttachmentFilename: toString(connMap["attachmentFilename"]),
			}
			connections = append(connections, connection)
		}
	}

	return connections
}

// fetchContributorsDirect reads cached contributor relationship data from record cache.
func fetchContributorsDirect(recordID, cacheKey string, client *Client) []data.Contributor {
	if client == nil || recordID == "" || cacheKey == "" {
		return []data.Contributor{}
	}

	if !client.cacheLoaded {
		_, err := client.GetAllRecords()
		if err != nil {
			log.Printf("Warning: Failed to get cached records for contributor record %s: %v", recordID, err)
			return []data.Contributor{}
		}
	}

	record, found := client.getCachedRecordByID(recordID)
	if !found {
		return []data.Contributor{}
	}

	if cachedData, exists := record[cacheKey]; exists {
		if cachedData == nil {
			return []data.Contributor{}
		}

		if contributors, ok := cachedData.([]data.Contributor); ok {
			return contributors
		}

		if interfaceSlice, ok := cachedData.([]interface{}); ok {
			return convertInterfaceSliceToContributors(interfaceSlice)
		}
	}

	return []data.Contributor{}
}

func convertInterfaceSliceToContributors(interfaceSlice []interface{}) []data.Contributor {
	var contributors []data.Contributor

	for _, item := range interfaceSlice {
		if contributorMap, ok := item.(map[string]interface{}); ok {
			name := strings.TrimSpace(toString(contributorMap["name"]))
			if name == "" {
				name = strings.TrimSpace(toString(contributorMap["Name"]))
			}
			if name == "" {
				continue
			}

			email := strings.TrimSpace(toString(contributorMap["email"]))
			if email == "" {
				email = strings.TrimSpace(toString(contributorMap["Email"]))
			}

			approved := strings.TrimSpace(toString(contributorMap["approved"]))
			if approved == "" {
				approved = strings.TrimSpace(toString(contributorMap["Approved"]))
			}

			contributors = append(contributors, data.Contributor{
				Name:     name,
				Email:    email,
				Approved: approved,
			})
		}
	}

	return contributors
}
