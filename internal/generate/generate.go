// Package generate provides functionality for generating static HTML pages
// from templates and data for the Dudley People's School for Climate Justice website.
package generate

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/store"
)

// CopyImagesToOutput copies all images from the images directory to the out/images directory.
func CopyImagesToOutput() error {
	log.Println("Starting image copy process")

	err := os.MkdirAll("out/images", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output images directory: %w", err)
	}

	files, err := os.ReadDir("images")
	if err != nil {
		return fmt.Errorf("failed to read images directory: %w", err)
	}

	copyCount := 0
	skippedCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		ext := strings.ToLower(filepath.Ext(filename))

		// Only copy image files
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
			srcPath := filepath.Join("images", filename)
			dstPath := filepath.Join("out/images", filename)

			src, err := os.Open(srcPath)
			if err != nil {
				return fmt.Errorf("failed to open source image %s: %w", srcPath, err)
			}
			defer src.Close()

			dst, err := os.Create(dstPath)
			if err != nil {
				return fmt.Errorf("failed to create destination image %s: %w", dstPath, err)
			}
			defer dst.Close()

			_, err = io.Copy(dst, src)
			if err != nil {
				return fmt.Errorf("failed to copy image %s: %w", filename, err)
			}

			copyCount++
			log.Printf("Copied %s", filename)
		default:
			skippedCount++
			log.Printf("Skipped non-image file: %s", filename)
		}
	}

	log.Printf("Successfully copied %d images to output directory (skipped %d non-image files)", copyCount, skippedCount)
	return nil
}

// WriteHomePage generates the homepage HTML file and writes it to the out/ directory.
func WriteHomePage() error {
	log.Println("Starting homepage generation")
	themes := store.GetThemes()
	types := store.GetTypes()
	images := store.GetImages()

	page := data.Page{
		Title:       "Dudley People's School for Climate Justice – time portal",
		Description: "The time portal for the Dudley People's School for Climate Justice",
		Themes:      themes,
		Types:       types,
		Images:      images,
	}

	tmpl, err := template.ParseFiles("templates/homepage.html")
	if err != nil {
		return fmt.Errorf("failed to parse homepage template: %w", err)
	}

	fileName := "index.html"
	outputPath := filepath.Join("out", fileName)

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer file.Close()

	err = tmpl.Execute(file, page)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	log.Printf("Successfully wrote homepage to %s", outputPath)
	return nil
}
