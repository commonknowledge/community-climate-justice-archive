// Package generate handles processing and copying images, CSS, and other files.
//
// When building the website, we need to prepare all the assets - images, stylesheets,
// audio files, documents - and copy them to the output folder.
//
// What gets handled:
// - Images: Resized into different sizes and converted to WebP (loads faster)
// - CSS: Just copied over
// - JavaScript: Just copied over
// - Audio files: Just copied over
// - Documents (PDFs, Word files): Just copied over
//
// The image processing is the most involved:
// 1. Read the original image from the images/ folder
// 2. Create three versions: thumbnail, medium, and large
// 3. Convert all of them to WebP format (it compresses better than JPG or PNG)
// 4. Save them in images/processed/ so they can be copied into out/images/processed/
//
// Why WebP? It makes images much smaller without losing quality, so pages load
// faster for visitors. Good for accessibility and better for the environment too
// (less data transfer = less energy used).
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

// copyFileIfChanged copies one file only when the destination is missing or stale.
//
// This keeps rebuilds fast because unchanged assets do not get rewritten on every
// run. When we do copy a file, we also preserve the source modification time so
// the next build can make the same freshness check reliably.
func copyFileIfChanged(sourcePath string, destinationPath string) (bool, error) {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return false, fmt.Errorf("failed to stat source file %s: %w", sourcePath, err)
	}

	destinationInfo, err := os.Stat(destinationPath)
	if err == nil {
		// Skip when destination is already up to date and file size matches.
		if destinationInfo.Size() == sourceInfo.Size() && !destinationInfo.ModTime().Before(sourceInfo.ModTime()) {
			return false, nil
		}
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to stat destination file %s: %w", destinationPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
		return false, fmt.Errorf("failed to create destination directory for %s: %w", destinationPath, err)
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return false, fmt.Errorf("failed to open source file %s: %w", sourcePath, err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return false, fmt.Errorf("failed to create destination file %s: %w", destinationPath, err)
	}
	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		return false, fmt.Errorf("failed to copy file %s: %w", sourcePath, err)
	}

	// Preserve source timestamps so future incremental copies can compare accurately.
	if err := os.Chtimes(destinationPath, sourceInfo.ModTime(), sourceInfo.ModTime()); err != nil {
		log.Printf(
			"Warning: copied %s to %s but could not preserve the source timestamp (%s): %v",
			sourcePath,
			destinationPath,
			sourceInfo.ModTime().Format(time.RFC3339),
			err,
		)
	}

	return true, nil
}

// getProcessedImagePaths returns all output WebP paths expected for a source image.
func getProcessedImagePaths(srcPath string) []string {
	baseName := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))

	paths := make([]string, 0, len(ImageSizes)+1)
	for suffix := range ImageSizes {
		paths = append(paths, filepath.Join("images/processed", fmt.Sprintf("%s_%s.webp", baseName, suffix)))
	}
	paths = append(paths, filepath.Join("images/processed", fmt.Sprintf("%s.webp", baseName)))

	return paths
}

// isOutputUpToDate reports whether one generated file is at least as new as its source.
//
// We use this small helper in both image processing and asset copying so the
// build can skip work that has already been done.
func isOutputUpToDate(outputPath string, sourceModTime time.Time) bool {
	outputInfo, err := os.Stat(outputPath)
	if err != nil {
		return false
	}

	// If output is as new or newer than source, we can reuse it.
	return !outputInfo.ModTime().Before(sourceModTime)
}

// areProcessedImagesUpToDate checks whether every expected WebP derivative exists
// and is newer than the original source image.
func areProcessedImagesUpToDate(srcPath string) bool {
	sourceInfo, err := os.Stat(srcPath)
	if err != nil {
		return false
	}

	for _, outputPath := range getProcessedImagePaths(srcPath) {
		if !isOutputUpToDate(outputPath, sourceInfo.ModTime()) {
			return false
		}
	}

	return true
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

// CopyCSSToOutput copies all CSS files to the out/css directory.
func CopyCSSToOutput() error {
	log.Println("Starting CSS copy process")

	err := os.MkdirAll("out/css", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output CSS directory: %w", err)
	}

	entries, err := os.ReadDir("css")
	if err != nil {
		return fmt.Errorf("failed to read CSS directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".css" {
			continue
		}

		srcPath := filepath.Join("css", entry.Name())
		dstPath := filepath.Join("out/css", entry.Name())

		src, err := os.Open(srcPath)
		if err != nil {
			return fmt.Errorf("failed to open source CSS file %s: %w", srcPath, err)
		}

		dst, err := os.Create(dstPath)
		if err != nil {
			src.Close()
			return fmt.Errorf("failed to create destination CSS file %s: %w", dstPath, err)
		}

		_, err = io.Copy(dst, src)
		src.Close()
		closeErr := dst.Close()
		if err != nil {
			return fmt.Errorf("failed to copy CSS file %s: %w", srcPath, err)
		}
		if closeErr != nil {
			return fmt.Errorf("failed to close destination CSS file %s: %w", dstPath, closeErr)
		}

		log.Printf("Successfully copied CSS file to %s", dstPath)
	}

	return nil
}

// readImage opens an image file and decodes it into Go's generic image type.
//
// The archive only processes JPEG and PNG source files, so unsupported formats
// are rejected here with a clear error before any resize work starts.
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
	sourceInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat source image: %w", err)
	}

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

		// Skip if output exists and is already fresh.
		if isOutputUpToDate(outPath, sourceInfo.ModTime()) {
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

	// Skip if output exists and is already fresh.
	if isOutputUpToDate(mainOutPath, sourceInfo.ModTime()) {
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

// ProcessImages updates the generated WebP image set in place.
//
// It only reprocesses source files whose derived outputs are missing or older
// than the source image, then logs a full summary so it is obvious how much work
// the build actually had to do.
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

	processedCount := 0
	skippedUpToDateCount := 0
	failureCount := 0
	skippedUnsupportedCount := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := file.Name()
		ext := strings.ToLower(filepath.Ext(filename))

		// Skip if it's already a WebP file
		if ext == ".webp" {
			skippedUnsupportedCount++
			continue
		}

		// Skip if the file is not a JPEG or PNG
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			skippedUnsupportedCount++
			continue
		}

		sourcePath := filepath.Join("images", filename)
		if areProcessedImagesUpToDate(sourcePath) {
			skippedUpToDateCount++
			continue
		}

		// Process the image
		err := compressImage(sourcePath)
		if err != nil {
			failureCount++
			log.Printf("Warning: Failed to process image %s: %v", filename, err)
			continue
		}

		processedCount++
	}

	log.Printf(
		"Image processing complete: processed=%d skipped_up_to_date=%d skipped_unsupported=%d failures=%d",
		processedCount,
		skippedUpToDateCount,
		skippedUnsupportedCount,
		failureCount,
	)

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
	unchangedCount := 0
	skippedNonImageCount := 0
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
			copied, err := copyFileIfChanged(path, destinationPath)
			if err != nil {
				return fmt.Errorf("failed to copy image %s: %w", path, err)
			}

			if copied {
				copyCount++
			} else {
				unchangedCount++
			}
		default:
			skippedNonImageCount++
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking images directory: %w", err)
	}

	log.Printf("Image copy complete: copied=%d unchanged=%d skipped_non_image=%d", copyCount, unchangedCount, skippedNonImageCount)
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

	copyCount := 0
	unchangedCount := 0

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

		copied, err := copyFileIfChanged(srcPath, dstPath)
		if err != nil {
			return fmt.Errorf("failed to copy JS file %s: %w", srcPath, err)
		}

		if copied {
			copyCount++
		} else {
			unchangedCount++
		}
	}

	log.Printf("JavaScript copy complete: copied=%d unchanged=%d", copyCount, unchangedCount)
	return nil
}

