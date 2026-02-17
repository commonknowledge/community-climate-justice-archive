// Package data contains the Story type and everything related to it.
//
// Stories are the heart of the archive - each one is something someone from the
// community has shared, like a photo, poem, video, drawing, or piece of writing.
//
// The main Story struct holds everything about a story:
// - What it's called and what it's about
// - When and where it happened
// - Pictures, audio, or documents attached to it
// - Tags that help organise it (themes, types, weather)
// - Connections to other stories
//
// The ImageVideoSound field is where attachments live. It comes from NocoDB as
// a JSON string, and the code here unpacks that into something we can work with -
// figuring out if it's an image, audio, or document, and creating the right URLs.
//
// This file also has helpful functions for working with stories, like creating
// consistent colours from text (so "Climate Change" always has the same colour
// wherever you see it).
package data

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

const nocodbStoryDashboardURLTemplate = "https://nocodb-r87d.onrender.com/dashboard/#/nc/pqw5yaekkqvo25h/me04vwwhvh4jbsg?rowId=%s&path="

// StoryConnection is when one story is connected to another.
//
// Stories can inspire each other, creating links throughout the archive. When you
// look at a story, you might see "This story inspired..." or "This story was inspired by..."
// - that's what these connections are for.
//
// This struct holds the information we need to show that related story.
type StoryConnection struct {
	ID                 string `json:"id"`                 // The ID of the other story
	Title              string `json:"title"`              // What the other story is called
	Finding            string `json:"finding"`            // The other story's title (we call it "Finding")
	Image              string `json:"image"`              // The main image from that story
	ThumbURL           string `json:"thumbUrl"`           // A smaller version of that image
	URL                string `json:"url"`                // Where to find that story's page
	AttachmentType     string `json:"attachmentType"`     // What kind of file is attached ("image", "audio", "document", or "none")
	AttachmentFilename string `json:"attachmentFilename"` // The filename (for when it's not an image)
}

// GiftedBy is who made or shared a story with the archive.
//
// This helps us recognise and celebrate all the different people and groups who've
// contributed - like local schools, community groups, or individuals. You can browse
// to see all the stories from a particular contributor.
type GiftedBy struct {
	Title  string // Who shared this story (like "Local School" or "Community Group")
	URL    string // Where to find all stories from this person or group
	Colour string // The colour for this contributor
}

// ScalePermanence is about how long-lasting something is.
//
// This is borrowed from permaculture (a way of thinking about sustainable gardening).
// It classifies things by how permanent they are - from temporary (like yearly flowers)
// to permanent (like buildings or hills).
//
// When applied to stories, it adds an interesting way to think about time. Some
// stories might be about fleeting moments, others about lasting changes.
type ScalePermanence struct {
	Title  string // How permanent it is (like "Temporary" or "Permanent")
	URL    string // Where to find all stories at this level of permanence
	Colour string // The colour for this permanence level
}

// WhatWasIsIf is about whether a story looks at the past, present, or future.
//
// Is it about something that already happened? Something happening now? Or imagining
// what might be?
//
// - "What Was": Memories, history, things that have been
// - "What Is": The present moment, what's happening now
// - "What If": Imagining futures, possibilities, dreams
//
// It's a lovely way to think about stories and how they relate to time.
type WhatWasIsIf struct {
	Title  string // Which it is: "What Was", "What Is", or "What If"
	URL    string // Where to find all stories with this perspective
	Colour string // The colour for this temporal perspective
}

// TimePeriod is when in history a story is from.
//
// This might be "1960s", "Victorian Era", "Present Day" - whatever time the story
// relates to. It helps people explore stories from particular eras.
type TimePeriod struct {
	Title  string // What time period it is (like "1960s" or "Victorian Era")
	URL    string // Where to find all stories from this period
	Colour string // The colour for this time period
}

// Story is a single contribution to the archive.
//
// Each story is something someone from the community has shared - could be a photo,
// poem, video, drawing, piece of writing, or anything really. This struct holds
// all the information about that story.
//
// Everything lives in here: what it's called, when it happened, where it was,
// what files are attached to it, how it's tagged, and which other stories it
// connects to.
//
// When you fetch stories from the database (NocoDB at the moment), you get them
// as Story structs. These then get passed to HTML templates to create the actual
// web pages.
type Story struct {
	ID                      string            // Unique identifier for the story in the database
	CreatedTime             string            // When this story was first added to the archive
	Finding                 string            // The primary title/name of the story
	HighStExperiment        string            // The project or event this story is associated with
	ImageVideoSound         string            // JSON string of attachments (images, audio, documents)
	Location                string            // Where the story takes place or was created
	StartDateTime           string            // When the story's events began or when it was experienced
	EndDateTime             string            // When the story's events ended or when it was added to archive
	Season                  string            // The season when the story occurred
	StreetDetectoristClue   string            // Clue text for the Street Detectorist map feature
	Themes                  []Theme           // Thematic categories this story belongs to
	Experience              string            // Description or narrative content of the story
	TimeSpan                string            // How long the events lasted or their duration
	InspiredBy              []StoryConnection // Other stories that inspired this one
	HasInspired             []StoryConnection // Other stories this one has inspired
	OtherComments           string            // Additional notes or context about the story
	Type                    []Type            // Format types (Photo, Poem, Video, etc.)
	Weather                 []Weather         // Weather conditions associated with the story
	GiftedBy                []GiftedBy        // Who contributed or co-created this story
	ScalePermanence         []ScalePermanence // Permanence classification (permaculture concept)
	WhatWasIsIf             []WhatWasIsIf     // Temporal perspective (past/present/future)
	TimePeriod              []TimePeriod      // Historical era or timeframe
	PersonFinder            string            // Person who found or contributed the story
	MapCache                string            // Cached map data for location visualization
	MapSize                 string            // Size specifications for the map display
	Created                 string            // Creation timestamp (may differ from CreatedTime)
	StreetDetectoristMapURL string            // URL to the Street Detectorist map for this story
	OtherTheme              string            // Custom theme not in the predefined list
	OtherWeather            string            // Custom weather condition not in the predefined list
	ShareStatus             string            // Publishing status (published, draft, etc.)
	PostDate                string            // When this story was officially published
	TwitterText             string            // Text for Twitter sharing
	CharacterCount          string            // Character count for Twitter text
	InstaText               string            // Text for Instagram sharing
	InstaCount              string            // Character count for Instagram text
	InstaImage              string            // Image optimized for Instagram sharing
	ReflectionLearning      string            // Reflections or learnings associated with the story
	UpdatedAt               string            // When this story was last modified
	URL                     string            // The URL path to this story's page in the archive
}

