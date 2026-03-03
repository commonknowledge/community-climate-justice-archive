// Package generate builds all the HTML pages for the archive.
//
// This is where the website actually gets made! It takes story data from the database,
// runs it through HTML templates, and creates a complete static website in the out/
// folder.
//
// What gets built:
// - The homepage
// - The archive page (where you can filter and browse everything)
// - The "wander" page (for exploring stories randomly)
// - One page for each story
// - Pages for each theme, type, weather condition, contributor, time period, etc.
//
// How it works:
// 1. Fetch all the story data from the database
// 2. For each type of page, load the right HTML template
// 3. Fill in the template with the actual data (stories, tags, etc.)
// 4. Turn that into HTML
// 5. Save it as a file in the out/ folder
//
// Everything is static HTML - once it's built, you just need a simple web server
// to show it. No database or fancy server required. This makes the archive fast
// and easy to host anywhere!
package generate

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/store"
	"community-climate-justice-archive/internal/util"
)

var (
	// cachedStories stores converted Story structs for one build run.
	// This avoids repeated record->Story conversion across multiple page writers.
	cachedStories []data.Story
	// cachedTemplates stores parsed templates for one build run.
	// This avoids parsing templates repeatedly in each writer function.
	cachedTemplates *template.Template
)

const (
	initialStoriesDisplayCount = 40
	connectedStoriesLimit      = 20
	relatedStoriesLimit        = 5
)

// ResetBuildCache clears per-run cached data so each build starts fresh.
func ResetBuildCache() {
	cachedStories = nil
	cachedTemplates = nil
}

// WarmBuildCache primes story and template caches before parallel build steps run.
// This keeps cache writes out of concurrent page writers.
func WarmBuildCache() error {
	_ = getAllStories()

	if _, err := loadTemplatesCached(); err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	return nil
}

// createTypeIndexOutputPathFromTitle creates a path to the output file for a type index page.
func createTypeIndexOutputPathFromTitle(title string) string {
	slug := util.Slugify(title)
	fileName := fmt.Sprintf("%s.html", slug)
	return filepath.Join("out", "types", fileName)
}

// createThemeIndexOutputPathFromTitle creates a path to the output file for a theme index page.
func createThemeIndexOutputPathFromTitle(title string) string {
	slug := util.Slugify(title)
	fileName := fmt.Sprintf("%s.html", slug)
	return filepath.Join("out", "themes", fileName)
}

// createStoryOutputPathFromFindingWithID creates a path to the output file for a story page with ID suffix.
func createStoryOutputPathFromFindingWithID(finding, id string) string {
	slug := util.Slugify(finding)
	fileName := fmt.Sprintf("%s-%s.html", slug, id)
	return filepath.Join("out", "stories", fileName)
}

// createWeatherOutputPathFromTitle creates a path to the output file for a weather page.
func createWeatherOutputPathFromTitle(title string) string {
	slug := util.Slugify(title)
	fileName := fmt.Sprintf("%s.html", slug)
	return filepath.Join("out", "weather", fileName)
}

// createGiftedByOutputPathFromTitle creates a path to the output file for a gifted by page.
func createGiftedByOutputPathFromTitle(title string) string {
	slug := util.Slugify(title)
	fileName := fmt.Sprintf("%s.html", slug)
	return filepath.Join("out", "giftedby", fileName)
}

// createScalePermanenceOutputPathFromTitle creates a path to the output file for a scale permanence page.
func createScalePermanenceOutputPathFromTitle(title string) string {
	slug := util.Slugify(title)
	fileName := fmt.Sprintf("%s.html", slug)
	return filepath.Join("out", "scalepermanence", fileName)
}

// createWhatWasIsIfOutputPathFromTitle creates a path to the output file for a what was/is/if page.
func createWhatWasIsIfOutputPathFromTitle(title string) string {
	slug := util.Slugify(title)
	fileName := fmt.Sprintf("%s.html", slug)
	return filepath.Join("out", "whatwasisif", fileName)
}

// createTimePeriodOutputPathFromTitle creates a path to the output file for a time period page.
func createTimePeriodOutputPathFromTitle(title string) string {
	slug := util.Slugify(title)
	fileName := fmt.Sprintf("%s.html", slug)
	return filepath.Join("out", "timeperiod", fileName)
}

// loadTemplates loads all templates and partials needed by the application.
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

func loadTemplatesCached() (*template.Template, error) {
	if cachedTemplates != nil {
		return cachedTemplates, nil
	}

	tmpl, err := loadTemplates()
	if err != nil {
		return nil, err
	}

	cachedTemplates = tmpl
	return cachedTemplates, nil
}

func getAllStories() []data.Story {
	if cachedStories != nil {
		return cachedStories
	}

	cachedStories = store.GetAllStories()
	return cachedStories
}

