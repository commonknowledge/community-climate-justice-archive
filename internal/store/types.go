// Retrieves and processes types from the database as required.
package store

import (
	"log"

	"community-climate-justice-archive/data"
)

// GetStoriesForType retrieves all stories for a given type from the data source.
func GetStoriesForType(typeTitle string) []data.Story {
	adapter := GetAdapter()
	stories, err := adapter.GetStoriesForType(typeTitle)
	if err != nil {
		log.Fatalf("Failed to get stories for type: %v", err)
	}
	return stories
}

// GetTypes retrieves all types from the database and returns them as a slice of Type.
// Intended for passing to HTML templates.
func GetTypes() []data.Type {
	adapter := GetAdapter()
	types, err := adapter.GetTypes()
	if err != nil {
		log.Fatalf("Failed to get types: %v", err)
	}
	return types
}

// uniqueTypes returns a slice of unique types.
func uniqueTypes(types []data.Type) []data.Type {
	seen := make(map[string]bool)
	unique := []data.Type{}

	// Loop over the slice and only keep first occurrence of each type.
	for _, t := range types {
		if !seen[t.Title] {
			seen[t.Title] = true
			unique = append(unique, t)
		}
	}

	return unique
}
