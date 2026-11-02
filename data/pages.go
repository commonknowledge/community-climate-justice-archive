// Package data contains the structures we use when creating pages.
//
// When we generate HTML pages, we need to give the templates the data they should
// show. This file defines those data bundles - one struct for each type of page.
//
// Page types:
// - Page: A general page with stories
// - TaxonomyIndexPage: Pages that list stories by theme, type, weather, etc.
// - StoryPage: An individual story's page
// - RelatedStories: A group of stories that share something in common
//
// How it works:
// 1. We fetch data from the database (stories, themes, etc.)
// 2. We organize it into one of these page structures
// 3. We pass that structure to an HTML template
// 4. The template turns it into the actual HTML you see
//
// This keeps the "getting data" part separate from the "showing data" part,
// which makes everything easier to work with.
package data

import (
	"path/filepath"
	"strings"
)

type Page struct {
	Title            string
	Description      string
	HighStExperiment []string // Unique project names used to decide whether to show the archive project filter.
	Themes           []Theme
	Types            []Type
	HomepageCollages []HomepageCollage
	Weather          []Weather
	WhatWasIsIf      []WhatWasIsIf
	ScalePermanence  []ScalePermanence
	TimePeriod       []TimePeriod
	Stories          []Story
	ConnectedStories []Story
	RandomStoryURL   string
	StoriesJSON      string
}

type HomepageCollage struct {
	Filename  string // Image file name inside images/
	Title     string
	AltText   string
	StartDate string
	EndDate   string
}

// ImageURL returns the processed WebP version once image processing has run,
// otherwise the original downloaded image.
func (c HomepageCollage) ImageURL() string {
	baseName := strings.TrimSuffix(c.Filename, filepath.Ext(c.Filename))
	return processedImageURLOrOriginal("/images/processed/"+baseName+".webp", "/images/"+c.Filename)
}

type TaxonomyIndexPage struct {
	Title          string
	Description    string
	Stories        []Story
	TaxonomyColour string
	RandomStoryURL string
	StoriesJSON    string
}

type RelatedStories struct {
	Tag     interface{} // Can be Theme, Type, or Weather
	TagType string
	Stories []Story
}

type StoryPage struct {
	Title                   string
	Description             string
	Story                   Story
	Attachments             []StoryAttachment
	NocoDBURL               string
	LastStory               Story
	NextStory               Story
	FirstMoreTaggedStories  RelatedStories
	SecondMoreTaggedStories RelatedStories
	ThirdMoreTaggedStories  RelatedStories
	RandomStoryURL          string
	StoriesJSON             string
}
