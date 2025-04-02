package main

import (
	"fmt"
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

	return tmpl.Execute(file, page)
}

func main() {
	err := writePage()
	if err != nil {
		fmt.Println(err)
	}
}
