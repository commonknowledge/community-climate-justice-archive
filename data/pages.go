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

type Page struct {
	Title            string
	Description      string
	Themes           []Theme
	Types            []Type
	Stories          []Story
	ConnectedStories []Story
	RandomStoryURL   string
	StoriesJSON      string
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
