// Retrieves and processes themes as required.
package store

import (
	"log"

	"community-climate-justice-archive/data"
)

// GetThemes retrieves all themes from the themes directory and returns them as a slice of Theme.
// Intended for passing to HTML templates.
// For the moment this is hardcoded, but we will use SQLite to populate this.
func GetThemes() []data.Theme {
	log.Println("Getting themes")

	return []data.Theme{
		{
			Title: "people",
		},
		{
			Title: "planet",
		},
		{
			Title: "architecture",
		},
		{
			Title: "old",
		},
	}
}
