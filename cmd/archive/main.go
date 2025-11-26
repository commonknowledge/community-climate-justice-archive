// The archive command builds the static website for the Dudley Climate Justice Archive.
//
// This is where everything starts! When you run this, it:
// 1. Loads settings from your environment variables
// 2. Connects to the database (NocoDB)
// 3. Processes all the images (resizing, converting to WebP)
// 4. Generates all the HTML pages from the templates
// 5. Optionally starts a local web server so you can view it
//
// You can use these flags:
// --development or -d: Run in development mode (watches for template changes)
// --skip-images or -s: Skip processing images (faster if you're just testing templates)
//
// Development mode is really handy - it watches your templates and regenerates
// pages automatically when you change them, so you can see your edits right away.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"community-climate-justice-archive/internal/config"
	"community-climate-justice-archive/internal/generate"
	"community-climate-justice-archive/internal/server"
	"community-climate-justice-archive/internal/store"

	"github.com/fsnotify/fsnotify"
)

// generateArchive builds the entire website.
//
// This does the main work:
// - Fetches all the story data from the database
// - Processes all the images (unless you've skipped them)
// - Fills in the HTML templates with the actual data
// - Writes all the HTML files to the out/ folder
// - Copies over the CSS and images
func generateArchive(skipImages bool) error {
	log.Println("Starting build process")

	// Warm the cache to ensure all subsequent operations are fast
	store.WarmCache()

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

	if err := generate.WriteGiftedByIndexPages(); err != nil {
		return fmt.Errorf("failed to write gifted by indexes: %v", err)
	}

	if err := generate.WriteScalePermanenceIndexPages(); err != nil {
		return fmt.Errorf("failed to write scale permanence indexes: %v", err)
	}

	if err := generate.WriteWhatWasIsIfIndexPages(); err != nil {
		return fmt.Errorf("failed to write what was/is/if indexes: %v", err)
	}

	if err := generate.WriteTimePeriodIndexPages(); err != nil {
		return fmt.Errorf("failed to write time period indexes: %v", err)
	}

	if err := generate.CopyImagesToOutput(); err != nil {
		return fmt.Errorf("failed to copy images: %v", err)
	}

	if err := generate.CopyAudioToOutput(); err != nil {
		return fmt.Errorf("failed to copy audio files: %v", err)
	}

	if err := generate.CopyDocumentsToOutput(); err != nil {
		return fmt.Errorf("failed to copy document files: %v", err)
	}

	if err := generate.CopyCSSToOutput(); err != nil {
		return fmt.Errorf("failed to copy CSS: %v", err)
	}

	if err := generate.CopyJSToOutput(); err != nil {
		return fmt.Errorf("failed to copy JavaScript: %v", err)
	}

	log.Println("Build process completed successfully")

	return nil
}

