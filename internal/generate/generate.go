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

// convertStoriesToJSON converts a slice of stories to a JSON array of URLs.
func convertStoriesToJSON(stories []data.Story) (string, error) {
	var urls []string
	for _, story := range stories {
		urls = append(urls, story.URL)
	}
	jsonData, err := json.Marshal(urls)
	if err != nil {
		return "", fmt.Errorf("failed to marshal stories to JSON: %w", err)
	}
	return string(jsonData), nil
}

// StoryData represents a story with all necessary data for filtering
type StoryData struct {
	ID            string     `json:"id"`
	Finding       string     `json:"finding"`
	URL           string     `json:"url"`
	Location      string     `json:"location"`
	StartDateTime string     `json:"startDateTime"`
	EndDateTime   string     `json:"endDateTime"`
	Season        string     `json:"season"`
	Experience    string     `json:"experience"`
	TimeSpan      string     `json:"timeSpan"`
	Themes        []string   `json:"themes"`
	Types         []string   `json:"types"`
	Weather       []string   `json:"weather"`
	Image         StoryImage `json:"image"`
}

// StoryImage represents the image data for a story
type StoryImage struct {
	URL       string `json:"url"`
	ThumbURL  string `json:"thumbUrl"`
	MediumURL string `json:"mediumUrl"`
	LargeURL  string `json:"largeUrl"`
	Alt       string `json:"alt"`
}

// FilterData represents all the data needed for client-side filtering
type FilterData struct {
	Themes  []FilterOption `json:"themes"`
	Types   []FilterOption `json:"types"`
	Weather []FilterOption `json:"weather"`
	Stories []StoryData    `json:"stories"`
}

// FilterOption represents a filter option with title and count
type FilterOption struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Count int    `json:"count"`
	Color string `json:"color"`
}

// convertStoriesToFilterData converts stories to comprehensive filter data
func convertStoriesToFilterData(stories []data.Story, themes []data.Theme, types []data.Type, weather []data.Weather) (string, error) {
	// Create story data
	storyData := make([]StoryData, len(stories))
	for i, story := range stories {
		image := story.GetStoryImage()

		// Extract theme titles
		themeNames := make([]string, len(story.Themes))
		for j, theme := range story.Themes {
			themeNames[j] = theme.Title
		}

		// Extract type titles
		typeNames := make([]string, len(story.Type))
		for j, typ := range story.Type {
			typeNames[j] = typ.Title
		}

		// Extract weather titles
		weatherNames := make([]string, len(story.Weather))
		for j, w := range story.Weather {
			weatherNames[j] = w.Title
		}

		storyData[i] = StoryData{
			ID:            story.ID,
			Finding:       story.Finding,
			URL:           story.URL,
			Location:      story.Location,
			StartDateTime: story.StartDateTime,
			EndDateTime:   story.EndDateTime,
			Season:        story.Season,
			Experience:    story.Experience,
			TimeSpan:      story.TimeSpan,
			Themes:        themeNames,
			Types:         typeNames,
			Weather:       weatherNames,
			Image: StoryImage{
				URL:       image.URL,
				ThumbURL:  image.ThumbURL,
				MediumURL: image.MediumURL,
				LargeURL:  image.LargeURL,
				Alt:       image.AlternativeText,
			},
		}
	}

	// Create filter options with counts
	themeOptions := make([]FilterOption, len(themes))
	for i, theme := range themes {
		count := 0
		for _, story := range stories {
			for _, storyTheme := range story.Themes {
				if storyTheme.Title == theme.Title {
					count++
					break
				}
			}
		}
		themeOptions[i] = FilterOption{
			Title: theme.Title,
			URL:   theme.URL,
			Count: count,
			Color: theme.Colour,
		}
	}

	typeOptions := make([]FilterOption, len(types))
	for i, typ := range types {
		count := 0
		for _, story := range stories {
			for _, storyType := range story.Type {
				if storyType.Title == typ.Title {
					count++
					break
				}
			}
		}
		typeOptions[i] = FilterOption{
			Title: typ.Title,
			URL:   typ.URL,
			Count: count,
			Color: typ.Colour,
		}
	}

	weatherOptions := make([]FilterOption, len(weather))
	for i, w := range weather {
		count := 0
		for _, story := range stories {
			for _, storyWeather := range story.Weather {
				if storyWeather.Title == w.Title {
					count++
					break
				}
			}
		}
		weatherOptions[i] = FilterOption{
			Title: w.Title,
			URL:   w.URL,
			Count: count,
			Color: w.Colour,
		}
	}

	filterData := FilterData{
		Themes:  themeOptions,
		Types:   typeOptions,
		Weather: weatherOptions,
		Stories: storyData,
	}

	jsonData, err := json.Marshal(filterData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal filter data to JSON: %w", err)
	}
	return string(jsonData), nil
}

