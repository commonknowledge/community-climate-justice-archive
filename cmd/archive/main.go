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
// --development or -d: Run in development mode (serves locally, CSS auto-copy, enter-to-regenerate)
// --skip-images or -s: Skip processing images (faster if you're just testing templates)
//
// Development mode is really handy - it starts a local server, watches CSS
// changes, and lets you press Enter to regenerate pages after template edits.
package main

import (
	"bufio"
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

const localDevServerURL = "http://localhost:8080"

// generateArchive builds the entire website.
//
// This does the main work:
// - Fetches all the story data from the database
// - Processes all the images (unless you've skipped them)
// - Fills in the HTML templates with the actual data
// - Writes all the HTML files to the out/ folder
// - Copies over the CSS, JS, and other assets
//
// The skipImages and skipImageCopy parameters let you skip parts of the build:
// - skipImages: Don't process/resize images (for quick template testing)
// - skipImageCopy: Don't copy images to output (for hot reloads where images haven't changed)
func generateArchive(skipImages bool, skipImageCopy bool) error {
	// Record when the build started so we can show how long it took
	buildStartTime := time.Now()

	log.Println("Starting build process")
	generate.ResetBuildCache()

	// Warm the cache to ensure all subsequent operations are fast
	store.WarmCache()

	// Process images (resize, convert to WebP) unless skipped
	if !skipImages {
		if err := generate.ProcessImages(); err != nil {
			return fmt.Errorf("failed to process images: %v", err)
		}
	}

	// Generate all the HTML pages
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

	// Copy assets to output folder
	if !skipImageCopy {
		if err := generate.CopyImagesToOutput(); err != nil {
			return fmt.Errorf("failed to copy images: %v", err)
		}
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

	// Calculate how long the build took
	buildDuration := time.Since(buildStartTime)

	// Display completion message with build time
	if skipImages {
		log.Printf("Build completed (images skipped) in %s", formatDuration(buildDuration))
	} else {
		log.Printf("Build completed in %s", formatDuration(buildDuration))
	}

	return nil
}

// watchCSS sets up a file watcher for the CSS directory.
// When CSS files change, they're automatically copied to the output folder.
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

// formatDuration converts a duration into a friendly, human-readable string.
//
// Examples:
//   - 1.234s → "1.2s"
//   - 45.678s → "45.7s"
//   - 1m30s → "1m 30s"
func formatDuration(d time.Duration) string {
	// For durations under a minute, show seconds with one decimal place
	if d < time.Minute {
		seconds := float64(d) / float64(time.Second)
		return fmt.Sprintf("%.1fs", seconds)
	}

	// For longer durations, show minutes and seconds
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

// waitForInput waits for the user to press enter.
// Used in development mode to trigger a rebuild.
func waitForInput() {
	log.Println("Press enter to regenerate the archive...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadRune()
}

// main builds the archive and optionally serves it in development mode.
func main() {
	// Define command-line flags
	devMode := flag.Bool("development", false, "Run in development mode with live reload")
	flag.BoolVar(devMode, "d", false, "Run in development mode with live reload (shorthand)")
	skipImages := flag.Bool("skip-images", false, "Skip image processing and generation")
	flag.BoolVar(skipImages, "s", false, "Skip image processing and generation (shorthand)")
	clearCache := flag.Bool("clear-cache", false, "Clear the disk cache and fetch fresh data from NocoDB")
	useCacheOnly := flag.Bool("cache-only", false, "Use only disk cache, fail if not available (for offline debugging)")
	flag.Parse()

	// Load configuration from environment variables and .env file
	config.LoadConfig()

	// Initialize the store (connects to NocoDB)
	if err := store.Initialize(); err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	// Handle cache management flags
	if *clearCache {
		log.Println("Clearing all caches...")
		if err := store.DropCache(); err != nil {
			log.Printf("Warning: Failed to drop in-memory cache: %v", err)
		}
		if err := store.ClearDiskCache(); err != nil {
			log.Printf("Warning: Failed to clear disk cache: %v", err)
		}
		log.Println("Cache clearing completed")
		return
	}

	if *useCacheOnly {
		log.Println("Cache-only mode: Will only use disk cache, no API calls")
		store.SetCacheOnlyMode(true)
	}

	if *skipImages {
		log.Println("Skipping image processing and generation")
	}

	// Build the archive (full build on first run)
	if err := generateArchive(*skipImages, false); err != nil {
		log.Fatalf("Initial build failed: %v", err)
	}

	// If in development mode, start the server and watch for changes
	if *devMode {
		log.Println("Starting development server...")

		// Start a simple HTTP server to serve the archive on port 8080.
		// The "go" keyword runs the server in a separate goroutine (parallel thread).
		// This allows the server to run while we wait for user input.
		go server.Serve()

		// Watch the CSS directory for changes
		watcher, err := watchCSS()
		if err != nil {
			log.Printf("Failed to initialize CSS watcher: %v", err)
		} else {
			defer watcher.Close()
		}

		log.Printf("Development server running at %s", localDevServerURL)

		// Wait for user to press enter, then rebuild
		for {
			waitForInput()
			// Hot regenerate: skip image processing and image copying (they're unchanged)
			if err := generateArchive(true, true); err != nil {
				log.Printf("Hot regeneration failed: %v", err)
			}
		}
	}
}
