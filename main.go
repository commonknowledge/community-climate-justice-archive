package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"community-climate-justice-archive/data"
)

func getThemes() []data.Theme {
	log.Println("Getting themes")

	return []data.Theme{
		{
			Title: "people",
		},
		{
			Title: "planet",
		},
		{
			Title: "architecture",
		},
		{
			Title: "old",
		},
	}
}

func getImages() []data.StoryImage {
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

func getTypes() []data.Type {
	log.Println("Getting types")

	return []data.Type{
		{Title: "text"},
		{Title: "image"},
		{Title: "video"},
		{Title: "sound"},
		{Title: "textile"},
	}
}

func writeHomePage() error {
	log.Println("Starting homepage generation")
	themes := getThemes()
	types := getTypes()
	images := getImages()

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

func copyImagesToOutput() error {
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

func serve() {
	log.Println("Serving current directory at http://localhost:8080")

	err := http.ListenAndServe(":8080", http.FileServer(http.Dir("./out")))
	if err != nil {
		log.Printf("Server error: %v", err)
	}
}

func regenerate() error {
	log.Println("Starting build process")

	if err := writeHomePage(); err != nil {
		return fmt.Errorf("failed to write homepage: %v", err)
	}

	if err := copyImagesToOutput(); err != nil {
		return fmt.Errorf("failed to copy images: %v", err)
	}

	log.Println("Build process completed successfully")
	return nil
}

func waitForInput() {
	log.Println("Press enter to regenerate the archive...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadRune()
}

func main() {
	// Initial build
	if err := regenerate(); err != nil {
		log.Fatalf("Initial build failed: %v", err)
	}

	go serve()

	for {
		waitForInput()
		if err := regenerate(); err != nil {
			log.Printf("Regeneration failed: %v", err)
		}
	}
}
