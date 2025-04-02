// The archive command builds the archive and serves it on port 8080.
package main

import (
	"bufio"
	"community-climate-justice-archive/internal/generate"
	"community-climate-justice-archive/internal/server"
	"fmt"
	"log"
	"os"
)

// regenerate builds the archive.
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

// main builds the archive and serves it on port 8080.
func main() {
	// Build the archive.
	if err := regenerate(); err != nil {
		log.Fatalf("Initial build failed: %v", err)
	}

	// Start a simple HTTP server to serve the archive on port 8080.
	go server.Serve()

	// Wait for input and then rebuild the archive when enter is pressed.
	for {
		waitForInput()
		if err := regenerate(); err != nil {
			log.Printf("Regeneration failed: %v", err)
		}
	}
}
