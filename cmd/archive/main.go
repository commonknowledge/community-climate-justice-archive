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
	"sync"
	"time"

	"community-climate-justice-archive/internal/config"
	"community-climate-justice-archive/internal/generate"
	"community-climate-justice-archive/internal/server"
	"community-climate-justice-archive/internal/store"

	"github.com/fsnotify/fsnotify"
)

const localDevServerURL = "http://localhost:8080"

// buildTask pairs a short label with the work function for one build step.
type buildTask struct {
	name string
	run  func() error
}

// runBuildTasks launches a group of independent build steps in parallel.
//
// Each task writes its own output files, so we can safely run them concurrently
// and then return the first error that came back from the group.
func runBuildTasks(tasks []buildTask) error {
	if len(tasks) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(tasks))

	for _, task := range tasks {
		task := task
		wg.Add(1)

		// Each task runs in its own goroutine so unrelated build work can happen at once.
		go func() {
			defer wg.Done()
			if err := task.run(); err != nil {
				errCh <- fmt.Errorf("%s: %w", task.name, err)
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		return err
	}

	return nil
}

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
	if err := generate.WarmBuildCache(); err != nil {
		return fmt.Errorf("failed to warm generator cache: %v", err)
	}

	// Process images (resize, convert to WebP) unless skipped
	if !skipImages {
		if err := generate.ProcessImages(); err != nil {
			return fmt.Errorf("failed to process images: %v", err)
		}
	}

	// Generate independent page groups in parallel.
	pageTasks := []buildTask{
		{name: "write stories", run: generate.WriteStories},
		{name: "write homepage", run: generate.WriteHomePage},
		{name: "write wander page", run: generate.WriteWanderPage},
		{name: "write archive page", run: generate.WriteArchivePage},
		{name: "write about page", run: generate.WriteAboutPage},
		{name: "write filter data", run: generate.WriteFilterData},
		{name: "write types indexes", run: generate.WriteTypesIndexes},
		{name: "write themes indexes", run: generate.WriteThemesIndexes},
		{name: "write weather indexes", run: generate.WriteWeatherIndexes},
		{name: "write gifted by indexes", run: generate.WriteGiftedByIndexPages},
		{name: "write scale permanence indexes", run: generate.WriteScalePermanenceIndexPages},
		{name: "write what was/is/if indexes", run: generate.WriteWhatWasIsIfIndexPages},
		{name: "write time period indexes", run: generate.WriteTimePeriodIndexPages},
	}
	if err := runBuildTasks(pageTasks); err != nil {
		return err
	}

	// Copy assets to output folder in parallel.
	assetTasks := []buildTask{
		{name: "copy audio files", run: generate.CopyAudioToOutput},
		{name: "copy document files", run: generate.CopyDocumentsToOutput},
		{name: "copy CSS", run: generate.CopyCSSToOutput},
		{name: "copy JavaScript", run: generate.CopyJSToOutput},
	}
	if !skipImageCopy {
		assetTasks = append(assetTasks, buildTask{name: "copy images", run: generate.CopyImagesToOutput})
	}
	if err := runBuildTasks(assetTasks); err != nil {
		return err
	}

	if err := generate.CopyStaticToOutput(); err != nil {
		return fmt.Errorf("failed to copy static assets: %v", err)
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
	diskCacheMode := flag.Bool("debug-disk-cache", false, "Enable disk cache reads/writes for debugging")
	flag.BoolVar(diskCacheMode, "disk-cache", false, "Enable disk cache reads/writes for debugging")
	clearCache := flag.Bool("clear-cache", false, "Clear the disk cache file and exit")
	useCacheOnly := flag.Bool("cache-only", false, "Use only disk cache, fail if not available (for offline debugging)")
	flag.Parse()

	// Load configuration from environment variables and .env file
	config.LoadConfig()

	// Initialize the store (connects to NocoDB)
	if err := store.Initialize(); err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}

	// Enable on-disk cache only when explicitly requested for debugging.
	if *diskCacheMode {
		log.Println("Debug disk cache mode enabled")
		store.SetDiskCacheMode(true)
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