func hotRegenerate() error {
	log.Println("Starting partial build process")

	// Warm the cache to ensure all subsequent operations are fast
	store.WarmCache()

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

	if err := generate.WriteGiftedByIndexPages(); err != nil {
		return fmt.Errorf("failed to write gifted by indexes: %v", err)
	}

	if err := generate.WriteScalePermanenceIndexPages(); err != nil {
		return fmt.Errorf("failed to write scale permanence indexes: %v", err)
	}

	if err := generate.WriteWhatWasIsIfIndexPages(); err != nil {
		return fmt.Errorf("failed to write what was/is/if indexes: %v", err)
	}

	if err := generate.WriteTimePeriodIndexPages(); err != nil {
		return fmt.Errorf("failed to write time period indexes: %v", err)
	}

	// Copy non-image files during hot regeneration
	if err := generate.CopyAudioToOutput(); err != nil {
		return fmt.Errorf("failed to copy audio files: %v", err)
	}

	if err := generate.CopyDocumentsToOutput(); err != nil {
		return fmt.Errorf("failed to copy document files: %v", err)
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
	log.Println("Press enter to generate the archive...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadRune()
}

// dumpNocoDBData dumps raw NocoDB API response data to JSON file for debugging
func dumpNocoDBData() error {
	log.Println("Starting raw NocoDB API data dump for debugging...")

	// Load configuration from environment variables and .env file
	config.LoadConfig()

	// We need to get the raw NocoDB records directly, bypassing the story conversion
	adapter := store.GetAdapter()

	// Check if we're using NocoDB adapter
	nocodbAdapter, ok := adapter.(*store.NocoDBAdapter)
	if !ok {
		return fmt.Errorf("debug dump only works with NocoDB adapter, currently using: %T", adapter)
	}

	// Get raw records directly from NocoDB client
	rawRecords, err := nocodbAdapter.GetRawRecords()
	if err != nil {
		return fmt.Errorf("failed to get raw records from NocoDB: %v", err)
	}

	log.Printf("Retrieved %d raw records from NocoDB API for debugging dump", len(rawRecords))

	// Create debug data structure with raw NocoDB response
	debugData := struct {
		TotalRecords int                      `json:"total_records"`
		RawRecords   []map[string]interface{} `json:"raw_records"`
		DumpTime     string                   `json:"dump_time"`
		Note         string                   `json:"note"`
	}{
		TotalRecords: len(rawRecords),
		RawRecords:   rawRecords,
		DumpTime:     time.Now().Format(time.RFC3339),
		Note:         "This contains the raw NocoDB API response before any processing or conversion to Story structs",
	}

	// Write to JSON file
	file, err := os.Create("debug-raw-nocodb-data.json")
	if err != nil {
		return fmt.Errorf("failed to create debug file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(debugData); err != nil {
		return fmt.Errorf("failed to write debug data: %v", err)
	}

	log.Println("Raw NocoDB API data dump completed successfully -> debug-raw-nocodb-data.json")
	return nil
}

// generateSingleStory regenerates a single story by ID for debugging
func generateSingleStory(storyID string) error {
	log.Printf("Starting single story regeneration for ID: %s", storyID)

	// Warm cache to ensure data is available
	store.WarmCache()

	// Get the specific story
	adapter := store.GetAdapter()
	story, err := adapter.GetStoryByID(storyID)
	if err != nil {
		return fmt.Errorf("failed to get story %s: %v", storyID, err)
	}

	if story.ID == "" {
		return fmt.Errorf("story with ID %s not found", storyID)
	}

	log.Printf("Found story: %s", story.Finding)

	// Generate just this story
	if err := generate.WriteSingleStory(story); err != nil {
		return fmt.Errorf("failed to write story %s: %v", storyID, err)
	}

	log.Printf("Single story regeneration completed successfully for: %s", storyID)
	return nil
}

// main builds the archive and optionally serves it in development mode.
func main() {
	devMode := flag.Bool("development", false, "Run in development mode with live reload")
	flag.BoolVar(devMode, "d", false, "Run in development mode with live reload (shorthand)")
	skipImages := flag.Bool("skip-images", false, "Skip image processing and generation")
	flag.BoolVar(skipImages, "s", false, "Skip image processing and generation (shorthand)")
	debugDump := flag.Bool("debug-dump", false, "Dump raw NocoDB data to JSON file for debugging")
	storyID := flag.String("story-id", "", "Regenerate a specific story by ID (for debugging)")
	clearCache := flag.Bool("clear-cache", false, "Clear the disk cache and fetch fresh data from NocoDB")
	useCacheOnly := flag.Bool("cache-only", false, "Use only disk cache, fail if not available (for offline debugging)")
	flag.Parse()

	// Load configuration from environment variables and .env file
	config.LoadConfig()

	// Initialize the data adapter based on configuration
	if err := store.InitializeAdapter(); err != nil {
		log.Fatalf("Failed to initialize data adapter: %v", err)
	}

	// Handle cache management flags
	if *clearCache {
		log.Println("Clearing all caches...")
		adapter := store.GetAdapter()
		if err := adapter.DropCache(); err != nil {
			log.Printf("Warning: Failed to drop in-memory cache: %v", err)
		}
		if err := adapter.ClearDiskCache(); err != nil {
			log.Printf("Warning: Failed to clear disk cache: %v", err)
		}
		log.Println("Cache clearing completed")
		return
	}

	if *useCacheOnly {
		log.Println("Cache-only mode: Will only use disk cache, no API calls")
		adapter := store.GetAdapter()
		adapter.SetCacheOnlyMode(true)
	}

	// Handle debug dump flag
	if *debugDump {
		if err := dumpNocoDBData(); err != nil {
			log.Fatalf("Debug dump failed: %v", err)
		}
		return
	}

	// Handle single story regeneration flag
	if *storyID != "" {
		if err := generateSingleStory(*storyID); err != nil {
			log.Fatalf("Single story generation failed: %v", err)
		}
		return
	}

	if *skipImages {
		log.Println("Skipping image processing and generation")
	}

	// Build the archive.
	if err := generateArchive(*skipImages); err != nil {
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
