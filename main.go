package main

import (
	"html/template"
	"log"
	"os"
)

type Page struct {
	Title       string
	Description string
}

func writePage() error {
	page := Page{
		Title:       "Dudley People's School for Climate Justice – time portal",
		Description: "The time portal for the Dudley People's School for Climate Justice",
	}

	tmpl, err := template.ParseFiles("templates/homepage.html")
	if err != nil {
		log.Fatal("error parsing homepage template:", err)
	}

	fileName := "index.html"

	file, err := os.Create("out/" + fileName)
	if err != nil {
		log.Fatal("error creating homepage output file:", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, page)
	if err != nil {
		log.Fatal("error executing homepage template:", err)
	}

	log.Println("homepage output file written successfully")

	return nil
}

func main() {
	writePage()
}
