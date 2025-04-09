// Package generate provides asset copying functionality for the Dudley People's School for Climate Justice website.
// This file contains functions for copying CSS and image files to the output directory.
package generate

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// CopyCSSToOutput copies the CSS file to the out/css directory.
func CopyCSSToOutput() error {
	log.Println("Starting CSS copy process")

	srcPath := "css/styles.css"
	dstPath := "out/css/styles.css"

	err := os.MkdirAll("out/css", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output CSS directory: %w", err)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source CSS file %s: %w", srcPath, err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination CSS file %s: %w", dstPath, err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy CSS file %s: %w", srcPath, err)
	}

	log.Printf("Successfully copied CSS file to %s", dstPath)

	return nil
}

// CopyImagesToOutput copies all images from the images directory to the out/images directory.
func CopyImagesToOutput() error {
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
