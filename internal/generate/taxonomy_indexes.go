package generate

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/store"
)

type taxonomyIndexConfig[T any] struct {
	label       string
	outputDir   string
	template    string
	description func(string) string
	list        func() []T
	storiesFor  func(string) []data.Story
	title       func(T) string
	color       func(T) string
}

func writeTaxonomyIndexPages[T any](cfg taxonomyIndexConfig[T]) error {
	log.Printf("Starting %s generation", cfg.label)
	items := cfg.list()
	allStories := getAllStories()
	randomURL := randomStoryURL(allStories)

	storiesJSON, err := convertStoriesToJSON(allStories)
	if err != nil {
		return err
	}

	tmpl, err := loadTemplatesCached()
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	outputDirectory := filepath.Join("out", cfg.outputDir)
	if err := os.MkdirAll(outputDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create output %s directory: %w", cfg.outputDir, err)
	}

	for _, item := range items {
		title := cfg.title(item)
		outputPath := createTaxonomyOutputPathFromTitle(cfg.outputDir, title)
		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}

		renderErr := tmpl.ExecuteTemplate(file, cfg.template, data.TaxonomyIndexPage{
			Title:          title,
			Description:    cfg.description(title),
			Stories:        cfg.storiesFor(title),
			TaxonomyColour: cfg.color(item),
			RandomStoryURL: randomURL,
			StoriesJSON:    storiesJSON,
		})
		closeErr := file.Close()

		if renderErr != nil {
			return fmt.Errorf("failed to execute %s template for %q: %w", cfg.label, title, renderErr)
		}
		if closeErr != nil {
			return fmt.Errorf("failed to close output file %s: %w", outputPath, closeErr)
		}
	}

	log.Printf("%s generation complete: %d pages", cfg.label, len(items))
	return nil
}

// WriteWeatherIndexes generates the weather index pages and writes them to the out/weather directory.
func WriteWeatherIndexes() error {
	return writeTaxonomyIndexPages(taxonomyIndexConfig[data.Weather]{
		label:       "weather index",
		outputDir:   "weather",
		template:    "weather-index.html",
		description: func(title string) string { return "A list of stories for the weather " + title },
		list:        store.GetWeather,
		storiesFor:  store.GetStoriesForWeather,
		title:       func(weather data.Weather) string { return weather.Title },
		color:       func(weather data.Weather) string { return weather.Colour },
	})
}

// WriteTypesIndexes generates the type index pages and writes them to the out/types directory.
func WriteTypesIndexes() error {
	return writeTaxonomyIndexPages(taxonomyIndexConfig[data.Type]{
		label:       "type index",
		outputDir:   "types",
		template:    "type-index.html",
		description: func(title string) string { return "A list of stories for the type " + title },
		list:        store.GetTypes,
		storiesFor:  store.GetStoriesForType,
		title:       func(typ data.Type) string { return typ.Title },
		color:       func(typ data.Type) string { return typ.Colour },
	})
}

// WriteThemesIndexes generates the theme index pages and writes them to the out/themes directory.
func WriteThemesIndexes() error {
	return writeTaxonomyIndexPages(taxonomyIndexConfig[data.Theme]{
		label:       "theme index",
		outputDir:   "themes",
		template:    "theme-index.html",
		description: func(title string) string { return "A list of stories for the theme " + title },
		list:        store.GetThemes,
		storiesFor:  store.GetStoriesForTheme,
		title:       func(theme data.Theme) string { return theme.Title },
		color:       func(theme data.Theme) string { return theme.Colour },
	})
}

// WriteGiftedByIndexPages generates the gifted by index pages and writes them to the out/giftedby directory.
func WriteGiftedByIndexPages() error {
	return writeTaxonomyIndexPages(taxonomyIndexConfig[data.GiftedBy]{
		label:       "gifted-by index",
		outputDir:   "giftedby",
		template:    "giftedby-index.html",
		description: func(title string) string { return "A list of stories gifted or co-created by " + title },
		list:        store.GetGiftedByTypes,
		storiesFor:  store.GetStoriesForGiftedBy,
		title:       func(giftedBy data.GiftedBy) string { return giftedBy.Title },
		color:       func(giftedBy data.GiftedBy) string { return giftedBy.Colour },
	})
}

// WriteScalePermanenceIndexPages generates the scale permanence index pages and writes them to the out/scalepermanence directory.
func WriteScalePermanenceIndexPages() error {
	return writeTaxonomyIndexPages(taxonomyIndexConfig[data.ScalePermanence]{
		label:       "scale permanence index",
		outputDir:   "scalepermanence",
		template:    "scalepermanence-index.html",
		description: func(title string) string { return "A list of stories with scale of permanence " + title },
		list:        store.GetScalePermanenceTypes,
		storiesFor:  store.GetStoriesForScalePermanence,
		title:       func(scalePermanence data.ScalePermanence) string { return scalePermanence.Title },
		color:       func(scalePermanence data.ScalePermanence) string { return scalePermanence.Colour },
	})
}

// WriteWhatWasIsIfIndexPages generates the what was/is/if index pages and writes them to the out/whatwasisif directory.
func WriteWhatWasIsIfIndexPages() error {
	return writeTaxonomyIndexPages(taxonomyIndexConfig[data.WhatWasIsIf]{
		label:       "what-was-is-if index",
		outputDir:   "whatwasisif",
		template:    "whatwasisif-index.html",
		description: func(title string) string { return "A list of stories for " + title },
		list:        store.GetWhatWasIsIfTypes,
		storiesFor:  store.GetStoriesForWhatWasIsIf,
		title:       func(whatWasIsIf data.WhatWasIsIf) string { return whatWasIsIf.Title },
		color:       func(whatWasIsIf data.WhatWasIsIf) string { return whatWasIsIf.Colour },
	})
}

// WriteTimePeriodIndexPages generates the time period index pages and writes them to the out/timeperiod directory.
func WriteTimePeriodIndexPages() error {
	return writeTaxonomyIndexPages(taxonomyIndexConfig[data.TimePeriod]{
		label:       "time-period index",
		outputDir:   "timeperiod",
		template:    "timeperiod-index.html",
		description: func(title string) string { return "A list of stories from time period " + title },
		list:        store.GetTimePeriodTypes,
		storiesFor:  store.GetStoriesForTimePeriod,
		title:       func(timePeriod data.TimePeriod) string { return timePeriod.Title },
		color:       func(timePeriod data.TimePeriod) string { return timePeriod.Colour },
	})
}
