// Contains the Story struct that models stories in the archive and related functions, as part of the data package.
package data

import (
	"community-climate-justice-archive/internal/util"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

type Story struct {
	ID                      string
	CreatedTime             string
	Finding                 string
	HighStExperiment        string
	WhatWasIsIf             string
	Image                   string
	SourceImage             string
	Location                string
	StartDateTime           string
	EndDateTime             string
	Season                  string
	StreetDetectoristClue   string
	Themes                  []Theme
	Experience              string
	TimeSpan                string
	OtherComments           string
	Type                    []Type
	Weather                 []Weather
	PersonFinder            string
	MapCache                string
	MapSize                 string
	Created                 string
	StreetDetectoristMapURL string
	OtherTheme              string
	OtherWeather            string
	ShareStatus             string
	PostDate                string
	TwitterText             string
	CharacterCount          string
	InstaText               string
	InstaCount              string
	InstaImage              string
	URL                     string
}

type StoryImage struct {
	Filename        string
	AlternativeText string
	Type            string
	Size            int
	Width           int
	Height          int
	URL             string
	ThumbURL        string
	MediumURL       string
	LargeURL        string
	Thumbnails      map[string]Thumbnail
}

// GetStoryImage returns the image for a story.
func (s Story) GetStoryImages() []StoryImage {
	// The image is stored as a following JSON blob format from Airtable:
	// [
	//   {
	//     "id": "attENPLcNOsEdYP5E",
	//     "width": 3252,
	//     "height": 2193,
	//     "url": "https://v5.airtableusercontent.com/...",
	//     "filename": "Michelle Gartside Model 5.png",
	//     "size": 7205950,
	//     "type": "image/png",
	//     "thumbnails": {
	//       "small": {
	//         "url": "https://v5.airtableusercontent.com/...",
	//         "width": 53,
	//         "height": 36
	//       },
	//       "large": {
	//         "url": "https://v5.airtableusercontent.com/...",
	//         "width": 759,
	//         "height": 512
	//       },
	//       "full": {
	//         "url": "https://v5.airtableusercontent.com/...",
	//         "width": 3000,
	//         "height": 3000
	//       }
	//     }
	//   }
	// ]
	//
	// In the Airtable dump this is in two columns:
	// - Image
	// - Source Image
	//
	// We need to combine them into a single slice of StoryImage structs.

	var images []StoryImage
	json.Unmarshal([]byte(s.Image), &images)

	for i := range images {
		// Split filename into name and extension
		ext := filepath.Ext(images[i].Filename)
		name := strings.TrimSuffix(images[i].Filename, ext)

		// Check if WebP version exists
		webpPath := filepath.Join("images", name+".webp")
		if _, err := os.Stat(webpPath); err == nil {
			// WebP version exists, use it
			images[i].Filename = name + ".webp"
			images[i].Type = "image/webp"
		}

		// Set URLs for different sizes
		images[i].URL = "/images/processed/" + name + ".webp"
		images[i].ThumbURL = "/images/processed/" + name + "_thumb.webp"
		images[i].MediumURL = "/images/processed/" + name + "_medium.webp"
		images[i].LargeURL = "/images/processed/" + name + "_large.webp"
	}

	json.Unmarshal([]byte(s.SourceImage), &images)

	for i := range images {
		// Split filename into name and extension
		ext := filepath.Ext(images[i].Filename)
		name := strings.TrimSuffix(images[i].Filename, ext)

		// Check if WebP version exists
		webpPath := filepath.Join("images", name+".webp")
		if _, err := os.Stat(webpPath); err == nil {
			// WebP version exists, use it
			images[i].Filename = name + ".webp"
			images[i].Type = "image/webp"
		}

		// Set URLs for different sizes
		images[i].URL = "/images/processed/" + name + ".webp"
		images[i].ThumbURL = "/images/processed/" + name + "_thumb.webp"
		images[i].MediumURL = "/images/processed/" + name + "_medium.webp"
		images[i].LargeURL = "/images/processed/" + name + "_large.webp"
	}

	return images
}

func (s Story) GetStoryImage() StoryImage {
	images := s.GetStoryImages()
	if len(images) > 0 {
		log.Println("Found", len(images), "images for story", images[0].URL)
		return images[0]
	}
	return StoryImage{}
}

// StoryDTO is a data transfer object that handles NULL values from the database.
type StoryDTO struct {
	ID                                        sql.NullInt64   // id INTEGER NO NULL
	CreatedAt                                 sql.NullString  // created_at datetime YES NULL
	UpdatedAt                                 sql.NullString  // updated_at datetime YES NULL
	CreatedBy                                 sql.NullString  // created_by varchar YES NULL
	UpdatedBy                                 sql.NullString  // updated_by varchar YES NULL
	NCOrder                                   sql.NullFloat64 // nc_order REAL YES NULL
	NCRecordID                                sql.NullString  // ncRecordId varchar YES NULL
	NCRecordHash                              sql.NullString  // ncRecordHash varchar YES NULL
	Finding                                   sql.NullString  // Finding TEXT YES NULL
	Location                                  sql.NullString  // Location TEXT YES NULL
	StartDateTime                             sql.NullString  // Start_date_and_time datetime YES NULL
	EndDateTime                               sql.NullString  // End_date_and_time datetime YES NULL
	Weather                                   sql.NullString  // Weather TEXT YES NULL
	MapCache                                  sql.NullString  // Map_Cache TEXT YES NULL
	MapSize                                   sql.NullInt64   // Map_Size INTEGER YES '0'
	Type                                      sql.NullString  // Type TEXT YES NULL
	Image                                     sql.NullString  // Image TEXT YES NULL
	SourceImage                               sql.NullString  // Source_image TEXT YES NULL
	StreetDetectoristClue                     sql.NullString  // Street_Detectorist_Clue TEXT YES NULL
	Season                                    sql.NullString  // Season TEXT YES NULL
	Themes                                    sql.NullString  // Themes TEXT YES NULL
	HighStExperiment                          sql.NullString  // High_St_Experiment TEXT YES NULL
	Experience                                sql.NullString  // Experience TEXT YES NULL
	PersonFinderImaginerStreetDetectorist     sql.NullString  // Person___Finder___Imaginer___Street_Detectorist TEXT YES NULL
	IfYouWouldLikeToFillOutAStreetDetectorist sql.NullString  // If_you_would_like_to_fill_out_a_Street_Detectorist TEXT YES NULL
	TimeSpan                                  sql.NullString  // Time_span TEXT YES NULL
	OtherTheme                                sql.NullString  // Other_theme TEXT YES NULL
	OtherWeather                              sql.NullString  // Other_weather TEXT YES NULL
	OtherCommentsSources                      sql.NullString  // Other_comments___sources TEXT YES NULL
	WhatWasIsIf                               sql.NullString  // What_was_is_if TEXT YES NULL
	ShareStatus                               sql.NullString  // Share_Status TEXT YES NULL
	PostDate                                  sql.NullString  // Post_date date YES NULL
	TwitterText                               sql.NullString  // Twitter_text TEXT YES NULL
	InstaText                                 sql.NullString  // Insta_text TEXT YES NULL
	InstaImage                                sql.NullString  // Insta_image TEXT YES NULL
	Created                                   sql.NullString  // Created datetime YES NULL
}

// ToStory converts the DTO to a Story so we can use it in our own code.
func (dto *StoryDTO) ToStory() Story {
	// Stories have themes, which are a comma-separated string in the database, that looks like this:
	// "Climate Change,Extreme Weather,Social Justice"
	// We want to convert this into a slice of Theme structs so we can use it in our templates.
	var themes []Theme

	if dto.Themes.Valid && dto.Themes.String != "" {
		themeStrings := strings.Split(dto.Themes.String, ",")
		// Convert string array to Theme structs, constructing the URL from the title.
		for _, themeTitle := range themeStrings {
			themes = append(themes, Theme{
				Title:  strings.TrimSpace(themeTitle),
				URL:    "/themes/" + util.Slugify(strings.TrimSpace(themeTitle)) + ".html",
				Colour: TitleToHexColor(strings.TrimSpace(themeTitle)),
			})
		}
	}

	// Stories have types, which are a comma-separated string in the database, that looks like this:
	// "Collage,Photograph,Poem,Text"
	// We want to convert this into a slice of Type structs so we can use it in our templates.
	var types []Type

	if dto.Type.Valid && dto.Type.String != "" {
		typeStrings := strings.Split(dto.Type.String, ",")
		// Convert string array to Type structs, constructing the URL from the title.
		for _, typeTitle := range typeStrings {
			types = append(types, Type{
				Title:  strings.TrimSpace(typeTitle),
				URL:    "/types/" + util.Slugify(strings.TrimSpace(typeTitle)) + ".html",
				Colour: TitleToHexColor(strings.TrimSpace(typeTitle)),
			})
		}
	}

	// Stories have weather, which are a comma-separated string in the database, that looks like this:
	// "Sunny,Cloudy,Rainy"
	// We want to convert this into a slice of Weather structs so we can use it in our templates.
	var weather []Weather

	if dto.Weather.Valid && dto.Weather.String != "" {
		weatherStrings := strings.Split(dto.Weather.String, ",")
		// Convert string array to Weather structs, constructing the URL from the title.
		for _, weatherTitle := range weatherStrings {
			weather = append(weather, Weather{
				Title:  strings.TrimSpace(weatherTitle),
				URL:    "/weather/" + util.Slugify(strings.TrimSpace(weatherTitle)) + ".html",
				Colour: TitleToHexColor(strings.TrimSpace(weatherTitle)),
			})
		}
	}

	idStr := ""
	if dto.ID.Valid {
		idStr = fmt.Sprintf("%d", dto.ID.Int64)
	}

	mapSizeStr := ""
	if dto.MapSize.Valid {
		mapSizeStr = fmt.Sprintf("%d", dto.MapSize.Int64)
	}

	return Story{
		ID:                    idStr,
		CreatedTime:           dto.CreatedAt.String,
		Finding:               dto.Finding.String,
		HighStExperiment:      dto.HighStExperiment.String,
		WhatWasIsIf:           dto.WhatWasIsIf.String,
		Image:                 dto.Image.String,
		SourceImage:           dto.SourceImage.String,
		Location:              dto.Location.String,
		StartDateTime:         dto.StartDateTime.String,
		EndDateTime:           dto.EndDateTime.String,
		Season:                dto.Season.String,
		Weather:               weather,
		StreetDetectoristClue: dto.StreetDetectoristClue.String,
		Themes:                themes,
		Experience:            dto.Experience.String,
		TimeSpan:              dto.TimeSpan.String,
		OtherComments:         dto.OtherCommentsSources.String,
		Type:                  types,
		PersonFinder:          dto.PersonFinderImaginerStreetDetectorist.String,
		MapCache:              dto.MapCache.String,
		MapSize:               mapSizeStr,
		Created:               dto.Created.String,
		OtherTheme:            dto.OtherTheme.String,
		OtherWeather:          dto.OtherWeather.String,
		ShareStatus:           dto.ShareStatus.String,
		PostDate:              dto.PostDate.String,
		TwitterText:           dto.TwitterText.String,
		InstaText:             dto.InstaText.String,
		InstaImage:            dto.InstaImage.String,
	}
}

func TitleToHexColor(title string) string {
	// Initialize random with title's hash for deterministic output
	titleHash := sha256.Sum256([]byte(title))
	seed := int64(titleHash[0]) | int64(titleHash[1])<<8 | int64(titleHash[2])<<16 | int64(titleHash[3])<<24
	randomGenerator := rand.New(rand.NewSource(seed))

	// Generate hue (0-360), saturation (60-100%), brightness (60-90%)
	hueValue := randomGenerator.Float64() * 360
	saturationValue := 60.0 + randomGenerator.Float64()*40.0
	brightnessValue := 60.0 + randomGenerator.Float64()*30.0

	// Convert HSB to RGB
	redValue, greenValue, blueValue := hsbToRGB(hueValue, saturationValue, brightnessValue)

	// Format as hex
	return fmt.Sprintf("#%02x%02x%02x", redValue, greenValue, blueValue)
}

// hsbToRGB converts HSB (HSV) color values to RGB
func hsbToRGB(hue, saturation, brightness float64) (uint8, uint8, uint8) {
	saturationNormalized := saturation / 100
	brightnessNormalized := brightness / 100

	chroma := brightnessNormalized * saturationNormalized
	secondComponent := chroma * (1 - math.Abs(math.Mod(hue/60, 2)-1))
	matchValue := brightnessNormalized - chroma

	var redComponent, greenComponent, blueComponent float64

	switch {
	case hue < 60:
		redComponent, greenComponent, blueComponent = chroma, secondComponent, 0
	case hue < 120:
		redComponent, greenComponent, blueComponent = secondComponent, chroma, 0
	case hue < 180:
		redComponent, greenComponent, blueComponent = 0, chroma, secondComponent
	case hue < 240:
		redComponent, greenComponent, blueComponent = 0, secondComponent, chroma
	case hue < 300:
		redComponent, greenComponent, blueComponent = secondComponent, 0, chroma
	default:
		redComponent, greenComponent, blueComponent = chroma, 0, secondComponent
	}

	finalRed := uint8((redComponent + matchValue) * 255)
	finalGreen := uint8((greenComponent + matchValue) * 255)
	finalBlue := uint8((blueComponent + matchValue) * 255)

	return finalRed, finalGreen, finalBlue
}
