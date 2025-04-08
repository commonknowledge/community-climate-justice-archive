package data

import (
	"database/sql"
	"encoding/json"
	"log"
)

type Page struct {
	Title       string
	Description string
	Themes      []Theme
	Types       []Type
	Stories     []Story
}

type TaxonomyIndexPage struct {
	Title       string
	Description string
	Stories     []Story
}

type Theme struct {
	Title string
	URL   string
}

type Type struct {
	Title string
	URL   string
}

type StoryImage struct {
	Filename        string
	AlternativeText string
	Type            string
	Size            int
	Width           int
	Height          int
	URL             string
	Thumbnails      map[string]Thumbnail
}

type Thumbnail struct {
	URL    string
	Width  int
	Height int
}

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
	Weather                 string
	StreetDetectoristClue   string
	Themes                  string
	Experience              string
	TimeSpan                string
	OtherComments           string
	Type                    string
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
}

// GetStoryImage returns the image for a story.
func (s *Story) GetStoryImages() []StoryImage {
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

	var images []StoryImage
	json.Unmarshal([]byte(s.Image), &images)

	for i := range images {
		images[i].URL = "/images/" + images[i].Filename
	}

	return images
}

func (s *Story) GetStoryImage() StoryImage {
	images := s.GetStoryImages()
	if len(images) > 0 {
		log.Println("Found", len(images), "images for story", images[0].URL)
		return images[0]
	}
	return StoryImage{}
}

// StoryDTO is a data transfer object that handles NULL values from the database.
type StoryDTO struct {
	ID                      sql.NullString
	CreatedTime             sql.NullString
	Finding                 sql.NullString
	HighStExperiment        sql.NullString
	WhatWasIsIf             sql.NullString
	Image                   sql.NullString
	SourceImage             sql.NullString
	Location                sql.NullString
	StartDateTime           sql.NullString
	EndDateTime             sql.NullString
	Season                  sql.NullString
	Weather                 sql.NullString
	StreetDetectoristClue   sql.NullString
	Themes                  sql.NullString
	Experience              sql.NullString
	TimeSpan                sql.NullString
	OtherComments           sql.NullString
	Type                    sql.NullString
	PersonFinder            sql.NullString
	MapCache                sql.NullString
	MapSize                 sql.NullString
	Created                 sql.NullString
	StreetDetectoristMapURL sql.NullString
	OtherTheme              sql.NullString
	OtherWeather            sql.NullString
	ShareStatus             sql.NullString
	PostDate                sql.NullString
	TwitterText             sql.NullString
	CharacterCount          sql.NullString
	InstaText               sql.NullString
	InstaCount              sql.NullString
	InstaImage              sql.NullString
}

// ToStory converts the DTO to a Story so we can use it in our own code.
func (dto *StoryDTO) ToStory() Story {
	return Story{
		ID:                      dto.ID.String,
		CreatedTime:             dto.CreatedTime.String,
		Finding:                 dto.Finding.String,
		HighStExperiment:        dto.HighStExperiment.String,
		WhatWasIsIf:             dto.WhatWasIsIf.String,
		Image:                   dto.Image.String,
		SourceImage:             dto.SourceImage.String,
		Location:                dto.Location.String,
		StartDateTime:           dto.StartDateTime.String,
		EndDateTime:             dto.EndDateTime.String,
		Season:                  dto.Season.String,
		Weather:                 dto.Weather.String,
		StreetDetectoristClue:   dto.StreetDetectoristClue.String,
		Themes:                  dto.Themes.String,
		Experience:              dto.Experience.String,
		TimeSpan:                dto.TimeSpan.String,
		OtherComments:           dto.OtherComments.String,
		Type:                    dto.Type.String,
		PersonFinder:            dto.PersonFinder.String,
		MapCache:                dto.MapCache.String,
		MapSize:                 dto.MapSize.String,
		Created:                 dto.Created.String,
		StreetDetectoristMapURL: dto.StreetDetectoristMapURL.String,
		OtherTheme:              dto.OtherTheme.String,
		OtherWeather:            dto.OtherWeather.String,
		ShareStatus:             dto.ShareStatus.String,
		PostDate:                dto.PostDate.String,
		TwitterText:             dto.TwitterText.String,
		CharacterCount:          dto.CharacterCount.String,
		InstaText:               dto.InstaText.String,
		InstaCount:              dto.InstaCount.String,
		InstaImage:              dto.InstaImage.String,
	}
}
