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
	"community-climate-justice-archive/internal/util"
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

// createTypeIndexOutputPathFromTitle creates a path to the output file for a type index page.
func createTypeIndexOutputPathFromTitle(title string) string {
	lowerCaseTitle := strings.ToLower(title)
	fileName := fmt.Sprintf("%s.html", lowerCaseTitle)
	return filepath.Join("out", "types", fileName)
}

// createThemeIndexOutputPathFromTitle creates a path to the output file for a type index page.
func createThemeIndexOutputPathFromTitle(title string) string {
	lowerCaseTitle := strings.ToLower(title)
	fileName := fmt.Sprintf("%s.html", lowerCaseTitle)
	return filepath.Join("out", "themes", fileName)
}

// createStoryOutputPathFromFinding creates a path to the output file for a story page.
func createStoryOutputPathFromFinding(finding string) string {
	slug := util.Slugify(finding)
	fileName := fmt.Sprintf("%s.html", slug)
	return filepath.Join("out", "stories", fileName)
}

// WriteTypesIndexes generates the type index pages and writes them to the out/types directory.
func WriteTypesIndexes() error {
	log.Println("Starting types generation")
	types := store.GetTypes()

	tmpl, err := template.ParseFiles("templates/type-index.html")
	if err != nil {
		return fmt.Errorf("failed to parse types template: %w", err)
	}

	err = os.MkdirAll("out/types", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output types directory: %w", err)
	}

	for _, typeInQuestion := range types {
		outputPath := createTypeIndexOutputPathFromTitle(typeInQuestion.Title)

		log.Printf("Writing types %s to %s", typeInQuestion.Title, outputPath)

		file, err := os.Create(outputPath)

		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}
		defer file.Close()

		stories := store.GetStoriesForType(typeInQuestion.Title)

		err = tmpl.Execute(file, data.TaxonomyIndexPage{
			Title:       typeInQuestion.Title,
			Description: "A list of stories for the type " + typeInQuestion.Title,
			Stories:     stories,
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		log.Printf("Successfully wrote types to %s", outputPath)
	}

	return nil
}

// WriteStories generates a story page for each story and writes them to the out/stories directory.
func WriteStories() error {
	log.Println("Starting story generation")
	stories := store.GetAllStories()

	tmpl, err := template.ParseFiles("templates/story.html")
	if err != nil {
		return fmt.Errorf("failed to parse story template: %w", err)
	}

	err = os.MkdirAll("out/stories", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output stories directory: %w", err)
	}

	for _, storyInQuestion := range stories {
		outputPath := createStoryOutputPathFromFinding(storyInQuestion.Finding)

		log.Printf("Writing story with finding %s to %s", storyInQuestion.Finding, outputPath)

		file, err := os.Create(outputPath)

		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}
		defer file.Close()

		err = tmpl.Execute(file, data.StoryPage{
			Title:       storyInQuestion.Finding,
			Description: "A story that says:" + storyInQuestion.Finding,
			Story:       storyInQuestion,
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		log.Printf("Successfully wrote story to %s", outputPath)
	}

	return nil
}

// WriteThemesIndexes generates the theme index pages and writes them to the out/themes directory.
func WriteThemesIndexes() error {
	log.Println("Starting themes generation")
	themes := store.GetThemes()

	tmpl, err := template.ParseFiles("templates/theme-index.html")
	if err != nil {
		return fmt.Errorf("failed to parse types template: %w", err)
	}

	err = os.MkdirAll("out/themes", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output types directory: %w", err)
	}

	for _, themeInQuestion := range themes {
		outputPath := createThemeIndexOutputPathFromTitle(themeInQuestion.Title)

		log.Printf("Writing types %s to %s", themeInQuestion.Title, outputPath)

		file, err := os.Create(outputPath)

		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}
		defer file.Close()

		stories := store.GetStoriesForTheme(themeInQuestion.Title)

		err = tmpl.Execute(file, data.TaxonomyIndexPage{
			Title:       themeInQuestion.Title,
			Description: "A list of stories for the theme " + themeInQuestion.Title,
			Stories:     stories,
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		log.Printf("Successfully wrote types to %s", outputPath)
	}

	return nil
}

// WriteHomePage generates the homepage HTML file and writes it to the out/ directory.
func WriteHomePage() error {
	log.Println("Starting homepage generation")
	themes := store.GetThemes()
	types := store.GetTypes()
	stories := store.GetAllStories()

	page := data.Page{
		Title:       "Dudley People's School for Climate Justice – time portal",
		Description: "The time portal for the Dudley People's School for Climate Justice",
		Themes:      themes,
		Types:       types,
		Stories:     stories,
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
