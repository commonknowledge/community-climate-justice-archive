// Retrieves and processes types as required.
package store

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"community-climate-justice-archive/data"
)

// GetTypes retrieves all types from the database and returns them as a slice of Type.
// Intended for passing to HTML templates.
func GetTypes() []data.Type {
	log.Println("Getting types")

	dbPath := "airtable-export.db"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT Themes type FROM stories")
	if err != nil {
		log.Fatalf("Failed to query types: %v", err)
	}
	defer rows.Close()

	var types []data.Type

	for rows.Next() {
		var (
			Theme sql.NullString
		)
		if err := rows.Scan(&Theme); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		if Theme.Valid {
			// First unmarshal into a string array since it's in format ["Tiny Things", "Care"]
			var themeStrings []string
			if err := json.Unmarshal([]byte(Theme.String), &themeStrings); err != nil {
				log.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			for _, themeStr := range themeStrings {
				newType := data.Type{Title: themeStr, URL: strings.ToLower(themeStr)}
				types = append(types, newType)
			}
		}
	}

	log.Printf("Found %d types", len(types))

	types = uniqueTypes(types)
	log.Printf("Found %d unique types", len(types))

	return types
}

// uniqueTypes returns a slice of unique types
func uniqueTypes(types []data.Type) []data.Type {
	seen := make(map[string]bool)
	unique := []data.Type{}

	// Loop over the slice and only keep first occurrence
	for _, t := range types {
		if !seen[t.Title] {
			seen[t.Title] = true
			unique = append(unique, t)
		}
	}

	return unique
}
