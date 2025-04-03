// The archive command builds the archive and optionally serves it in development mode.
package main

import (
	"bufio"
	"community-climate-justice-archive/internal/generate"
	"community-climate-justice-archive/internal/server"
	"flag"
	"fmt"
	"log"
	"os"
)

// regenerate builds the archive by taking the following steps:
// - Getting the data from the database
// - Getting the images from the images directory
// - Adding this data to the templates to create pages which are static HTML files, intially the homepage
// - Copying the images to the output directory
func regenerate() error {
	log.Println("Starting build process")

	if err := generate.WriteHomePage(); err != nil {
		return fmt.Errorf("failed to write homepage: %v", err)
	}

	if err := generate.CopyImagesToOutput(); err != nil {
		return fmt.Errorf("failed to copy images: %v", err)
	}

	log.Println("Build process completed successfully")
	return nil
}

// waitForInput waits for input and then rebuilds the archive when enter is pressed.
func waitForInput() {
	log.Println("Press enter to regenerate the archive...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadRune()
}

// main builds the archive and optionally serves it in development mode.
func main() {
	devMode := flag.Bool("development", false, "Run in development mode with live reload")
	flag.BoolVar(devMode, "d", false, "Run in development mode with live reload (shorthand)")
	flag.Parse()

	// Build the archive.
	if err := regenerate(); err != nil {
		log.Fatalf("Initial build failed: %v", err)
	}

	if *devMode {
		log.Println("Starting development server...")
		// Start a simple HTTP server to serve the archive on port 8080.
		// The "go" keyword is used to run the server in a separate "goroutine".
		// A goroutine is a lightweight way of running a function in parallel with the main program.
		// This allows the server to run concurrently with the main program.
		// For more information see: https://go.dev/doc/effective_go#goroutines
		go server.Serve()

		log.Println("Development server running at http://localhost:8080")

		// Wait for input and then rebuild the archive when enter is pressed.
		for {
			waitForInput()
			if err := regenerate(); err != nil {
				log.Printf("Regeneration failed: %v", err)
			}
		}
	}
}
