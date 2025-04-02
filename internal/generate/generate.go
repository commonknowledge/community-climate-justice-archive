package generate

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/store"
)

func WriteHomePage() error {
	log.Println("Starting homepage generation")
	themes := store.GetThemes()
	types := store.GetTypes()
	images := store.GetImages()

	page := data.Page{
		Title:       "Dudley People's School for Climate Justice – time portal",
		Description: "The time portal for the Dudley People's School for Climate Justice",
		Themes:      themes,
		Types:       types,
		Images:      images,
	}

	tmpl, err := template.ParseFiles("templates/homepage.html")
	if err != nil {
		return fmt.Errorf("failed to parse homepage template: %w", err)
	}

	fileName := "index.html"
	outputPath := filepath.Join("out", fileName)

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer file.Close()

	err = tmpl.Execute(file, page)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	log.Printf("Successfully wrote homepage to %s", outputPath)
	return nil
}
