// Package server runs a simple web server for testing locally.
//
// When you're working on the archive on your computer, you need to be able to see
// it in your browser. This file contains Serve(), which starts a basic web server
// that shows you the generated HTML files from the out/ folder.
//
// This is only for local development - when you're testing changes before you deploy.
// When the archive is deployed for real to Render, Render serves the generated
// static files.
//
// How to use it:
// Run the archive with the --development flag, and it'll generate everything and
// then start this server so you can view it at http://localhost:8080
package server

import (
	"log"
	"net/http"
)

// Serve starts a simple web server on port 8080 to show the generated HTML files.
func Serve() {
	err := http.ListenAndServe(":8080", http.FileServer(http.Dir("./out")))

	if err != nil {
		log.Printf("Server error: %v", err)
	}
}
