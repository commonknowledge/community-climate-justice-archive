package store

import (
	"sort"

	"community-climate-justice-archive/data"
)

// -------------------------------------------------------------------
// Themes
// -------------------------------------------------------------------

// GetStoriesForTheme finds all stories tagged with a particular theme.
//
// Themes are things like "Climate Change", "Community", "Nature" - the big topics
// that stories can be about. This loops through all stories and returns the ones
// that have the given theme in their tags.
func GetStoriesForTheme(themeTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, theme := range story.Themes {
			if theme.Title == themeTitle {
				result = append(result, story)
				break // Found it, no need to check other themes on this story
			}
		}
	}

	return result
}

// GetThemes collects all the unique themes from across the archive.
//
// This looks at every story and gathers up all the themes that appear. Each theme
// only shows up once in the results, even if dozens of stories have it.
func GetThemes() []data.Theme {
	allStories := GetAllStories()

	// Use a map to collect unique themes
	themeMap := make(map[string]data.Theme)

	for _, story := range allStories {
		for _, theme := range story.Themes {
			if theme.Title != "" {
				themeMap[theme.Title] = theme
			}
		}
	}

	// Turn the map into a list
	var themes []data.Theme
	for _, theme := range themeMap {
		themes = append(themes, theme)
	}

	// Sort alphabetically by title
	sort.Slice(themes, func(i, j int) bool {
		return themes[i].Title < themes[j].Title
	})

	return themes
}

// -------------------------------------------------------------------
// Types
// -------------------------------------------------------------------

// GetStoriesForType finds all stories of a particular type.
//
// Types describe what form the story takes - "Photo", "Poem", "Video", "Drawing",
// "Text", and so on. A story can have multiple types (like a photo with a poem).
func GetStoriesForType(typeTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, typ := range story.Type {
			if typ.Title == typeTitle {
				result = append(result, story)
				break
			}
		}
	}

	return result
}

// GetTypes collects all the unique types from across the archive.
func GetTypes() []data.Type {
	allStories := GetAllStories()

	typeMap := make(map[string]data.Type)

	for _, story := range allStories {
		for _, typ := range story.Type {
			if typ.Title != "" {
				typeMap[typ.Title] = typ
			}
		}
	}

	var types []data.Type
	for _, typ := range typeMap {
		types = append(types, typ)
	}

	// Sort alphabetically by title
	sort.Slice(types, func(i, j int) bool {
		return types[i].Title < types[j].Title
	})

	return types
}

// -------------------------------------------------------------------
// Weather
// -------------------------------------------------------------------

// GetStoriesForWeather finds all stories tagged with a particular weather condition.
//
// Weather is a lovely way to browse the archive - was it sunny when this story
// happened? Rainy? Foggy? It adds an atmospheric dimension to exploring.
func GetStoriesForWeather(weatherTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, weather := range story.Weather {
			if weather.Title == weatherTitle {
				result = append(result, story)
				break
			}
		}
	}

	return result
}

// GetWeather collects all the unique weather conditions from across the archive.
func GetWeather() []data.Weather {
	allStories := GetAllStories()

	weatherMap := make(map[string]data.Weather)

	for _, story := range allStories {
		for _, weather := range story.Weather {
			if weather.Title != "" {
				weatherMap[weather.Title] = weather
			}
		}
	}

	var weather []data.Weather
	for _, w := range weatherMap {
		weather = append(weather, w)
	}

	// Sort alphabetically by title
	sort.Slice(weather, func(i, j int) bool {
		return weather[i].Title < weather[j].Title
	})

	return weather
}

// -------------------------------------------------------------------
// GiftedBy (Contributors)
// -------------------------------------------------------------------

// GetStoriesForGiftedBy finds all stories from a particular contributor.
//
// "Gifted by" tracks who shared or co-created each story - local schools,
// community groups, individuals. It's a nice way to celebrate everyone
// who's contributed to the archive.
func GetStoriesForGiftedBy(giftedByTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, giftedBy := range story.GiftedBy {
			if giftedBy.Title == giftedByTitle {
				result = append(result, story)
				break
			}
		}
	}

	return result
}

// GetGiftedByTypes collects all the unique contributors from across the archive.
func GetGiftedByTypes() []data.GiftedBy {
	allStories := GetAllStories()

	giftedByMap := make(map[string]data.GiftedBy)

	for _, story := range allStories {
		for _, giftedBy := range story.GiftedBy {
			if giftedBy.Title != "" {
				giftedByMap[giftedBy.Title] = giftedBy
			}
		}
	}

	var giftedByTypes []data.GiftedBy
	for _, giftedBy := range giftedByMap {
		giftedByTypes = append(giftedByTypes, giftedBy)
	}

	// Sort alphabetically by title
	sort.Slice(giftedByTypes, func(i, j int) bool {
		return giftedByTypes[i].Title < giftedByTypes[j].Title
	})

	return giftedByTypes
}

