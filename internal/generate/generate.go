// Package generate provides functionality for generating static HTML pages
// from templates and data for the Dudley People's School for Climate Justice website.
package generate

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/store"
	"community-climate-justice-archive/internal/util"
)

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

// createWeatherOutputPathFromTitle creates a path to the output file for a weather page.
func createWeatherOutputPathFromTitle(title string) string {
	lowerCaseTitle := strings.ToLower(title)
	fileName := fmt.Sprintf("%s.html", lowerCaseTitle)
	return filepath.Join("out", "weather", fileName)
}

// WriteWeatherIndexes generates the weather index pages and writes them to the out/weather directory.
func WriteWeatherIndexes() error {
	log.Println("Starting weather generation")
	weathers := store.GetWeather()

	tmpl, err := template.ParseFiles("templates/weather-index.html")
	if err != nil {
		return fmt.Errorf("failed to parse weather template: %w", err)
	}

	err = os.MkdirAll("out/weather", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output weather directory: %w", err)
	}

	for _, weatherInQuestion := range weathers {
		outputPath := createWeatherOutputPathFromTitle(weatherInQuestion.Title)

		log.Printf("Writing weather %s to %s", weatherInQuestion.Title, outputPath)

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}
		defer file.Close()

		stories := store.GetStoriesForWeather(weatherInQuestion.Title)

		err = tmpl.Execute(file, data.TaxonomyIndexPage{
			Title:       weatherInQuestion.Title,
			Description: "A list of stories for the weather " + weatherInQuestion.Title,
			Stories:     stories,
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		log.Printf("Successfully wrote weather to %s", outputPath)
	}

	return nil
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

	totalStories := len(stories)

	for i, storyInQuestion := range stories {
		outputPath := createStoryOutputPathFromFinding(storyInQuestion.Finding)

		log.Printf("Writing story with finding %s to %s", storyInQuestion.Finding, outputPath)

		file, err := os.Create(outputPath)

		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}
		defer file.Close()

		// Get previous story, wrapping to the end if at the beginning
		var previousStory data.Story
		if i > 0 {
			previousStory = stories[i-1]
		} else {
			// If this is the first story, the previous is the last story
			previousStory = stories[totalStories-1]
		}

		// Get next story, wrapping to the beginning if at the end
		var nextStory data.Story
		if i < totalStories-1 {
			nextStory = stories[i+1]
		} else {
			// If this is the last story, the next is the first story
			nextStory = stories[0]
		}

		// Reformat the date fields to be more human readable
		storyInQuestion.StartDateTime = util.FormatDate(storyInQuestion.StartDateTime)
		storyInQuestion.EndDateTime = util.FormatDate(storyInQuestion.EndDateTime)

		err = tmpl.Execute(file, data.StoryPage{
			Title:       storyInQuestion.Finding,
			Description: "A story that says:" + storyInQuestion.Finding,
			Story:       storyInQuestion,
			LastStory:   previousStory,
			NextStory:   nextStory,
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

	// Only give the template the first 40 stories
	stories = stories[:40]

	// Shuffle the stories
	rand.Shuffle(len(stories), func(i, j int) {
		stories[i], stories[j] = stories[j], stories[i]
	})

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
