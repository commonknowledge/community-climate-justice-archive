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

	"github.com/bep/gowebp/libwebp"
	"github.com/bep/gowebp/libwebp/webpoptions"
	"github.com/disintegration/imaging"
)

// ImageSizes defines the different sizes we generate for each image
var ImageSizes = map[string]int{
	"thumb":  300,  // thumbnail size
	"medium": 800,  // medium size
	"large":  1200, // full size
}

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

// compressImage creates multiple WebP versions of an image at different sizes
func compressImage(srcPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	var img image.Image
	ext := strings.ToLower(filepath.Ext(srcPath))
	if ext == ".png" {
		img, err = png.Decode(file)
	} else if ext == ".jpg" || ext == ".jpeg" {
		img, err = jpeg.Decode(file)
	} else {
		return fmt.Errorf("unsupported image format: %s", ext)
	}
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Create each size variant
	for suffix, width := range ImageSizes {
		// Calculate proportional height
		bounds := img.Bounds()
		ratio := float64(bounds.Dy()) / float64(bounds.Dx())
		height := int(float64(width) * ratio)

		resized := imaging.Resize(img, width, height, imaging.Lanczos)

		// Create WebP with size suffix
		baseName := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
		outPath := filepath.Join("images/processed", fmt.Sprintf("%s_%s.webp", baseName, suffix))

		// Skip if file already exists
		if _, err := os.Stat(outPath); err == nil {
			log.Printf("Skipping existing image: %s", outPath)
			continue
		}

		output, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer output.Close()

		if err := libwebp.Encode(output, resized, webpoptions.EncodingOptions{
			Quality:        75,
			EncodingPreset: webpoptions.EncodingPreset(webpoptions.EncodingPresetDefault),
			UseSharpYuv:    true,
		}); err != nil {
			return fmt.Errorf("failed to encode WebP: %w", err)
		}

		log.Printf("Created %s", outPath)
	}

	// Encode the main image, but with lossless quality
	mainPath := filepath.Join("images", filepath.Base(srcPath))
	mainOutPath := filepath.Join("images/processed", filepath.Base(srcPath))

	ext = strings.ToLower(filepath.Ext(mainPath))
	if ext == ".png" {
		img, err = png.Decode(file)
	} else if ext == ".jpg" || ext == ".jpeg" {
		img, err = jpeg.Decode(file)
	} else {
		return fmt.Errorf("unsupported image format: %s", ext)
	}
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	output, err := os.Create(mainOutPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	if err := libwebp.Encode(output, img, webpoptions.EncodingOptions{
		// Quality of 0 sets the image to lossless
		Quality:        0,
		EncodingPreset: webpoptions.EncodingPreset(webpoptions.EncodingPresetDefault),
		UseSharpYuv:    true,
	}); err != nil {
		return fmt.Errorf("failed to encode WebP: %w", err)
	}

	log.Printf("Created %s", mainOutPath)

	return nil
}

func ProcessImages() error {
	log.Println("Starting image processing process")

	err := os.MkdirAll("images/processed", 0755)
	if err != nil {
		return fmt.Errorf("failed to create resized images directory: %w", err)
	}

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

		// Skip if it's already a WebP file
		if ext == ".webp" {
			log.Printf("Skipping WebP file: %s", filename)
			continue
		}

		// Skip if the file is not a JPEG or PNG
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			log.Printf("Skipping non-image file: %s", filename)
			continue
		}

		// Process the image
		err := compressImage(filepath.Join("images", filename))
		if err != nil {
			return fmt.Errorf("failed to process image %s: %w", filename, err)
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
