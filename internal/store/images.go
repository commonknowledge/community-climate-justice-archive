// Retrieves and processes images as required.
package store

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"community-climate-justice-archive/data"
)

// GetImages retrieves all images from the images directory and returns them as a slice of StoryImage.
// Intended for passing to HTML templates.
// For the moment this is hardcoded, but we will use SQLite to populate this.
func GetImages() []data.StoryImage {
	log.Println("Getting images")

	var images []data.StoryImage

	files, err := os.ReadDir("images")
	if err != nil {
		log.Printf("Error reading images directory: %v", err)
		return images
	}

	processedCount := 0
	for _, file := range files {
		if !file.IsDir() {
			filename := file.Name()
			ext := strings.ToLower(filepath.Ext(filename))

			// Common image extensions
			switch ext {
			case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
				images = append(images, data.StoryImage{Filename: filename, AlternativeText: filename})
				processedCount++
			}
		}
	}

	log.Printf("Processed %d images from images directory", processedCount)
	if len(images) == 0 {
		log.Println("Warning: No valid images found in images directory")
	}

	return images
}