func randomStoryURL(stories []data.Story) string {
	if len(stories) == 0 {
		return ""
	}
	return stories[rand.Intn(len(stories))].URL
}

func limitStories(stories []data.Story, count int) []data.Story {
	if len(stories) <= count {
		return stories
	}
	return stories[:count]
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
	ID            string          `json:"id"`
	Finding       string          `json:"finding"`
	URL           string          `json:"url"`
	Location      string          `json:"location"`
	StartDateTime string          `json:"startDateTime"`
	EndDateTime   string          `json:"endDateTime"`
	Season        string          `json:"season"`
	Experience    string          `json:"experience"`
	TimeSpan      string          `json:"timeSpan"`
	Themes        []string        `json:"themes"`
	Types         []string        `json:"types"`
	Weather       []string        `json:"weather"`
	Attachment    StoryAttachment `json:"attachment"`
}

// StoryAttachment represents the attachment data for a story
type StoryAttachment struct {
	URL       string `json:"url"`
	ThumbURL  string `json:"thumbUrl"`
	MediumURL string `json:"mediumUrl"`
	LargeURL  string `json:"largeUrl"`
	Alt       string `json:"alt"`
	FileType  string `json:"fileType"`
	Filename  string `json:"filename"`
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
		attachment := story.GetStoryAttachment()

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
			Attachment: StoryAttachment{
				URL:       attachment.URL,
				ThumbURL:  attachment.ThumbURL,
				MediumURL: attachment.MediumURL,
				LargeURL:  attachment.LargeURL,
				Alt:       attachment.AlternativeText,
				FileType:  attachment.FileType,
				Filename:  attachment.Filename,
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
	stories := getAllStories()
	if len(stories) == 0 {
		log.Println("No stories found; skipping story page generation")
		return nil
	}

	// Convert stories to JSON
	storiesJSON, err := convertStoriesToJSON(stories)
	if err != nil {
		return err
	}

	tmpl, err := loadTemplatesCached()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// Clean out the stories directory to remove any files from filtered-out stories
	// This ensures that unapproved stories don't persist from previous builds
	if err := os.RemoveAll("out/stories"); err != nil {
		return fmt.Errorf("failed to clean output stories directory: %w", err)
	}
	log.Println("Cleaned out/stories directory")

	err = os.MkdirAll("out/stories", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output stories directory: %w", err)
	}

	totalStories := len(stories)

	for i, storyInQuestion := range stories {
		outputPath := createStoryOutputPathFromFindingWithID(storyInQuestion.Finding, storyInQuestion.ID)

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
		storyInQuestion.CreatedTime = util.FormatDate(storyInQuestion.CreatedTime)
		storyInQuestion.UpdatedAt = util.FormatDate(storyInQuestion.UpdatedAt)

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
			firstMoreTaggedStories = store.GetMoreTaggedStories(storyInQuestion, firstTag, relatedStoriesLimit)
		}

		var secondMoreTaggedStories []data.Story
		var thirdMoreTaggedStories []data.Story

		var secondTag interface{}
		var thirdTag interface{}

		if len(allTagsWeHave) > 1 {
			secondTag = allTagsWeHave[1]
			secondMoreTaggedStories = store.GetMoreTaggedStories(storyInQuestion, secondTag, relatedStoriesLimit)
		}

		if len(allTagsWeHave) > 2 {
			thirdTag = allTagsWeHave[2]
			thirdMoreTaggedStories = store.GetMoreTaggedStories(storyInQuestion, thirdTag, relatedStoriesLimit)
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

		// Pre-compute attachments to avoid calling method from template
		attachments := storyInQuestion.GetStoryAttachments()

		err = tmpl.ExecuteTemplate(file, "story.html", data.StoryPage{
			Title:                   storyInQuestion.Finding,
			Description:             "A story that says:" + storyInQuestion.Finding,
			Story:                   storyInQuestion,
			Attachments:             attachments,
			NocoDBURL:               storyInQuestion.GetNocoDBURL(),
			LastStory:               previousStory,
			NextStory:               nextStory,
			FirstMoreTaggedStories:  firstRelated,
			SecondMoreTaggedStories: secondRelated,
			ThirdMoreTaggedStories:  thirdRelated,
			RandomStoryURL:          randomStoryURL(stories),
			StoriesJSON:             storiesJSON,
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		if (i+1)%100 == 0 {
			log.Printf("Story generation progress: %d/%d", i+1, totalStories)
		}
	}

	log.Printf("Story generation complete: %d pages", totalStories)
	return nil
}

// WriteThemesIndexes generates the theme index pages and writes them to the out/themes directory.

// WriteFilterData generates a comprehensive JSON file with all filter data for client-side filtering
func WriteFilterData() error {
	log.Println("Starting filter data generation")
	themes := store.GetThemes()
	types := store.GetTypes()
	weather := store.GetWeather()
	allStories := getAllStories()

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
