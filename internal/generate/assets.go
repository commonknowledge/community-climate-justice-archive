// assets.go contains functions for copying CSS and image files to the output directory at out/.
package generate

import (
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"image/jpeg"
	"image/png"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
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

func compressImage(srcPath, dstPath string) error {
	ext := strings.ToLower(filepath.Ext(srcPath))

	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return fmt.Errorf("unsupported image format: %s", ext)
	}

	log.Printf("Beginning compression of image %s", srcPath)

	file, err := os.Open(srcPath)
	if err != nil {
		log.Fatalln(err)
	}

	defer file.Close()

	var img image.Image

	if ext == ".jpg" || ext == ".jpeg" {
		img, err = jpeg.Decode(file)
		if err != nil {
			log.Fatalln(err)
		}
	} else if ext == ".png" {
		img, err = png.Decode(file)
		if err != nil {
			log.Fatalln(err)
		}
	}

	output, err := os.Create(dstPath)
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
	if err != nil {
		log.Printf("Failed to create encoder options for image %s: %w", srcPath, err)
		log.Fatalln(err)
	}

	log.Printf("Encoding image %s", dstPath)

	if err := webp.Encode(output, img, options); err != nil {
		log.Printf("Failed to encode image %s: %w", srcPath, err)
		log.Fatalln(err)
	}

	log.Printf("Successfully encoded image %s", dstPath)

	return nil
}

func ProcessImages() error {
	log.Println("Starting image processing process")

	files, err := os.ReadDir("images")
	if err != nil {
		return fmt.Errorf("failed to read images directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		ext := strings.ToLower(filepath.Ext(filename))

		// Skip if the file is not a JPEG or PNG
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			log.Printf("Skipping non-image file: %s", filename)
			continue
		}

		// Create a new filename with the same name but with a new extension
		compressedFilename := strings.TrimSuffix(filename, ext) + ".webp"

		// Check if the compressed file already exists
		if _, err := os.Stat(filepath.Join("images", compressedFilename)); os.IsNotExist(err) {
			// Compress the image
			err := compressImage(filepath.Join("images", filename), filepath.Join("images", compressedFilename))
			if err != nil {
				return fmt.Errorf("failed to compress image %s: %w", filename, err)
			}
		} else {
			log.Printf("Skipping compressed image %s", compressedFilename)
		}
	}

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