// StoryAttachment is a file attached to a story.
//
// Stories can have pictures, audio files, or documents attached to them. This struct
// holds all the information we need about each file - its name, size, type, and
// where to find it.
//
// We handle different types of files in different ways:
// - Pictures get resized into different sizes and converted to WebP (better compression)
// - Audio files get an audio player
// - Documents (PDFs, Word files) get a download link
type StoryAttachment struct {
	Filename        string               // What the file is called (like "photo.jpg")
	AlternativeText string               // Description for screen readers (usually just the filename)
	Type            string               // The technical file type (like "image/jpeg" or "audio/mp3")
	FileType        string               // Simple category: "image", "audio", or "document"
	Size            int                  // How big the file is, in bytes
	Width           int                  // How wide the image is (0 if it's not an image)
	Height          int                  // How tall the image is (0 if it's not an image)
	URL             string               // Where to find the full-size file
	ThumbURL        string               // Where to find the small thumbnail (images only)
	MediumURL       string               // Where to find the medium size (images only)
	LargeURL        string               // Where to find the large size (images only)
	Thumbnails      map[string]Thumbnail // Extra thumbnail info (not used much anymore)
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
// GetFileTypeFromExtension works out what kind of file something is.
//
// When we have a file attached to a story, we need to know: is it a picture? An
// audio file? A document? This function looks at the file extension (the bit after
// the dot) and tells us.
//
// Here's how files get sorted:
// - Pictures: .jpg, .jpeg, .png, .gif, .bmp, .webp, .tiff, .svg
// - Audio: .mp3, .wav, .ogg, .m4a, .aac, .flac
// - Documents: .pdf, .doc, .docx, .txt, .rtf (and anything else we don't recognise)
//
// Why this matters:
// - Pictures get resized and converted to WebP for faster loading
// - Audio files get a player interface
// - Documents get a download button
//
// So "photo.jpg" → "image", "recording.mp3" → "audio", "essay.pdf" → "document"
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

// GetStoryAttachments gets all the files attached to a story.
//
// Stories can have images, audio files, or documents attached. NocoDB stores these
// in a JSON format in the ImageVideoSound field, and this function unpacks that
// into something we can actually work with.
//
// What it does:
// - Takes the ImageVideoSound JSON string
// - Parses it to extract file information (names, sizes, types, etc.)
// - Converts each file into a StoryAttachment struct
// - Returns them all as a list
//
// If there's no ImageVideoSound data or something goes wrong with the parsing,
// you just get an empty list back. Better to show a story without pictures than
// to break the whole site!
func (s Story) GetStoryAttachments() []StoryAttachment {
	var allAttachments []StoryAttachment

	// The ImageVideoSound field contains all the attached files as JSON from NocoDB
	if s.ImageVideoSound != "" {
		// First, parse the JSON into a list of generic data structures
		// (NocoDB sends us stuff like "title", "mimetype", "size", "width", "height")
		var nocoAttachments []map[string]interface{}
		if err := json.Unmarshal([]byte(s.ImageVideoSound), &nocoAttachments); err != nil {
			// If the JSON is malformed, log a warning but keep going - we'd rather
			// show a story without its attachments than crash the whole thing
			log.Printf("Warning: Failed to unmarshal ImageVideoSound field for story %s: %v", s.ID, err)
		} else {
			// Now convert each attachment from NocoDB's format into our format
			for _, nocoAttachment := range nocoAttachments {
				attachment := s.convertNocoDBAttachment(nocoAttachment)
				// Only keep attachments that actually have a filename
				// (filters out any weird empty ones)
				if attachment.Filename != "" {
					allAttachments = append(allAttachments, attachment)
				}
			}
		}
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

// TitleToHexColor creates a colour from text that's always the same.
//
// We want each theme, type, and weather condition to have its own colour. But we
// don't want to manually pick colours - and we want "Climate Change" to always be
// the same colour wherever you see it.
//
// So this function does something clever: it turns the text into a colour in a way
// that's "random" but always gives the same result for the same input.
//
// The magic:
// 1. Hash the title into a number (SHA256 - this always gives the same number for the same text)
// 2. Use that number to seed a random generator
// 3. Generate a "random" colour (but it's always the same "random" for that title)
// 4. Make sure the colour looks nice (not too dark, not too washed out)
//
// So "Climate Change" always becomes (for example) #a3c4f2, "Community" always
// becomes some other specific colour, and so on. Consistent but automated!
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

// GetNocoDBURL returns a direct link to this story in the NocoDB interface for debugging
func (s Story) GetNocoDBURL() string {
	return fmt.Sprintf(nocodbStoryDashboardURLTemplate, s.ID)
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
