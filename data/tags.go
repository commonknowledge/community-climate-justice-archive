// Package data contains the types for organising stories.
//
// This file is about "tags" - all the different ways you can browse and filter
// stories in the archive. Think of tags like the sections in a library: you might
// browse by topic (themes), by format (types), or by weather.
//
// Tag types:
// - Theme: Topics like "Climate Change", "Community", "Nature"
// - Type: Formats like "Photo", "Poem", "Video", "Drawing"
// - Weather: Conditions like "Sunny", "Rainy", "Cloudy"
// - Thumbnail: Preview images (used in some templates)
//
// Each tag has:
// - Title: What it's called
// - URL: Where to find the page listing all stories with this tag
// - Colour: A consistent colour generated from the name
//
// Stories can have multiple tags of each type, so you might find the same story
// when browsing "Climate Change" and also when browsing "Photos". These tags are
// how people explore the archive and discover connections.
package data

// Theme is a topic or subject that a story is about.
//
// Themes are things like "Climate Change", "Community", "Nature" - the big topics
// that stories touch on. Each story can have multiple themes, so you might find
// the same story when browsing different theme pages.
//
// Each theme has its own colour that's generated from its name, so "Climate Change"
// will always be the same colour wherever you see it in the archive.
type Theme struct {
	Title  string // What the theme is called (like "Climate Change")
	URL    string // Where to find the page listing all stories for this theme
	Colour string // The colour for this theme (like "#a3c4f2")
}

// Type is what format a story is in.
//
// Is it a photo? A poem? A video? A drawing? That's what the type tells you.
// This helps people find the kind of stories they want to look at - some might
// prefer photos, others might like reading poems.
type Type struct {
	Title  string // What kind of story it is (like "Photo" or "Poem")
	URL    string // Where to find all stories of this type
	Colour string // The colour for this type
}

// Weather is what the weather was like when the story happened.
//
// Was it sunny? Rainy? Foggy? The weather adds a bit of atmosphere and context
// to stories. It's also quite a nice way to browse - you can see all the stories
// that happened in the rain, for example.
type Weather struct {
	Title  string // What the weather was (like "Sunny" or "Rainy")
	URL    string // Where to find all stories with this weather
	Colour string // The colour for this weather condition
}

// Thumbnail is a preview image with its dimensions.
//
// Some templates use thumbnails - small preview versions of images. This struct
// holds the info about them: where to find the thumbnail and how big it is.
type Thumbnail struct {
	URL    string // Where to find the thumbnail image
	Width  int    // How wide it is in pixels
	Height int    // How tall it is in pixels
	Colour string // The colour associated with it
}
