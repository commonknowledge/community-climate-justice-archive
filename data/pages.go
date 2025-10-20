// Contains the data models for the pages in the archive in the data package.
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
