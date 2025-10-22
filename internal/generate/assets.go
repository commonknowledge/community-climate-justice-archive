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
	"time"

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

// logWebPFailure logs WebP encoding failures to a file instead of crashing
func logWebPFailure(imagePath string, err error) {
	errorLog := "webp_failures.log"
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	errorMessage := fmt.Sprintf("[%s] Failed to encode WebP for %s: %v\n", timestamp, imagePath, err)

	// Append to log file
	file, openErr := os.OpenFile(errorLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openErr != nil {
		log.Printf("Warning: Could not open WebP failure log file: %v", openErr)
		log.Printf("Original WebP error for %s: %v", imagePath, err)
		return
	}
	defer file.Close()

	if _, writeErr := file.WriteString(errorMessage); writeErr != nil {
		log.Printf("Warning: Could not write to WebP failure log: %v", writeErr)
		log.Printf("Original WebP error for %s: %v", imagePath, err)
		return
	}

	log.Printf("Warning: WebP encoding failed for %s (logged to %s): %v", imagePath, errorLog, err)
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

func readImage(srcPath string) (image.Image, error) {
	file, err := os.Open(srcPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	var img image.Image
	ext := strings.ToLower(filepath.Ext(srcPath))
	if ext == ".png" {
		img, err = png.Decode(file)
	} else if ext == ".jpg" || ext == ".jpeg" {
		img, err = jpeg.Decode(file)
	} else {
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// compressImage creates multiple WebP versions of an image at different sizes
func compressImage(srcPath string) error {
	// Read the original image
	originalImg, err := readImage(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	// Get the base name without extension
	baseName := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))

	// Create each size variant
	for suffix, width := range ImageSizes {
		// Calculate proportional height to maintain aspect ratio
		bounds := originalImg.Bounds()
		ratio := float64(bounds.Dy()) / float64(bounds.Dx())
		height := int(float64(width) * ratio)

		// Create the output path
		outPath := filepath.Join("images/processed", fmt.Sprintf("%s_%s.webp", baseName, suffix))

		// Skip if file already exists
		if _, err := os.Stat(outPath); err == nil {
			log.Printf("Skipping existing image: %s", outPath)
			continue
		}

		// Resize the image
		resized := imaging.Resize(originalImg, width, height, imaging.Lanczos)

		// Create and encode the WebP file
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
			logWebPFailure(outPath, err)
			output.Close()
			os.Remove(outPath)
			continue
		}

		log.Printf("Created %s as WebP version of %s", outPath, srcPath)
	}

	// Create the main WebP version (original size)
	mainOutPath := filepath.Join("images/processed", fmt.Sprintf("%s.webp", baseName))

	// Skip if file already exists
	if _, err := os.Stat(mainOutPath); err == nil {
		log.Printf("Skipping existing main image: %s", mainOutPath)
		return nil
	}

	// Create and encode the main WebP file
	output, err := os.Create(mainOutPath)
	if err != nil {
		return fmt.Errorf("failed to create main output file: %w", err)
	}
	defer output.Close()

	// Encode the main image, but with lossless quality
	if err := libwebp.Encode(output, originalImg, webpoptions.EncodingOptions{
		Quality:        0, // Lossless
		EncodingPreset: webpoptions.EncodingPreset(webpoptions.EncodingPresetDefault),
		UseSharpYuv:    true,
	}); err != nil {
		logWebPFailure(mainOutPath, err)
		output.Close()
		os.Remove(mainOutPath)
		return nil
	}

	log.Printf("Created %s as main WebP version of %s", mainOutPath, srcPath)
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
			log.Printf("Warning: Failed to process image %s: %v", filename, err)
			continue
		}
	}

	return nil
}

// CopyImagesToOutput copies all images from the images directory to the out/images directory.
func CopyImagesToOutput() error {
	log.Println("Starting image copy process")

	// Create the output directory for images
	err := os.MkdirAll("out/images", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output images directory: %w", err)
	}

	// Walk through the images directory recursively
	copyCount := 0
	skippedCount := 0
	err = filepath.Walk("images", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the relative path within the images directory
		relativePath, err := filepath.Rel("images", path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Create the destination path
		destinationPath := filepath.Join("out/images", relativePath)

		if info.IsDir() {
			// Create the directory in the output path
			if err := os.MkdirAll(destinationPath, 0755); err != nil {
				return fmt.Errorf("failed to create destination directory: %w", err)
			}
			return nil
		}

		// Only copy image files
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
			// Copy the file
			sourceFile, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open source image %s: %w", path, err)
			}
			defer sourceFile.Close()

			destinationFile, err := os.Create(destinationPath)
			if err != nil {
				return fmt.Errorf("failed to create destination image %s: %w", destinationPath, err)
			}
			defer destinationFile.Close()

			if _, err := io.Copy(destinationFile, sourceFile); err != nil {
				return fmt.Errorf("failed to copy image %s: %w", path, err)
			}

			copyCount++
			log.Printf("Copied %s", relativePath)
		default:
			skippedCount++
			log.Printf("Skipped non-image file: %s", path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking images directory: %w", err)
	}

	log.Printf("Successfully copied %d images to output directory (skipped %d non-image files)", copyCount, skippedCount)
	return nil
}

// CopyJSToOutput copies JavaScript files to the out/js directory.
func CopyJSToOutput() error {
	log.Println("Starting JavaScript copy process")

	err := os.MkdirAll("out/js", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output JS directory: %w", err)
	}

	// Copy all JS files from static/js
	jsFiles, err := os.ReadDir("static/js")
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("No static/js directory found, skipping JS copy")
			return nil
		}
		return fmt.Errorf("failed to read static/js directory: %w", err)
	}

	for _, file := range jsFiles {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		if !strings.HasSuffix(filename, ".js") {
			continue
		}

		srcPath := filepath.Join("static/js", filename)
		dstPath := filepath.Join("out/js", filename)

		src, err := os.Open(srcPath)
		if err != nil {
			return fmt.Errorf("failed to open source JS file %s: %w", srcPath, err)
		}
		defer src.Close()

		dst, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("failed to create destination JS file %s: %w", dstPath, err)
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			return fmt.Errorf("failed to copy JS file %s: %w", srcPath, err)
		}

		log.Printf("Successfully copied JS file to %s", dstPath)
	}

	return nil
}

