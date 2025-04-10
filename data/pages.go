// Contains the data models for the pages in the archive in the data package.
package data

type Page struct {
	Title       string
	Description string
	Themes      []Theme
	Types       []Type
	Stories     []Story
}

type TaxonomyIndexPage struct {
	Title          string
	Description    string
	Stories        []Story
	TaxonomyColour string
}

type StoryPage struct {
	Title       string
	Description string
	Story       Story
	LastStory   Story
	NextStory   Story
}
