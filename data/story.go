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

type StoryConnection struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Finding            string `json:"finding"`
	Image              string `json:"image"`
	ThumbURL           string `json:"thumbUrl"`
	URL                string `json:"url"`
	AttachmentType     string `json:"attachmentType"`     // "image", "audio", "document", or "none"
	AttachmentFilename string `json:"attachmentFilename"` // filename for display with non-image attachments
}

type GiftedBy struct {
	Title  string
	URL    string
	Colour string
}

type ScalePermanence struct {
	Title  string
	URL    string
	Colour string
}

type WhatWasIsIf struct {
	Title  string
	URL    string
	Colour string
}

type TimePeriod struct {
	Title  string
	URL    string
	Colour string
}

type Story struct {
	ID                      string
	CreatedTime             string
	Finding                 string
	HighStExperiment        string
	Image                   string
	ImageVideoSound         string
	SourceImage             string
	Location                string
	StartDateTime           string
	EndDateTime             string
	Season                  string
	StreetDetectoristClue   string
	Themes                  []Theme
	Experience              string
	TimeSpan                string
	InspiredBy              []StoryConnection
	HasInspired             []StoryConnection
	OtherComments           string
	Type                    []Type
	Weather                 []Weather
	GiftedBy                []GiftedBy
	ScalePermanence         []ScalePermanence
	WhatWasIsIf             []WhatWasIsIf
	TimePeriod              []TimePeriod
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
	ReflectionLearning      string
	UpdatedAt               string
	URL                     string
}

type StoryAttachment struct {
	Filename        string
	AlternativeText string
	Type            string
	FileType        string // "image", "audio", "document"
	Size            int
	Width           int
	Height          int
	URL             string
	ThumbURL        string
	MediumURL       string
	LargeURL        string
	Thumbnails      map[string]Thumbnail
}

// IsImage returns true if the attachment is an image
func (a StoryAttachment) IsImage() bool {
	return a.FileType == "image"
}

// IsAudio returns true if the attachment is audio
func (a StoryAttachment) IsAudio() bool {
	return a.FileType == "audio"
}

// IsDocument returns true if the attachment is a document
func (a StoryAttachment) IsDocument() bool {
	return a.FileType == "document"
}

// IsPlayable returns true if the attachment can be played (audio files)
func (a StoryAttachment) IsPlayable() bool {
	return a.IsAudio()
}

// IsDownloadable returns true if the attachment should be downloaded (documents)
func (a StoryAttachment) IsDownloadable() bool {
	return a.IsDocument()
}

// GetFileTypeFromExtension determines the file type based on the file extension
func GetFileTypeFromExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	// Image files
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		return "image"
	}

	// Audio files
	switch ext {
	case ".mp3", ".wav", ".ogg":
		return "audio"
	}

	// Document files
	switch ext {
	case ".pdf", ".docx", ".doc":
		return "document"
	}

	// Default to document for unknown types
	return "document"
}

// processAttachmentSlice processes a slice of attachments and sets URLs/metadata
func processAttachmentSlice(attachments []StoryAttachment) []StoryAttachment {
	for i := range attachments {
		// Determine file type based on extension
		attachments[i].FileType = GetFileTypeFromExtension(attachments[i].Filename)

		// Split filename into name and extension
		ext := filepath.Ext(attachments[i].Filename)
		name := strings.TrimSuffix(attachments[i].Filename, ext)

		// Handle different file types
		if attachments[i].IsImage() {
			// For images, check if WebP version exists
			webpPath := filepath.Join("images", name+".webp")
			if _, err := os.Stat(webpPath); err == nil {
				// WebP version exists, use it
				attachments[i].Filename = name + ".webp"
				attachments[i].Type = "image/webp"
			}

			// Set URLs for different sizes (images only)
			attachments[i].URL = "/images/processed/" + name + ".webp"
			attachments[i].ThumbURL = "/images/processed/" + name + "_thumb.webp"
			attachments[i].MediumURL = "/images/processed/" + name + "_medium.webp"
			attachments[i].LargeURL = "/images/processed/" + name + "_large.webp"
		} else {
			// For non-images (audio, documents), just set the direct URL
			attachments[i].URL = "/images/" + attachments[i].Filename
			attachments[i].ThumbURL = ""
			attachments[i].MediumURL = ""
			attachments[i].LargeURL = ""
		}
	}
	return attachments
}