// -------------------------------------------------------------------
// Scale of Permanence
// -------------------------------------------------------------------

// GetStoriesForScalePermanence finds all stories with a particular permanence level.
//
// Scale of permanence comes from permaculture - it's about how long-lasting things
// are, from temporary to permanent. It's an interesting lens for thinking about
// the stories in the archive.
func GetStoriesForScalePermanence(scalePermanenceTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, sp := range story.ScalePermanence {
			if sp.Title == scalePermanenceTitle {
				result = append(result, story)
				break
			}
		}
	}

	return result
}

// GetScalePermanenceTypes collects all the unique permanence levels from the archive.
func GetScalePermanenceTypes() []data.ScalePermanence {
	allStories := GetAllStories()

	spMap := make(map[string]data.ScalePermanence)

	for _, story := range allStories {
		for _, sp := range story.ScalePermanence {
			if sp.Title != "" {
				spMap[sp.Title] = sp
			}
		}
	}

	var scalePermanenceTypes []data.ScalePermanence
	for _, sp := range spMap {
		scalePermanenceTypes = append(scalePermanenceTypes, sp)
	}

	// Sort alphabetically by title
	sort.Slice(scalePermanenceTypes, func(i, j int) bool {
		return scalePermanenceTypes[i].Title < scalePermanenceTypes[j].Title
	})

	return scalePermanenceTypes
}

// -------------------------------------------------------------------
// What Was/Is/If (Temporal Perspective)
// -------------------------------------------------------------------

// GetStoriesForWhatWasIsIf finds all stories with a particular temporal perspective.
//
// "What Was" is about the past, "What Is" about the present, "What If" about
// imagined futures. It's a lovely way to think about how stories relate to time.
func GetStoriesForWhatWasIsIf(whatWasIsIfTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, wwii := range story.WhatWasIsIf {
			if wwii.Title == whatWasIsIfTitle {
				result = append(result, story)
				break
			}
		}
	}

	return result
}

// GetWhatWasIsIfTypes collects all the unique temporal perspectives from the archive.
func GetWhatWasIsIfTypes() []data.WhatWasIsIf {
	allStories := GetAllStories()

	wwiiMap := make(map[string]data.WhatWasIsIf)

	for _, story := range allStories {
		for _, wwii := range story.WhatWasIsIf {
			if wwii.Title != "" {
				wwiiMap[wwii.Title] = wwii
			}
		}
	}

	var whatWasIsIfTypes []data.WhatWasIsIf
	for _, wwii := range wwiiMap {
		whatWasIsIfTypes = append(whatWasIsIfTypes, wwii)
	}

	// Sort alphabetically by title
	sort.Slice(whatWasIsIfTypes, func(i, j int) bool {
		return whatWasIsIfTypes[i].Title < whatWasIsIfTypes[j].Title
	})

	return whatWasIsIfTypes
}

// -------------------------------------------------------------------
// Time Period
// -------------------------------------------------------------------

// GetStoriesForTimePeriod finds all stories from a particular era.
//
// Time periods might be things like "1960s", "Victorian Era", "Present Day" -
// whenever the story relates to. It helps people explore stories from particular
// moments in history.
func GetStoriesForTimePeriod(timePeriodTitle string) []data.Story {
	allStories := GetAllStories()

	var result []data.Story
	for _, story := range allStories {
		for _, tp := range story.TimePeriod {
			if tp.Title == timePeriodTitle {
				result = append(result, story)
				break
			}
		}
	}

	return result
}

// GetTimePeriodTypes collects all the unique time periods from the archive.
func GetTimePeriodTypes() []data.TimePeriod {
	allStories := GetAllStories()

	tpMap := make(map[string]data.TimePeriod)

	for _, story := range allStories {
		for _, tp := range story.TimePeriod {
			if tp.Title != "" {
				tpMap[tp.Title] = tp
			}
		}
	}

	var timePeriodTypes []data.TimePeriod
	for _, tp := range tpMap {
		timePeriodTypes = append(timePeriodTypes, tp)
	}

	// Sort alphabetically by title
	sort.Slice(timePeriodTypes, func(i, j int) bool {
		return timePeriodTypes[i].Title < timePeriodTypes[j].Title
	})

	return timePeriodTypes
}