// CopyAudioToOutput copies audio files to the out/audio directory.
func CopyAudioToOutput() error {
	log.Println("Starting audio copy process")

	// Create the output directory for audio
	err := os.MkdirAll("out/audio", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output audio directory: %w", err)
	}

	// Walk through the images directory to find audio files (they might be mixed in)
	copyCount := 0
	err = filepath.Walk("images", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only copy audio files
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".mp3", ".wav", ".ogg", ".m4a", ".aac", ".flac":
			filename := filepath.Base(path)
			destinationPath := filepath.Join("out/audio", filename)

			// Copy the file
			sourceFile, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open source audio %s: %w", path, err)
			}
			defer sourceFile.Close()

			destinationFile, err := os.Create(destinationPath)
			if err != nil {
				return fmt.Errorf("failed to create destination audio %s: %w", destinationPath, err)
			}
			defer destinationFile.Close()

			if _, err := io.Copy(destinationFile, sourceFile); err != nil {
				return fmt.Errorf("failed to copy audio %s: %w", path, err)
			}

			copyCount++
			log.Printf("Copied audio file: %s", filename)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directories for audio files: %w", err)
	}

	log.Printf("Successfully copied %d audio files to output directory", copyCount)
	return nil
}

// CopyDocumentsToOutput copies document files to the out/documents directory.
func CopyDocumentsToOutput() error {
	log.Println("Starting documents copy process")

	// Create the output directory for documents
	err := os.MkdirAll("out/documents", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output documents directory: %w", err)
	}

	// Walk through the images directory to find document files (they might be mixed in)
	copyCount := 0
	err = filepath.Walk("images", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only copy document files
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".rtf":
			filename := filepath.Base(path)
			destinationPath := filepath.Join("out/documents", filename)

			// Copy the file
			sourceFile, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open source document %s: %w", path, err)
			}
			defer sourceFile.Close()

			destinationFile, err := os.Create(destinationPath)
			if err != nil {
				return fmt.Errorf("failed to create destination document %s: %w", destinationPath, err)
			}
			defer destinationFile.Close()

			if _, err := io.Copy(destinationFile, sourceFile); err != nil {
				return fmt.Errorf("failed to copy document %s: %w", path, err)
			}

			copyCount++
			log.Printf("Copied document file: %s", filename)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directories for document files: %w", err)
	}

	log.Printf("Successfully copied %d document files to output directory", copyCount)
	return nil
}
