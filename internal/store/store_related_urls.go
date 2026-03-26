package store

import (
	"fmt"
	"log"
	"math/rand"
	"path/filepath"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/util"
)

// -------------------------------------------------------------------
// URL Helpers
// -------------------------------------------------------------------

// CreateStoryURLFromFinding makes a URL for a story page based on its title.
//
// It turns the title into a URL-safe slug. So "Climate Change Story" becomes
// "/stories/climate-change-story.html"
//
// Note: If two stories happen to have identical titles, they'd get the same URL,
// which would be a problem. Use CreateStoryURLFromFindingWithID instead to be safe.
func CreateStoryURLFromFinding(finding string) string {
	slug := util.Slugify(finding)
	fileName := fmt.Sprintf("%s.html", slug)
	return filepath.Join("/stories", fileName)
}

// CreateStoryURLFromFindingWithID makes a unique URL by including the story's ID.
//
// So story #42 with title "Climate Change" becomes "/stories/climate-change-42.html"
//
// This is the safer option because IDs are always unique, even if two stories
// happen to have the same title.
func CreateStoryURLFromFindingWithID(finding, id string) string {
	slug := util.Slugify(finding)
	fileName := fmt.Sprintf("%s-%s.html", slug, id)
	return filepath.Join("/stories", fileName)
}

// -------------------------------------------------------------------
// Helpers for Templates
// -------------------------------------------------------------------

// GetMoreTaggedStories finds other stories with the same tag, for "More like this" sections.
//
// Given a story and one of its tags (theme or type), this returns a random selection
// of other stories that share that tag. Useful for showing related content at the
// bottom of story pages.
func GetMoreTaggedStories(story data.Story, tag interface{}, count int) []data.Story {
	var stories []data.Story

	switch t := tag.(type) {
	case data.Theme:
		stories = GetStoriesForTheme(t.Title)
	case data.Type:
		stories = GetStoriesForType(t.Title)
	default:
		log.Printf("Warning: GetMoreTaggedStories called with unsupported tag type: %T", tag)
		return []data.Story{}
	}

	// Remove the current story from the list (we don't want to suggest itself)
	var filteredStories []data.Story
	for _, s := range stories {
		if s.ID != story.ID {
			filteredStories = append(filteredStories, s)
		}
	}

	// Shuffle randomly so you get different suggestions each time
	rand.Shuffle(len(filteredStories), func(i, j int) {
		filteredStories[i], filteredStories[j] = filteredStories[j], filteredStories[i]
	})

	// Return up to 'count' stories
	if len(filteredStories) <= count {
		return filteredStories
	}
	return filteredStories[:count]
}