// GetStoryAttachments returns all attachments (images, audio, documents) for a story.
func (s Story) GetStoryAttachments() []StoryAttachment {
	var allAttachments []StoryAttachment

	// Debug logging for specific stories
	if s.ID == "1" || s.ID == "12" {
		log.Printf("DEBUG: Story %s attachment processing:", s.ID)
		log.Printf("DEBUG: Image field length: %d", len(s.Image))
		log.Printf("DEBUG: ImageVideoSound field length: %d", len(s.ImageVideoSound))
		log.Printf("DEBUG: SourceImage field length: %d", len(s.SourceImage))
		if s.Image != "" {
			log.Printf("DEBUG: Image field content: %s", s.Image)
		}
		if s.ImageVideoSound != "" {
			log.Printf("DEBUG: ImageVideoSound field content: %s", s.ImageVideoSound)
		}
		if s.SourceImage != "" {
			log.Printf("DEBUG: SourceImage field content: %s", s.SourceImage)
		}
	}

	// Process Image field
	if s.Image != "" {
		var imageAttachments []StoryAttachment
		if err := json.Unmarshal([]byte(s.Image), &imageAttachments); err != nil {
			log.Printf("Warning: Failed to unmarshal Image field for story %s: %v", s.ID, err)
			log.Printf("Image field content: %s", s.Image)
		} else {
			if s.ID == "1" {
				log.Printf("DEBUG: Successfully parsed %d Image attachments", len(imageAttachments))
			}
			processedAttachments := processAttachmentSlice(imageAttachments)
			allAttachments = append(allAttachments, processedAttachments...)
		}
	}

	// Process ImageVideoSound field (NocoDB format)
	if s.ImageVideoSound != "" {
		if s.ID == "1" || s.ID == "12" {
			log.Printf("DEBUG: Processing ImageVideoSound field with NocoDB format")
		}

		// Parse NocoDB attachment format
		var nocoAttachments []map[string]interface{}
		if err := json.Unmarshal([]byte(s.ImageVideoSound), &nocoAttachments); err != nil {
			log.Printf("Warning: Failed to unmarshal ImageVideoSound field for story %s: %v", s.ID, err)
			log.Printf("ImageVideoSound field content: %s", s.ImageVideoSound)
		} else {
			// Convert NocoDB format to StoryAttachment format
			for _, nocoAttachment := range nocoAttachments {
				attachment := s.convertNocoDBAttachment(nocoAttachment)
				if attachment.Filename != "" {
					allAttachments = append(allAttachments, attachment)
				}
			}

			if s.ID == "1" || s.ID == "12" {
				log.Printf("DEBUG: Successfully parsed %d ImageVideoSound attachments", len(nocoAttachments))
			}
		}
	}

	// Process SourceImage field
	if s.SourceImage != "" {
		var sourceAttachments []StoryAttachment
		if err := json.Unmarshal([]byte(s.SourceImage), &sourceAttachments); err != nil {
			log.Printf("Warning: Failed to unmarshal SourceImage field for story %s: %v", s.ID, err)
			log.Printf("SourceImage field content: %s", s.SourceImage)
		} else {
			if s.ID == "1" {
				log.Printf("DEBUG: Successfully parsed %d SourceImage attachments", len(sourceAttachments))
			}
			processedAttachments := processAttachmentSlice(sourceAttachments)
			allAttachments = append(allAttachments, processedAttachments...)
		}
	}

	// Debug logging for specific stories
	if s.ID == "1" || s.ID == "12" {
		log.Printf("DEBUG: Total attachments for story %s: %d", s.ID, len(allAttachments))
	}

	// Only log when there are no attachments for debugging purposes
	if len(allAttachments) == 0 && (s.Image != "" || s.ImageVideoSound != "" || s.SourceImage != "") {
		log.Printf("Warning: Story %s has image data but no valid attachments parsed", s.ID)
	}

	return allAttachments
}

// GetStoryAttachment returns the first attachment for a story
func (s Story) GetStoryAttachment() StoryAttachment {
	attachments := s.GetStoryAttachments()
	if len(attachments) > 0 {
		log.Println("Found", len(attachments), "attachments for story", attachments[0].URL)
		return attachments[0]
	}
	return StoryAttachment{}
}

