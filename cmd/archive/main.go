package main

import (
	"bufio"
	"community-climate-justice-archive/internal/generate"
	"community-climate-justice-archive/internal/server"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func copyImagesToOutput() error {
	log.Println("Starting image copy process")

	err := os.MkdirAll("out/images", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output images directory: %w", err)
	}

	files, err := os.ReadDir("images")
	if err != nil {
		return fmt.Errorf("failed to read images directory: %w", err)
	}

	copyCount := 0
	skippedCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		ext := strings.ToLower(filepath.Ext(filename))

		// Only copy image files
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
			srcPath := filepath.Join("images", filename)
			dstPath := filepath.Join("out/images", filename)

			src, err := os.Open(srcPath)
			if err != nil {
				return fmt.Errorf("failed to open source image %s: %w", srcPath, err)
			}
			defer src.Close()

			dst, err := os.Create(dstPath)
			if err != nil {
				return fmt.Errorf("failed to create destination image %s: %w", dstPath, err)
			}
			defer dst.Close()

			_, err = io.Copy(dst, src)
			if err != nil {
				return fmt.Errorf("failed to copy image %s: %w", filename, err)
			}

			copyCount++
			log.Printf("Copied %s", filename)
		default:
			skippedCount++
			log.Printf("Skipped non-image file: %s", filename)
		}
	}

	log.Printf("Successfully copied %d images to output directory (skipped %d non-image files)", copyCount, skippedCount)
	return nil
}

func regenerate() error {
	log.Println("Starting build process")

	if err := generate.WriteHomePage(); err != nil {
		return fmt.Errorf("failed to write homepage: %v", err)
	}

	if err := copyImagesToOutput(); err != nil {
		return fmt.Errorf("failed to copy images: %v", err)
	}

	log.Println("Build process completed successfully")
	return nil
}

func waitForInput() {
	log.Println("Press enter to regenerate the archive...")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadRune()
}

func main() {
	// Initial build
	if err := regenerate(); err != nil {
		log.Fatalf("Initial build failed: %v", err)
	}

	go server.Serve()

	for {
		waitForInput()
		if err := regenerate(); err != nil {
			log.Printf("Regeneration failed: %v", err)
		}
	}
}
