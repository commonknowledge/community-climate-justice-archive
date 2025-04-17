// Package generate provides functionality for generating static HTML pages
// from templates and data for the Dudley People's School for Climate Justice website.
package generate

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
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

// loadTemplates loads all templates and partials needed by the application
func loadTemplates() (*template.Template, error) {
	tmpl := template.New("")

	// Parse all HTML files in templates directory
	tmpl, err := tmpl.ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}

	// Parse all partials
	tmpl, err = tmpl.ParseGlob("templates/partials/*.html")
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// WriteWeatherIndexes generates the weather index pages and writes them to the out/weather directory.
func WriteWeatherIndexes() error {
	log.Println("Starting weather generation")
	weathers := store.GetWeather()
	allStories := store.GetAllStories()

	// Convert stories to JSON
	storiesJSON, err := json.Marshal(allStories)
	if err != nil {
		return fmt.Errorf("failed to marshal stories to JSON: %w", err)
	}

	tmpl, err := loadTemplates()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
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

		// Select a random story for the random link
		randomStory := allStories[rand.Intn(len(allStories))]

		err = tmpl.ExecuteTemplate(file, "weather-index.html", data.TaxonomyIndexPage{
			Title:          weatherInQuestion.Title,
			Description:    "A list of stories for the weather " + weatherInQuestion.Title,
			Stories:        stories,
			TaxonomyColour: weatherInQuestion.Colour,
			RandomStoryURL: randomStory.URL,
			StoriesJSON:    string(storiesJSON),
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
	allStories := store.GetAllStories()

	// Convert stories to JSON
	storiesJSON, err := json.Marshal(allStories)
	if err != nil {
		return fmt.Errorf("failed to marshal stories to JSON: %w", err)
	}

	tmpl, err := loadTemplates()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
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

		// Select a random story for the random link
		randomStory := allStories[rand.Intn(len(allStories))]

		err = tmpl.ExecuteTemplate(file, "type-index.html", data.TaxonomyIndexPage{
			Title:          typeInQuestion.Title,
			Description:    "A list of stories for the type " + typeInQuestion.Title,
			Stories:        stories,
			TaxonomyColour: typeInQuestion.Colour,
			RandomStoryURL: randomStory.URL,
			StoriesJSON:    string(storiesJSON),
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		log.Printf("Successfully wrote types to %s", outputPath)
	}

	return nil
}

// extractURLsAndMakeLinks extracts URLs from a string and makes them links.
//
// These links open in a new tab.
func extractURLsAndMakeLinks(text string) string {
	re := regexp.MustCompile(`(https?:\/\/[^\s]+)`)
	return re.ReplaceAllString(text, `<a href="$1" target="_blank" rel="noopener noreferrer">$1</a>`)
}

func getTagType(tag interface{}) string {
	if tag == nil {
		return "unknown"
	}

	switch tag.(type) {
	case data.Theme:
		return "theme"
	case data.Type:
		return "type"
	case data.Weather:
		return "weather"
	default:
		log.Printf("Unknown tag type: %v", tag)
		return "unknown"
	}
}

// WriteStories generates a story page for each story and writes them to the out/stories directory.
func WriteStories() error {
	log.Println("Starting story generation")
	stories := store.GetAllStories()

	tmpl, err := loadTemplates()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	err = os.MkdirAll("out/stories", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output stories directory: %w", err)
	}

	totalStories := len(stories)

	// Convert stories to JSON
	storiesJSON, err := json.Marshal(stories)
	if err != nil {
		return fmt.Errorf("failed to marshal stories to JSON: %w", err)
	}

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

		// Select a random story for the random link
		randomStory := stories[rand.Intn(len(stories))]

		// Reformat the date fields to be more human readable
		storyInQuestion.StartDateTime = util.FormatDate(storyInQuestion.StartDateTime)
		storyInQuestion.EndDateTime = util.FormatDate(storyInQuestion.EndDateTime)

		// For the "Other Comments" field, we want to extract the URLs and make them links
		storyInQuestion.OtherComments = extractURLsAndMakeLinks(storyInQuestion.OtherComments)

		// Chuck all the tagged story of things into a slice
		var allTagsWeHave []interface{}

		// Convert each tag type to []interface{} before appending
		// We aren't doing this for weather tags, yet
		for _, theme := range storyInQuestion.Themes {
			allTagsWeHave = append(allTagsWeHave, theme)
		}

		for _, typeTag := range storyInQuestion.Type {
			allTagsWeHave = append(allTagsWeHave, typeTag)
		}

		var firstTag interface{}
		var firstMoreTaggedStories []data.Story

		// Shuffle the tags
		rand.Shuffle(len(allTagsWeHave), func(i, j int) {
			allTagsWeHave[i], allTagsWeHave[j] = allTagsWeHave[j], allTagsWeHave[i]
		})

		if len(allTagsWeHave) > 0 {
			firstTag = allTagsWeHave[0]
			firstMoreTaggedStories = store.GetMoreTaggedStories(storyInQuestion, firstTag, 5)
		}

		var secondMoreTaggedStories []data.Story
		var thirdMoreTaggedStories []data.Story

		var secondTag interface{}
		var thirdTag interface{}

		if len(allTagsWeHave) > 1 {
			secondTag = allTagsWeHave[1]
			secondMoreTaggedStories = store.GetMoreTaggedStories(storyInQuestion, secondTag, 5)
		}

		if len(allTagsWeHave) > 2 {
			thirdTag = allTagsWeHave[2]
			thirdMoreTaggedStories = store.GetMoreTaggedStories(storyInQuestion, thirdTag, 5)
		}

		var firstRelated data.RelatedStories
		var secondRelated data.RelatedStories
		var thirdRelated data.RelatedStories

		if len(firstMoreTaggedStories) > 0 && firstTag != nil {
			firstRelated = data.RelatedStories{
				Tag:     firstTag,
				TagType: getTagType(firstTag),
				Stories: firstMoreTaggedStories,
			}
		}

		if len(secondMoreTaggedStories) > 0 && secondTag != nil {
			secondRelated = data.RelatedStories{
				Tag:     secondTag,
				TagType: getTagType(secondTag),
				Stories: secondMoreTaggedStories,
			}
		}

		if len(thirdMoreTaggedStories) > 0 && thirdTag != nil {
			thirdRelated = data.RelatedStories{
				Tag:     thirdTag,
				TagType: getTagType(thirdTag),
				Stories: thirdMoreTaggedStories,
			}
		}

		err = tmpl.ExecuteTemplate(file, "story.html", data.StoryPage{
			Title:                   storyInQuestion.Finding,
			Description:             "A story that says:" + storyInQuestion.Finding,
			Story:                   storyInQuestion,
			LastStory:               previousStory,
			NextStory:               nextStory,
			FirstMoreTaggedStories:  firstRelated,
			SecondMoreTaggedStories: secondRelated,
			ThirdMoreTaggedStories:  thirdRelated,
			RandomStoryURL:          randomStory.URL,
			StoriesJSON:             string(storiesJSON),
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
	allStories := store.GetAllStories()

	// Convert stories to JSON
	storiesJSON, err := json.Marshal(allStories)
	if err != nil {
		return fmt.Errorf("failed to marshal stories to JSON: %w", err)
	}

	tmpl, err := loadTemplates()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
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

		// Select a random story for the random link
		randomStory := allStories[rand.Intn(len(allStories))]

		err = tmpl.ExecuteTemplate(file, "theme-index.html", data.TaxonomyIndexPage{
			Title:          themeInQuestion.Title,
			Description:    "A list of stories for the theme " + themeInQuestion.Title,
			Stories:        stories,
			TaxonomyColour: themeInQuestion.Colour,
			RandomStoryURL: randomStory.URL,
			StoriesJSON:    string(storiesJSON),
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

	// Select a random story for the initial random link
	randomStory := stories[rand.Intn(len(stories))]

	// Convert stories to JSON
	storiesJSON, err := json.Marshal(stories)
	if err != nil {
		return fmt.Errorf("failed to marshal stories to JSON: %w", err)
	}

	page := data.Page{
		Title:          "Dudley People's School for Climate Justice – time portal",
		Description:    "The time portal for the Dudley People's School for Climate Justice",
		Themes:         themes,
		Types:          types,
		Stories:        stories,
		RandomStoryURL: randomStory.URL,
		StoriesJSON:    string(storiesJSON),
	}

	tmpl, err := loadTemplates()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	fileName := "index.html"
	outputPath := filepath.Join("out", fileName)

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer file.Close()

	err = tmpl.ExecuteTemplate(file, "homepage.html", page)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	log.Printf("Successfully wrote homepage to %s", outputPath)

	return nil
}
