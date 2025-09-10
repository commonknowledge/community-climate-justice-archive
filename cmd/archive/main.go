// The archive command builds the archive and optionally serves it in development mode.
package main

import (
	"bufio"

	"community-climate-justice-archive/internal/config"
	"community-climate-justice-archive/internal/generate"
	"community-climate-justice-archive/internal/server"
	"community-climate-justice-archive/internal/store"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// regenerate builds the archive by taking the following steps:
// - Getting the data from the database
// - Getting the images from the images directory
// - Adding this data to the templates to create pages which are static HTML files
// - Copying the images to the output directory
func regenerate(skipImages bool) error {
	log.Println("Starting build process")

	// Commented out for now while we debug the NocoDB connection
	if !skipImages {
		if err := generate.ProcessImages(); err != nil {
			return fmt.Errorf("failed to process images: %v", err)
		}
	}

	if err := generate.WriteStories(); err != nil {
		return fmt.Errorf("failed to write stories: %v", err)
	}

	if err := generate.WriteHomePage(); err != nil {
		return fmt.Errorf("failed to write homepage: %v", err)
	}

	if err := generate.WriteWanderPage(); err != nil {
		return fmt.Errorf("failed to write wander page: %v", err)
	}

	if err := generate.WriteArchivePage(); err != nil {
		return fmt.Errorf("failed to write archive page: %v", err)
	}

	if err := generate.WriteFilterData(); err != nil {
		return fmt.Errorf("failed to write filter data: %v", err)
	}

	if err := generate.WriteTypesIndexes(); err != nil {
		return fmt.Errorf("failed to write types indexes: %v", err)
	}

	if err := generate.WriteThemesIndexes(); err != nil {
		return fmt.Errorf("failed to write themes indexes: %v", err)
	}

	if err := generate.WriteWeatherIndexes(); err != nil {
		return fmt.Errorf("failed to write weather indexes: %v", err)
	}

	if err := generate.CopyImagesToOutput(); err != nil {
		return fmt.Errorf("failed to copy images: %v", err)
	}

	if err := generate.CopyCSSToOutput(); err != nil {
		return fmt.Errorf("failed to copy CSS: %v", err)
	}

	if err := generate.CopyJSToOutput(); err != nil {
		return fmt.Errorf("failed to copy JavaScript: %v", err)
	}

	if err := generate.WriteNocoDBTestPage(); err != nil {
		return fmt.Errorf("failed to write NocoDB test page: %v", err)
	}

	log.Println("Build process completed successfully")

	return nil
}

func hotRegenerate() error {
	log.Println("Starting partial build process")

	if err := generate.WriteStories(); err != nil {
		return fmt.Errorf("failed to write stories: %v", err)
	}

	if err := generate.WriteHomePage(); err != nil {
		return fmt.Errorf("failed to write homepage: %v", err)
	}

	if err := generate.WriteWanderPage(); err != nil {
		return fmt.Errorf("failed to write wander page: %v", err)
	}

	if err := generate.WriteArchivePage(); err != nil {
		return fmt.Errorf("failed to write archive page: %v", err)
	}

	if err := generate.WriteFilterData(); err != nil {
		return fmt.Errorf("failed to write filter data: %v", err)
	}

	if err := generate.WriteTypesIndexes(); err != nil {
		return fmt.Errorf("failed to write types indexes: %v", err)
	}

	if err := generate.WriteThemesIndexes(); err != nil {
		return fmt.Errorf("failed to write themes indexes: %v", err)
	}

	if err := generate.WriteWeatherIndexes(); err != nil {
		return fmt.Errorf("failed to write weather indexes: %v", err)
	}

	if err := generate.CopyCSSToOutput(); err != nil {
		return fmt.Errorf("failed to copy CSS: %v", err)
	}

	if err := generate.CopyJSToOutput(); err != nil {
		return fmt.Errorf("failed to copy JavaScript: %v", err)
	}

	log.Println("Partial build process completed successfully")

	return nil
}

// watchCSS sets up a file watcher for the CSS directory and copies the CSS when changes are detected.
func watchCSS() (*fsnotify.Watcher, error) {
	log.Println("Setting up CSS watcher...")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %v", err)
	}

	// Create a timer to debounce frequent events
	var debounceTimer *time.Timer
	const debounceDelay = 100 * time.Millisecond

	// Set up a goroutine to handle events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("Watcher channel closed")
					return
				}

				// Only proceed if this is a CSS file
				if !strings.HasSuffix(event.Name, ".css") {
					log.Printf("Ignoring non-CSS file: %s", event.Name)
					continue
				}

				// If it's a write or create event
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					log.Printf("CSS change detected on file: %s", event.Name)
					// Reset or start the debounce timer
					if debounceTimer != nil {
						debounceTimer.Stop()
					}

					debounceTimer = time.AfterFunc(debounceDelay, func() {
						log.Printf("Copying CSS file %s to output", event.Name)
						if err := generate.CopyCSSToOutput(); err != nil {
							log.Printf("Failed to copy CSS to output: %v", err)
						} else {
							log.Println("Successfully copied CSS to output – refresh your browser to see the changes...")
						}
					})
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Println("Watcher error channel closed")
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()

	// Watch the CSS directory
	cssDir := "css"
	log.Printf("Attempting to watch directory: %s", cssDir)

	err = watcher.Add(cssDir)
	if err != nil {
		watcher.Close()
		return nil, fmt.Errorf("failed to watch CSS directory: %v", err)
	}

	log.Printf("Successfully watching CSS files in %s for changes", cssDir)
	return watcher, nil
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
	skipImages := flag.Bool("skip-images", false, "Skip image processing and generation")
	flag.BoolVar(skipImages, "s", false, "Skip image processing and generation (shorthand)")
	flag.Parse()

	// Load configuration from environment variables and .env file
	config.LoadConfig()

	// Initialize the data adapter based on configuration
	if err := store.InitializeAdapter(); err != nil {
		log.Fatalf("Failed to initialize data adapter: %v", err)
	}

	if *skipImages {
		log.Println("Skipping image processing and generation")
	}

	// Build the archive.
	if err := regenerate(*skipImages); err != nil {
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

		// Watch the CSS directory for changes and keep the watcher alive
		watcher, err := watchCSS()
		if err != nil {
			log.Printf("Failed to initialize CSS watcher: %v", err)
		} else {
			// Defer closing the watcher until the program exits
			defer watcher.Close()
		}

		log.Println("Development server running at http://localhost:8080")

		// Wait for input and then rebuild the archive when enter is pressed.
		for {
			waitForInput()
			if err := hotRegenerate(); err != nil {
				log.Printf("Hot regeneration failed: %v", err)
			}
		}
	}
}
