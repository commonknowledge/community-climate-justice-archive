package generate

import (
	"fmt"
	"log"
	"os"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/store"
)

// WriteWeatherIndexes generates the weather index pages and writes them to the out/weather directory.
func WriteWeatherIndexes() error {
	log.Println("Starting weather generation")
	weathers := store.GetWeather()
	allStories := getAllStories()

	// Convert stories to JSON
	storiesJSON, err := convertStoriesToJSON(allStories)
	if err != nil {
		return err
	}

	tmpl, err := loadTemplatesCached()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	err = os.MkdirAll("out/weather", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output weather directory: %w", err)
	}

	for _, weatherInQuestion := range weathers {
		outputPath := createWeatherOutputPathFromTitle(weatherInQuestion.Title)

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}
		defer file.Close()

		stories := store.GetStoriesForWeather(weatherInQuestion.Title)

		err = tmpl.ExecuteTemplate(file, "weather-index.html", data.TaxonomyIndexPage{
			Title:          weatherInQuestion.Title,
			Description:    "A list of stories for the weather " + weatherInQuestion.Title,
			Stories:        stories,
			TaxonomyColour: weatherInQuestion.Colour,
			RandomStoryURL: randomStoryURL(allStories),
			StoriesJSON:    storiesJSON,
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

	}

	log.Printf("Weather index generation complete: %d pages", len(weathers))
	return nil
}

// WriteTypesIndexes generates the type index pages and writes them to the out/types directory.
func WriteTypesIndexes() error {
	log.Println("Starting types generation")
	types := store.GetTypes()
	allStories := getAllStories()

	// Convert stories to JSON
	storiesJSON, err := convertStoriesToJSON(allStories)
	if err != nil {
		return err
	}

	tmpl, err := loadTemplatesCached()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	err = os.MkdirAll("out/types", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output types directory: %w", err)
	}

	for _, typeInQuestion := range types {
		outputPath := createTypeIndexOutputPathFromTitle(typeInQuestion.Title)

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}
		defer file.Close()

		stories := store.GetStoriesForType(typeInQuestion.Title)

		err = tmpl.ExecuteTemplate(file, "type-index.html", data.TaxonomyIndexPage{
			Title:          typeInQuestion.Title,
			Description:    "A list of stories for the type " + typeInQuestion.Title,
			Stories:        stories,
			TaxonomyColour: typeInQuestion.Colour,
			RandomStoryURL: randomStoryURL(allStories),
			StoriesJSON:    storiesJSON,
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

	}

	log.Printf("Type index generation complete: %d pages", len(types))
	return nil
}

// WriteThemesIndexes generates the theme index pages and writes them to the out/themes directory.
func WriteThemesIndexes() error {
	log.Println("Starting themes generation")
	themes := store.GetThemes()
	allStories := getAllStories()

	// Convert stories to JSON
	storiesJSON, err := convertStoriesToJSON(allStories)
	if err != nil {
		return err
	}

	tmpl, err := loadTemplatesCached()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	err = os.MkdirAll("out/themes", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output themes directory: %w", err)
	}

	for _, themeInQuestion := range themes {
		outputPath := createThemeIndexOutputPathFromTitle(themeInQuestion.Title)

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}
		defer file.Close()

		stories := store.GetStoriesForTheme(themeInQuestion.Title)

		err = tmpl.ExecuteTemplate(file, "theme-index.html", data.TaxonomyIndexPage{
			Title:          themeInQuestion.Title,
			Description:    "A list of stories for the theme " + themeInQuestion.Title,
			Stories:        stories,
			TaxonomyColour: themeInQuestion.Colour,
			RandomStoryURL: randomStoryURL(allStories),
			StoriesJSON:    storiesJSON,
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

	}

	log.Printf("Theme index generation complete: %d pages", len(themes))
	return nil
}

// WriteGiftedByIndexPages generates the gifted by index pages and writes them to the out/giftedby directory.
func WriteGiftedByIndexPages() error {
	log.Println("Starting gifted by generation")
	giftedByTypes := store.GetGiftedByTypes()
	allStories := getAllStories()

	// Convert stories to JSON
	storiesJSON, err := convertStoriesToJSON(allStories)
	if err != nil {
		return err
	}

	tmpl, err := loadTemplatesCached()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	err = os.MkdirAll("out/giftedby", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output giftedby directory: %w", err)
	}

	for _, giftedByInQuestion := range giftedByTypes {
		outputPath := createGiftedByOutputPathFromTitle(giftedByInQuestion.Title)

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}
		defer file.Close()

		stories := store.GetStoriesForGiftedBy(giftedByInQuestion.Title)

		err = tmpl.ExecuteTemplate(file, "giftedby-index.html", data.TaxonomyIndexPage{
			Title:          giftedByInQuestion.Title,
			Description:    "A list of stories gifted or co-created by " + giftedByInQuestion.Title,
			Stories:        stories,
			TaxonomyColour: giftedByInQuestion.Colour,
			RandomStoryURL: randomStoryURL(allStories),
			StoriesJSON:    storiesJSON,
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

	}

	log.Printf("Gifted-by index generation complete: %d pages", len(giftedByTypes))
	return nil
}

// WriteScalePermanenceIndexPages generates the scale permanence index pages and writes them to the out/scalepermanence directory.
func WriteScalePermanenceIndexPages() error {
	log.Println("Starting scale permanence generation")
	scalePermanenceTypes := store.GetScalePermanenceTypes()
	allStories := getAllStories()

	// Convert stories to JSON
	storiesJSON, err := convertStoriesToJSON(allStories)
	if err != nil {
		return err
	}

	tmpl, err := loadTemplatesCached()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	err = os.MkdirAll("out/scalepermanence", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output scalepermanence directory: %w", err)
	}

	for _, scalePermanenceInQuestion := range scalePermanenceTypes {
		outputPath := createScalePermanenceOutputPathFromTitle(scalePermanenceInQuestion.Title)

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}
		defer file.Close()

		stories := store.GetStoriesForScalePermanence(scalePermanenceInQuestion.Title)

		err = tmpl.ExecuteTemplate(file, "scalepermanence-index.html", data.TaxonomyIndexPage{
			Title:          scalePermanenceInQuestion.Title,
			Description:    "A list of stories with scale of permanence " + scalePermanenceInQuestion.Title,
			Stories:        stories,
			TaxonomyColour: scalePermanenceInQuestion.Colour,
			RandomStoryURL: randomStoryURL(allStories),
			StoriesJSON:    storiesJSON,
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

	}

	log.Printf("Scale permanence index generation complete: %d pages", len(scalePermanenceTypes))
	return nil
}

// WriteWhatWasIsIfIndexPages generates the what was/is/if index pages and writes them to the out/whatwasisif directory.
func WriteWhatWasIsIfIndexPages() error {
	log.Println("Starting what was/is/if generation")
	whatWasIsIfTypes := store.GetWhatWasIsIfTypes()
	allStories := getAllStories()

	// Convert stories to JSON
	storiesJSON, err := convertStoriesToJSON(allStories)
	if err != nil {
		return err
	}

	tmpl, err := loadTemplatesCached()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	err = os.MkdirAll("out/whatwasisif", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output whatwasisif directory: %w", err)
	}

	for _, whatWasIsIfInQuestion := range whatWasIsIfTypes {
		outputPath := createWhatWasIsIfOutputPathFromTitle(whatWasIsIfInQuestion.Title)

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}
		defer file.Close()

		stories := store.GetStoriesForWhatWasIsIf(whatWasIsIfInQuestion.Title)

		err = tmpl.ExecuteTemplate(file, "whatwasisif-index.html", data.TaxonomyIndexPage{
			Title:          whatWasIsIfInQuestion.Title,
			Description:    "A list of stories for " + whatWasIsIfInQuestion.Title,
			Stories:        stories,
			TaxonomyColour: whatWasIsIfInQuestion.Colour,
			RandomStoryURL: randomStoryURL(allStories),
			StoriesJSON:    storiesJSON,
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

	}

	log.Printf("What-was-is-if index generation complete: %d pages", len(whatWasIsIfTypes))
	return nil
}

// WriteTimePeriodIndexPages generates the time period index pages and writes them to the out/timeperiod directory.
func WriteTimePeriodIndexPages() error {
	log.Println("Starting time period generation")
	timePeriodTypes := store.GetTimePeriodTypes()
	allStories := getAllStories()

	// Convert stories to JSON
	storiesJSON, err := convertStoriesToJSON(allStories)
	if err != nil {
		return err
	}

	tmpl, err := loadTemplatesCached()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	err = os.MkdirAll("out/timeperiod", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output timeperiod directory: %w", err)
	}

	for _, timePeriodInQuestion := range timePeriodTypes {
		outputPath := createTimePeriodOutputPathFromTitle(timePeriodInQuestion.Title)

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}
		defer file.Close()

		stories := store.GetStoriesForTimePeriod(timePeriodInQuestion.Title)

		err = tmpl.ExecuteTemplate(file, "timeperiod-index.html", data.TaxonomyIndexPage{
			Title:          timePeriodInQuestion.Title,
			Description:    "A list of stories from time period " + timePeriodInQuestion.Title,
			Stories:        stories,
			TaxonomyColour: timePeriodInQuestion.Colour,
			RandomStoryURL: randomStoryURL(allStories),
			StoriesJSON:    storiesJSON,
		})

		if err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

	}

	log.Printf("Time-period index generation complete: %d pages", len(timePeriodTypes))
	return nil
}