// GetStoryImage returns the first image attachment for backward compatibility
func (s Story) GetStoryImage() StoryAttachment {
	return s.GetFirstImageAttachment()
}

// GetFirstImageAttachment returns the first image attachment specifically, or empty if none
func (s Story) GetFirstImageAttachment() StoryAttachment {
	attachments := s.GetStoryAttachments()
	for _, attachment := range attachments {
		if attachment.IsImage() {
			return attachment
		}
	}
	return StoryAttachment{}
}

// GetFirstNonImageAttachment returns the first non-image attachment (audio/document), or empty if none
func (s Story) GetFirstNonImageAttachment() StoryAttachment {
	attachments := s.GetStoryAttachments()
	for _, attachment := range attachments {
		if !attachment.IsImage() {
			return attachment
		}
	}
	return StoryAttachment{}
}

// GetStoryImages returns all attachments for backward compatibility
func (s Story) GetStoryImages() []StoryAttachment {
	return s.GetStoryAttachments()
}

// StoryDTO is a data transfer object that handles NULL values from the database.
type StoryDTO struct {
	ID                      sql.NullString
	CreatedTime             sql.NullString
	Finding                 sql.NullString
	HighStExperiment        sql.NullString
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
	InspiredBy              sql.NullString
	HasInspired             sql.NullString
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
	// Stories have themes, which are a JSON array of strings in the database, that looks like this:
	// ["Climate Change", "Extreme Weather", "Social Justice"]
	// We want to convert this into a slice of Theme structs so we can use it in our templates.
	var themes []Theme

	if dto.Themes.Valid {
		var themeStrings []string
		err := json.Unmarshal([]byte(dto.Themes.String), &themeStrings)
		if err != nil {
			log.Fatalf("Failed to unmarshal themes: %v", err)
		}

		// Convert string array to Theme structs, constructing the URL from the title.
		for _, themeTitle := range themeStrings {
			themes = append(themes, Theme{
				Title:  themeTitle,
				URL:    "/themes/" + util.Slugify(themeTitle) + ".html",
				Colour: TitleToHexColor(themeTitle),
			})
		}
	}

	// Stories have types, which are a JSON array of strings in the database, that looks like this:
	// ["Collage", "Photograph", "Poem", "Text"]
	// We want to convert this into a slice of Type structs so we can use it in our templates.
	var types []Type

	if dto.Type.Valid {
		var typeStrings []string
		err := json.Unmarshal([]byte(dto.Type.String), &typeStrings)
		if err != nil {
			log.Fatalf("Failed to unmarshal types: %v", err)
		}

		// Convert string array to Type structs, constructing the URL from the title.
		for _, typeTitle := range typeStrings {
			types = append(types, Type{
				Title:  typeTitle,
				URL:    "/types/" + util.Slugify(typeTitle) + ".html",
				Colour: TitleToHexColor(typeTitle),
			})
		}
	}

	// Stories have weather, which is a JSON array of strings in the database, that looks like this:
	// ["Sunny", "Cloudy", "Rainy"]
	// We want to convert this into a slice of Weather structs so we can use it in our templates.
	var weather []Weather

	if dto.Weather.Valid {
		var weatherStrings []string
		err := json.Unmarshal([]byte(dto.Weather.String), &weatherStrings)
		if err != nil {
			log.Fatalf("Failed to unmarshal types: %v", err)
		}

		// Convert string array to Weather structs, constructing the URL from the title.
		for _, weatherTitle := range weatherStrings {
			weather = append(weather, Weather{
				Title:  weatherTitle,
				URL:    "/weather/" + util.Slugify(weatherTitle) + ".html",
				Colour: TitleToHexColor(weatherTitle),
			})
		}
	}

	return Story{
		ID:                      dto.ID.String,
		CreatedTime:             dto.CreatedTime.String,
		Finding:                 dto.Finding.String,
		HighStExperiment:        dto.HighStExperiment.String,
		Image:                   dto.Image.String,
		SourceImage:             dto.SourceImage.String,
		Location:                dto.Location.String,
		StartDateTime:           dto.StartDateTime.String,
		EndDateTime:             dto.EndDateTime.String,
		Season:                  dto.Season.String,
		Weather:                 weather,
		StreetDetectoristClue:   dto.StreetDetectoristClue.String,
		Themes:                  themes,
		Experience:              dto.Experience.String,
		TimeSpan:                dto.TimeSpan.String,
		InspiredBy:              parseStoryConnectionsFromString(dto.InspiredBy.String),
		HasInspired:             parseStoryConnectionsFromString(dto.HasInspired.String),
		OtherComments:           dto.OtherComments.String,
		Type:                    types,
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

// parseStoryConnectionsFromString converts a string to StoryConnection slice
// This is a fallback for SQLite data that doesn't have proper relationships
func parseStoryConnectionsFromString(connectionStr string) []StoryConnection {
	if connectionStr == "" {
		return []StoryConnection{}
	}

	// For SQLite, we only have the title/finding as a string
	// Create a basic connection
	connection := StoryConnection{
		Title:   connectionStr,
		Finding: connectionStr,
		URL:     "/stories/" + util.Slugify(connectionStr) + ".html",
	}

	return []StoryConnection{connection}
}

// GetNocoDBURL returns a direct link to this story in the NocoDB interface for debugging
func (s Story) GetNocoDBURL() string {
	return fmt.Sprintf("https://nocodb-r87d.onrender.com/dashboard/#/nc/pqw5yaekkqvo25h/me04vwwhvh4jbsg?rowId=%s&path=", s.ID)
}

// convertNocoDBAttachment converts a NocoDB attachment object to StoryAttachment format
func (s Story) convertNocoDBAttachment(nocoAttachment map[string]interface{}) StoryAttachment {
	// Extract basic info
	filename := ""
	if t, ok := nocoAttachment["title"].(string); ok {
		filename = t
	}

	mimetype := ""
	if m, ok := nocoAttachment["mimetype"].(string); ok {
		mimetype = m
	}

	if filename == "" {
		return StoryAttachment{} // Return empty if no filename
	}

	// Determine file type from mimetype (not just extension)
	fileType := s.getFileTypeFromMimeType(mimetype)

	attachment := StoryAttachment{
		Filename:        filename,
		AlternativeText: s.Finding + " - " + filename,
		Type:            mimetype,
		FileType:        fileType,
	}

	// Set size if available
	if size, ok := nocoAttachment["size"].(float64); ok {
		attachment.Size = int(size)
	}

	// Handle different file types appropriately
	switch fileType {
	case "image":
		// For images, set dimensions and create processed URLs
		if width, ok := nocoAttachment["width"].(float64); ok {
			attachment.Width = int(width)
		}
		if height, ok := nocoAttachment["height"].(float64); ok {
			attachment.Height = int(height)
		}

		// Create processed image URLs (webp versions)
		processedFilename := strings.TrimSuffix(filename, filepath.Ext(filename)) + ".webp"
		attachment.URL = "/images/processed/" + processedFilename
		attachment.ThumbURL = "/images/processed/" + strings.TrimSuffix(processedFilename, ".webp") + "_thumb.webp"
		attachment.MediumURL = "/images/processed/" + strings.TrimSuffix(processedFilename, ".webp") + "_medium.webp"
		attachment.LargeURL = "/images/processed/" + strings.TrimSuffix(processedFilename, ".webp") + "_large.webp"

	case "audio":
		// For audio files, use original file path (no processing needed)
		attachment.URL = "/audio/" + filename // Assuming audio files go in /audio/

	case "document":
		// For documents (PDF, Word), use original file path
		attachment.URL = "/documents/" + filename // Assuming docs go in /documents/
	}

	return attachment
}

// getFileTypeFromMimeType determines file category from MIME type
func (s Story) getFileTypeFromMimeType(mimetype string) string {
	switch {
	case strings.HasPrefix(mimetype, "image/"):
		return "image"
	case strings.HasPrefix(mimetype, "audio/"):
		return "audio"
	case strings.HasPrefix(mimetype, "video/"):
		return "video"
	case mimetype == "application/pdf":
		return "document"
	case mimetype == "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "document"
	case mimetype == "application/msword":
		return "document"
	default:
		return "document" // Default fallback
	}
}
