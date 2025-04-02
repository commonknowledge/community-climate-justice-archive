// Retrieves and processes types as required.
package store

import (
	"log"

	"community-climate-justice-archive/data"
)

// GetTypes retrieves all types from the types directory and returns them as a slice of Type.
// Intended for passing to HTML templates.
// For the moment this is hardcoded, but we will use SQLite to populate this.
func GetTypes() []data.Type {
	log.Println("Getting types")

	return []data.Type{
		{Title: "text"},
		{Title: "image"},
		{Title: "video"},
		{Title: "sound"},
		{Title: "textile"},
	}
}