// WriteWeatherIndexes generates the weather index pages and writes them to the out/weather directory.
func WriteWeatherIndexes() error {
	log.Println("Starting weather generation")
	weathers := store.GetWeather()
	allStories := store.GetAllStories()

	// Convert stories to JSON
	storiesJSON, err := convertStoriesToJSON(allStories)
	if err != nil {
		return err
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
			StoriesJSON:    storiesJSON,
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
	storiesJSON, err := convertStoriesToJSON(allStories)
	if err != nil {
		return err
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
			StoriesJSON:    storiesJSON,
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

	// Convert stories to JSON
	storiesJSON, err := convertStoriesToJSON(stories)
	if err != nil {
		return err
	}

	tmpl, err := loadTemplates()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
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
			StoriesJSON:             storiesJSON,
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
	storiesJSON, err := convertStoriesToJSON(allStories)
	if err != nil {
		return err
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
			StoriesJSON:    storiesJSON,
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		log.Printf("Successfully wrote types to %s", outputPath)
	}

	return nil
}

// WriteFilterData generates a comprehensive JSON file with all filter data for client-side filtering
func WriteFilterData() error {
	log.Println("Starting filter data generation")
	themes := store.GetThemes()
	types := store.GetTypes()
	weather := store.GetWeather()
	allStories := store.GetAllStories()

	// Convert to filter data JSON
	filterDataJSON, err := convertStoriesToFilterData(allStories, themes, types, weather)
	if err != nil {
		return err
	}

	fileName := "filter-data.json"
	outputPath := filepath.Join("out", fileName)

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create filter data file %s: %w", outputPath, err)
	}
	defer file.Close()

	_, err = file.WriteString(filterDataJSON)
	if err != nil {
		return fmt.Errorf("failed to write filter data: %w", err)
	}

	log.Printf("Successfully wrote filter data to %s", outputPath)
	return nil
}

// WriteWanderPage generates the wander page HTML file and writes it to the out/ directory.
func WriteWanderPage() error {
	log.Println("Starting wander page generation")
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
	storiesJSON, err := convertStoriesToJSON(stories)
	if err != nil {
		return err
	}

	page := data.Page{
		Title:          "Wander – Dudley People's School for Climate Justice",
		Description:    "Wander through a random selection of stories from the Dudley Climate Justice Archive",
		Themes:         themes,
		Types:          types,
		Stories:        stories,
		RandomStoryURL: randomStory.URL,
		StoriesJSON:    storiesJSON,
	}

	tmpl, err := loadTemplates()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	file, err := os.Create("out/wander.html")
	if err != nil {
		return fmt.Errorf("failed to create wander.html: %w", err)
	}
	defer file.Close()

	err = tmpl.ExecuteTemplate(file, "wander.html", page)
	if err != nil {
		return fmt.Errorf("failed to execute wander template: %w", err)
	}

	log.Println("Wander page generated successfully")
	return nil
}

// WriteArchivePage generates the archive page HTML file and writes it to the out/ directory.
func WriteArchivePage() error {
	log.Println("Starting archive page generation")
	themes := store.GetThemes()
	types := store.GetTypes()
	allStories := store.GetAllStories()

	// Select a random story for the initial random link
	randomStory := allStories[rand.Intn(len(allStories))]

	// Shuffle all stories and take first 40 for initial display
	shuffledStories := make([]data.Story, len(allStories))
	copy(shuffledStories, allStories)
	rand.Shuffle(len(shuffledStories), func(i, j int) {
		shuffledStories[i], shuffledStories[j] = shuffledStories[j], shuffledStories[i]
	})

	// Take first 40 stories for initial display
	stories := shuffledStories
	if len(stories) > 40 {
		stories = stories[:40]
	}

	// Convert stories to JSON (only the 40 displayed ones)
	storiesJSON, err := convertStoriesToJSON(stories)
	if err != nil {
		return err
	}

	page := data.Page{
		Title:          "Archive – Dudley People's School for Climate Justice",
		Description:    "Explore the complete Dudley Climate Justice Archive with filters for themes, types, and weather",
		Themes:         themes,
		Types:          types,
		Stories:        stories, // Only 40 random stories for initial display
		RandomStoryURL: randomStory.URL,
		StoriesJSON:    storiesJSON,
	}

	tmpl, err := template.ParseFiles("templates/archive.html", "templates/partials/header.html", "templates/partials/footer.html", "templates/partials/stories-list.html")
	if err != nil {
		return fmt.Errorf("failed to parse archive template: %w", err)
	}

	file, err := os.Create("out/archive.html")
	if err != nil {
		return fmt.Errorf("failed to create archive.html: %w", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, page)
	if err != nil {
		return fmt.Errorf("failed to execute archive template: %w", err)
	}

	log.Println("Archive page generated successfully")
	return nil
}

// WriteHomePage generates the homepage HTML file and writes it to the out/ directory.
func WriteHomePage() error {
	log.Println("Starting homepage generation")
	themes := store.GetThemes()
	types := store.GetTypes()
	allStories := store.GetAllStories()

	// Only give the template the first 40 stories for initial display
	stories := allStories[:40]

	// Shuffle the stories
	rand.Shuffle(len(stories), func(i, j int) {
		stories[i], stories[j] = stories[j], stories[i]
	})

	// Select a random story for the initial random link
	randomStory := stories[rand.Intn(len(stories))]

	// Convert stories to JSON (keep existing functionality)
	storiesJSON, err := convertStoriesToJSON(stories)
	if err != nil {
		return err
	}

	page := data.Page{
		Title:          "Dudley People's School for Climate Justice – time portal",
		Description:    "The time portal for the Dudley People's School for Climate Justice",
		Themes:         themes,
		Types:          types,
		Stories:        stories,
		RandomStoryURL: randomStory.URL,
		StoriesJSON:    storiesJSON,
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
