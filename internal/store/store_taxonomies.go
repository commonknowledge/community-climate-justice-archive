package store

import (
	"sort"
	"strings"

	"community-climate-justice-archive/data"
)

// getStoriesForTaxonomy returns every story tagged with the requested taxonomy title.
func getStoriesForTaxonomy[T any](taxonomyTitle string, extract func(data.Story) []T, titleOf func(T) string) []data.Story {
	allStories := GetAllStories()

	var stories []data.Story
	for _, story := range allStories {
		for _, value := range extract(story) {
			if titleOf(value) == taxonomyTitle {
				stories = append(stories, story)
				break
			}
		}
	}

	return stories
}

// getUniqueTaxonomies collects and alphabetizes the unique taxonomy values in the archive.
func getUniqueTaxonomies[T any](extract func(data.Story) []T, titleOf func(T) string) []T {
	allStories := GetAllStories()

	valuesByTitle := make(map[string]T)
	for _, story := range allStories {
		for _, value := range extract(story) {
			title := titleOf(value)
			if title != "" {
				valuesByTitle[title] = value
			}
		}
	}

	values := make([]T, 0, len(valuesByTitle))
	for _, value := range valuesByTitle {
		values = append(values, value)
	}

	sort.Slice(values, func(i, j int) bool {
		return titleOf(values[i]) < titleOf(values[j])
	})

	return values
}

// GetStoriesForTheme finds all stories tagged with a particular theme.
func GetStoriesForTheme(themeTitle string) []data.Story {
	return getStoriesForTaxonomy(
		themeTitle,
		func(story data.Story) []data.Theme { return story.Themes },
		func(theme data.Theme) string { return theme.Title },
	)
}

// GetThemes collects all the unique themes from across the archive.
func GetThemes() []data.Theme {
	return getUniqueTaxonomies(
		func(story data.Story) []data.Theme { return story.Themes },
		func(theme data.Theme) string { return theme.Title },
	)
}

// GetStoriesForType finds all stories of a particular type.
func GetStoriesForType(typeTitle string) []data.Story {
	return getStoriesForTaxonomy(
		typeTitle,
		func(story data.Story) []data.Type { return story.Type },
		func(typ data.Type) string { return typ.Title },
	)
}

// GetTypes collects all the unique types from across the archive.
func GetTypes() []data.Type {
	return getUniqueTaxonomies(
		func(story data.Story) []data.Type { return story.Type },
		func(typ data.Type) string { return typ.Title },
	)
}

// GetStoriesForWeather finds all stories tagged with a particular weather condition.
func GetStoriesForWeather(weatherTitle string) []data.Story {
	return getStoriesForTaxonomy(
		weatherTitle,
		func(story data.Story) []data.Weather { return story.Weather },
		func(weather data.Weather) string { return weather.Title },
	)
}

// GetWeather collects all the unique weather conditions from across the archive.
func GetWeather() []data.Weather {
	return getUniqueTaxonomies(
		func(story data.Story) []data.Weather { return story.Weather },
		func(weather data.Weather) string { return weather.Title },
	)
}

// GetStoriesForProject finds all stories linked to one project name.
//
// Unlike themes or weather, this value is stored as one plain string field on the story,
// so we wrap it in a one-item slice to reuse the generic taxonomy helper.
func GetStoriesForProject(projectTitle string) []data.Story {
	return getStoriesForTaxonomy(
		projectTitle,
		func(story data.Story) []string {
			title := strings.TrimSpace(story.HighStExperiment)
			if title == "" {
				return nil
			}
			return []string{title}
		},
		func(project string) string { return project },
	)
}

// GetProjectTypes collects unique project names from the `HighStExperiment` field.
func GetProjectTypes() []string {
	return getUniqueTaxonomies(
		func(story data.Story) []string {
			title := strings.TrimSpace(story.HighStExperiment)
			if title == "" {
				return nil
			}
			return []string{title}
		},
		func(project string) string { return project },
	)
}

// GetStoriesForGiftedBy finds all stories from a particular contributor.
func GetStoriesForGiftedBy(giftedByTitle string) []data.Story {
	return getStoriesForTaxonomy(
		giftedByTitle,
		func(story data.Story) []data.GiftedBy { return story.GiftedBy },
		func(giftedBy data.GiftedBy) string { return giftedBy.Title },
	)
}

// GetGiftedByTypes collects all the unique contributors from across the archive.
func GetGiftedByTypes() []data.GiftedBy {
	return getUniqueTaxonomies(
		func(story data.Story) []data.GiftedBy { return story.GiftedBy },
		func(giftedBy data.GiftedBy) string { return giftedBy.Title },
	)
}

// GetStoriesForScalePermanence finds all stories with a particular permanence level.
func GetStoriesForScalePermanence(scalePermanenceTitle string) []data.Story {
	return getStoriesForTaxonomy(
		scalePermanenceTitle,
		func(story data.Story) []data.ScalePermanence { return story.ScalePermanence },
		func(scalePermanence data.ScalePermanence) string { return scalePermanence.Title },
	)
}

// GetScalePermanenceTypes collects all the unique permanence levels from the archive.
func GetScalePermanenceTypes() []data.ScalePermanence {
	return getUniqueTaxonomies(
		func(story data.Story) []data.ScalePermanence { return story.ScalePermanence },
		func(scalePermanence data.ScalePermanence) string { return scalePermanence.Title },
	)
}

// GetStoriesForWhatWasIsIf finds all stories with a particular temporal perspective.
func GetStoriesForWhatWasIsIf(whatWasIsIfTitle string) []data.Story {
	return getStoriesForTaxonomy(
		whatWasIsIfTitle,
		func(story data.Story) []data.WhatWasIsIf { return story.WhatWasIsIf },
		func(whatWasIsIf data.WhatWasIsIf) string { return whatWasIsIf.Title },
	)
}

// GetWhatWasIsIfTypes collects all the unique temporal perspectives from the archive.
func GetWhatWasIsIfTypes() []data.WhatWasIsIf {
	return getUniqueTaxonomies(
		func(story data.Story) []data.WhatWasIsIf { return story.WhatWasIsIf },
		func(whatWasIsIf data.WhatWasIsIf) string { return whatWasIsIf.Title },
	)
}

// GetStoriesForTimePeriod finds all stories from a particular era.
func GetStoriesForTimePeriod(timePeriodTitle string) []data.Story {
	return getStoriesForTaxonomy(
		timePeriodTitle,
		func(story data.Story) []data.TimePeriod { return story.TimePeriod },
		func(timePeriod data.TimePeriod) string { return timePeriod.Title },
	)
}

// GetTimePeriodTypes collects all the unique time periods from the archive.
func GetTimePeriodTypes() []data.TimePeriod {
	return getUniqueTaxonomies(
		func(story data.Story) []data.TimePeriod { return story.TimePeriod },
		func(timePeriod data.TimePeriod) string { return timePeriod.Title },
	)
}