// CopyAudioToOutput copies audio files to the out/audio directory.
func CopyAudioToOutput() error {
	log.Println("Starting audio copy process")

	srcDir := "audio"
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		log.Println("No audio directory found, skipping audio copy")
		return nil
	}

	// Create the output directory for audio
	err := os.MkdirAll("out/audio", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output audio directory: %w", err)
	}

	copyCount := 0
	unchangedCount := 0
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
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

			copied, err := copyFileIfChanged(path, destinationPath)
			if err != nil {
				return fmt.Errorf("failed to copy audio %s: %w", path, err)
			}

			if copied {
				copyCount++
			} else {
				unchangedCount++
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directories for audio files: %w", err)
	}

	log.Printf("Audio copy complete: copied=%d unchanged=%d", copyCount, unchangedCount)
	return nil
}

// CopyVideosToOutput copies video files to the out/video directory.
func CopyVideosToOutput() error {
	log.Println("Starting video copy process")

	srcDir := "video"
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		log.Println("No video directory found, skipping video copy")
		return nil
	}

	err := os.MkdirAll("out/video", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output video directory: %w", err)
	}

	copyCount := 0
	unchangedCount := 0
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".mp4", ".mov", ".webm", ".m4v", ".avi":
			filename := filepath.Base(path)
			destinationPath := filepath.Join("out/video", filename)

			copied, err := copyFileIfChanged(path, destinationPath)
			if err != nil {
				return fmt.Errorf("failed to copy video %s: %w", path, err)
			}

			if copied {
				copyCount++
			} else {
				unchangedCount++
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directories for video files: %w", err)
	}

	log.Printf("Video copy complete: copied=%d unchanged=%d", copyCount, unchangedCount)
	return nil
}

// CopyDocumentsToOutput copies document files to the out/documents directory.
func CopyDocumentsToOutput() error {
	log.Println("Starting documents copy process")

	srcDir := "documents"
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		log.Println("No documents directory found, skipping documents copy")
		return nil
	}

	// Create the output directory for documents
	err := os.MkdirAll("out/documents", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output documents directory: %w", err)
	}

	copyCount := 0
	unchangedCount := 0
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
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

			copied, err := copyFileIfChanged(path, destinationPath)
			if err != nil {
				return fmt.Errorf("failed to copy document %s: %w", path, err)
			}

			if copied {
				copyCount++
			} else {
				unchangedCount++
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directories for document files: %w", err)
	}

	log.Printf("Document copy complete: copied=%d unchanged=%d", copyCount, unchangedCount)
	return nil
}

// CopyStaticToOutput copies static assets (images, etc.) from the static/ directory
// to out/static/ for use by pages like the About page.
func CopyStaticToOutput() error {
	log.Println("Starting static assets copy process")

	srcDir := "static"
	dstDir := "out/static"

	// Check if source directory exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		log.Println("No static directory found, skipping static copy")
		return nil
	}

	// Create the output directory
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create output static directory: %w", err)
	}

	// Walk through the static directory and copy files (excluding js/ which is handled separately)
	copyCount := 0
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the js subdirectory (handled by CopyJSToOutput)
		if info.IsDir() && info.Name() == "js" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		// Get relative path from static/
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		destPath := filepath.Join(dstDir, relPath)

		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
		}

		copied, err := copyFileIfChanged(path, destPath)
		if err != nil {
			return fmt.Errorf("failed to copy file %s: %w", path, err)
		}

		if copied {
			copyCount++
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking static directory: %w", err)
	}

	log.Printf("Successfully copied %d static files to output directory", copyCount)
	return nil
}
