// assets.go contains functions for copying CSS and image files to the output directory at out/.
package generate

import (
	"community-climate-justice-archive/data" // Added data package
	"encoding/json"                          // Added for NocoDBAttachment
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
		return nil, fmt.Errorf("unsupported image format: %s for source path %s", ext, srcPath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to decode image %s: %w", srcPath, err)
	}

	return img, nil
}

// compressImage creates multiple WebP versions of an image at different sizes.
// srcPath is the full path to the original NocoDB image (e.g., nc/uploads/.../image.png).
// baseOriginalFilename is the simple filename (e.g., image.png).
func compressImage(nocoDBLocalSrcPath string, baseOriginalFilename string) error {
	originalImg, err := readImage(nocoDBLocalSrcPath)
	if err != nil {
		return fmt.Errorf("compressImage failed to read image %s: %w", nocoDBLocalSrcPath, err)
	}

	baseNameWithoutExt := strings.TrimSuffix(baseOriginalFilename, filepath.Ext(baseOriginalFilename))
	outputBaseDir := filepath.Join("images", "processed")

	// Ensure the output directory exists
	if err := os.MkdirAll(outputBaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create processed images directory %s: %w", outputBaseDir, err)
	}

	// Create each size variant
	for suffix, width := range ImageSizes {
		bounds := originalImg.Bounds()
		ratio := float64(bounds.Dy()) / float64(bounds.Dx())
		height := int(float64(width) * ratio)
		if height == 0 && bounds.Dy() != 0 { // Avoid division by zero if width is 0, or if original height is 0
			height = width        // Default to square if calculation is problematic, or handle error
			if bounds.Dx() != 0 { // Recalculate if original Dx is not 0
				height = int(float64(width) * (float64(bounds.Dy()) / float64(bounds.Dx())))
			} else if width > 0 { // If original Dx is 0 but target width is not, maybe it's a very thin vertical line
				log.Printf("Warning: Original image %s has zero width. Cannot maintain aspect ratio for width %d.", nocoDBLocalSrcPath, width)
				height = width // or some other default logic
			}
		}

		outPath := filepath.Join(outputBaseDir, fmt.Sprintf("%s_%s.webp", baseNameWithoutExt, suffix))

		if _, err := os.Stat(outPath); err == nil {
			log.Printf("Skipping existing image: %s", outPath)
			continue
		}

		resized := imaging.Resize(originalImg, width, height, imaging.Lanczos)
		output, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outPath, err)
		}
		defer output.Close()

		if err := libwebp.Encode(output, resized, webpoptions.EncodingOptions{
			Quality:        75,
			EncodingPreset: webpoptions.EncodingPreset(webpoptions.EncodingPresetDefault),
			UseSharpYuv:    true,
		}); err != nil {
			output.Close() // Close immediately on error
			return fmt.Errorf("failed to encode WebP %s: %w", outPath, err)
		}
		output.Close() // Close after successful write
		log.Printf("Created %s from %s", outPath, nocoDBLocalSrcPath)
	}

	// Create the main WebP version (original size)
	mainOutPath := filepath.Join(outputBaseDir, fmt.Sprintf("%s.webp", baseNameWithoutExt))
	if _, err := os.Stat(mainOutPath); err == nil {
		// log.Printf("Skipping existing main image: %s", mainOutPath) // Optional
		return nil
	}

	output, err := os.Create(mainOutPath)
	if err != nil {
		return fmt.Errorf("failed to create main output file %s: %w", mainOutPath, err)
	}
	defer output.Close()

	if err := libwebp.Encode(output, originalImg, webpoptions.EncodingOptions{
		Quality:        90, // Higher quality for main full-size WebP
		EncodingPreset: webpoptions.EncodingPreset(webpoptions.EncodingPresetDefault),
		UseSharpYuv:    true,
	}); err != nil {
		output.Close()
		return fmt.Errorf("failed to encode main WebP %s: %w", mainOutPath, err)
	}
	output.Close()
	log.Printf("Created %s as main WebP version of %s", mainOutPath, nocoDBLocalSrcPath)
	return nil
}

// ProcessNocoDBImages generates WebP versions for NocoDB images.
// It reads originals from nc/uploads/... and writes WebPs to out/images/processed/.
func ProcessNocoDBImages(stories []data.Story) error {
	log.Println("Starting NocoDB image processing (WebP generation)")

	processedPaths := make(map[string]bool) // To avoid processing the same image multiple times

	for i, story := range stories {
		log.Printf("--- Processing Story %d/%d - ID: %s, Finding: %s ---", i+1, len(stories), story.ID, story.Finding)
		var allNocoDBAttachments []data.NocoDBAttachment

		// Process story.Image
		if story.Image != "" && story.Image != "[]" {
			var currentAttachments []data.NocoDBAttachment
			err := json.Unmarshal([]byte(story.Image), &currentAttachments)
			if err != nil {
				log.Printf("Error unmarshalling Story.Image for story ID %s: %v. JSON: %s", story.ID, err, story.Image)
			} else {
				allNocoDBAttachments = append(allNocoDBAttachments, currentAttachments...)
			}
		}

		// Process story.SourceImage
		if story.SourceImage != "" && story.SourceImage != "[]" {
			var currentAttachments []data.NocoDBAttachment
			err := json.Unmarshal([]byte(story.SourceImage), &currentAttachments)
			if err != nil {
				log.Printf("Error unmarshalling Story.SourceImage for story ID %s: %v. JSON: %s", story.ID, err, story.SourceImage)
			} else {
				allNocoDBAttachments = append(allNocoDBAttachments, currentAttachments...)
			}
		}

		for _, nocoAtt := range allNocoDBAttachments {
			if nocoAtt.Path == "" {
				continue
			}

			nocoDBLocalFilePath := ""
			if strings.HasPrefix(nocoAtt.Path, "download/") {
				nocoDBLocalFilePath = strings.Replace(nocoAtt.Path, "download/", "nc/uploads/", 1)
			} else {
				log.Printf("NocoDBAttachment path '%s' (for story ID %s, attachment Title '%s') does not start with 'download/'. Assuming it is a direct usable path for sourcing image. Ensure build process can handle this.", nocoAtt.Path, story.ID, nocoAtt.Title)
				nocoDBLocalFilePath = nocoAtt.Path // Assume it's a path relative to project root or absolute
			}

			if _, err := os.Stat(nocoDBLocalFilePath); os.IsNotExist(err) {
				log.Printf("Source image %s for story ID %s (attachment title '%s') does not exist. Skipping WebP generation for it.", nocoDBLocalFilePath, story.ID, nocoAtt.Title)
				continue
			}

			if processedPaths[nocoDBLocalFilePath] {
				continue // Already processed this source image
			}

			baseOriginalFilename := filepath.Base(nocoDBLocalFilePath)
			if baseOriginalFilename == "." || baseOriginalFilename == "/" {
				log.Printf("Invalid base filename for NocoDB image path %s (story ID %s). Skipping.", nocoDBLocalFilePath, story.ID)
				continue
			}

			// Skip if it\'s already a WebP file in source - though Noco usually doesn't store WebP directly from UI
			ext := strings.ToLower(filepath.Ext(baseOriginalFilename))
			if ext == ".webp" {
				log.Printf("Skipping WebP generation for %s as it is already a WebP file.", nocoDBLocalFilePath)
				processedPaths[nocoDBLocalFilePath] = true
				continue
			}
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" { // Added GIF
				log.Printf("Skipping WebP generation for %s: unsupported source format %s", nocoDBLocalFilePath, ext)
				processedPaths[nocoDBLocalFilePath] = true
				continue
			}

			err := compressImage(nocoDBLocalFilePath, baseOriginalFilename)
			if err != nil {
				log.Printf("Failed to process NocoDB image %s for story ID %s: %v", nocoDBLocalFilePath, story.ID, err)
			}
			processedPaths[nocoDBLocalFilePath] = true
		}
	}
	log.Println("NocoDB image processing (WebP generation) completed.")
	return nil
}

// CopyNocoDBImagesToOutput copies original NocoDB images from nc/uploads/... to out/images/.
func CopyNocoDBImagesToOutput(stories []data.Story) error {
	log.Println("Starting NocoDB original image copy to out/images/")
	outputDir := filepath.Join("out", "images")

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output images directory %s: %w", outputDir, err)
	}

	copiedFiles := make(map[string]bool) // To avoid copying the same file multiple times

	for _, story := range stories {
		var allNocoDBAttachments []data.NocoDBAttachment

		// Process story.Image (same unmarshalling logic as ProcessNocoDBImages)
		if story.Image != "" && story.Image != "[]" {
			var currentAttachments []data.NocoDBAttachment
			err := json.Unmarshal([]byte(story.Image), &currentAttachments)
			if err != nil { // Minor error, log and continue
				log.Printf("CopyNocoDB: Error unmarshalling Story.Image for story ID %s: %v.", story.ID, err)
			} else {
				allNocoDBAttachments = append(allNocoDBAttachments, currentAttachments...)
			}
		}
		if story.SourceImage != "" && story.SourceImage != "[]" {
			var currentAttachments []data.NocoDBAttachment
			err := json.Unmarshal([]byte(story.SourceImage), &currentAttachments)
			if err != nil { // Minor error, log and continue
				log.Printf("CopyNocoDB: Error unmarshalling Story.SourceImage for story ID %s: %v.", story.ID, err)
			} else {
				allNocoDBAttachments = append(allNocoDBAttachments, currentAttachments...)
			}
		}

		for _, nocoAtt := range allNocoDBAttachments {
			if nocoAtt.Path == "" {
				continue
			}

			nocoDBLocalFilePath := ""
			if strings.HasPrefix(nocoAtt.Path, "download/") {
				nocoDBLocalFilePath = strings.Replace(nocoAtt.Path, "download/", "nc/uploads/", 1)
			} else {
				// Path does not start with "download/", treat as direct path.
				// Build process needs to be able to source the image.
				nocoDBLocalFilePath = nocoAtt.Path
			}

			if copiedFiles[nocoDBLocalFilePath] {
				continue // Already copied
			}

			// Check if source file exists before attempting copy
			srcStat, err := os.Stat(nocoDBLocalFilePath)
			if os.IsNotExist(err) {
				log.Printf("Source image file does not exist, cannot copy: %s (Story ID: %s, Attachment Title: %s)", nocoDBLocalFilePath, story.ID, nocoAtt.Title)
				copiedFiles[nocoDBLocalFilePath] = true // Mark as "handled" to avoid re-logging
				continue
			}
			if err != nil {
				log.Printf("Error stating source file %s: %v. Skipping copy.", nocoDBLocalFilePath, err)
				copiedFiles[nocoDBLocalFilePath] = true
				continue
			}
			if srcStat.IsDir() {
				log.Printf("Source path is a directory, not a file. Skipping copy: %s", nocoDBLocalFilePath)
				copiedFiles[nocoDBLocalFilePath] = true
				continue
			}

			baseOriginalFilename := filepath.Base(nocoDBLocalFilePath)
			if baseOriginalFilename == "." || baseOriginalFilename == "/" {
				log.Printf("Invalid base filename for NocoDB image path %s (story ID %s). Skipping copy.", nocoDBLocalFilePath, story.ID)
				copiedFiles[nocoDBLocalFilePath] = true
				continue
			}

			destinationPath := filepath.Join(outputDir, baseOriginalFilename)

			// Optional: Check if destination already exists and is identical? For now, simple copy.
			// if _, err := os.Stat(destinationPath); err == nil {
			//  log.Printf("Destination file %s already exists. Overwriting.", destinationPath)
			// }

			sourceFile, err := os.Open(nocoDBLocalFilePath)
			if err != nil {
				log.Printf("Failed to open source image %s for copy: %v", nocoDBLocalFilePath, err)
				copiedFiles[nocoDBLocalFilePath] = true
				continue
			}
			defer sourceFile.Close()

			destinationFile, err := os.Create(destinationPath)
			if err != nil {
				sourceFile.Close()
				log.Printf("Failed to create destination image %s for copy: %v", destinationPath, err)
				copiedFiles[nocoDBLocalFilePath] = true
				continue
			}
			defer destinationFile.Close()

			_, err = io.Copy(destinationFile, sourceFile)
			if err != nil {
				sourceFile.Close()
				destinationFile.Close()
				log.Printf("Failed to copy image from %s to %s: %v", nocoDBLocalFilePath, destinationPath, err)
				copiedFiles[nocoDBLocalFilePath] = true
				continue
			}
			sourceFile.Close()
			destinationFile.Close()

			log.Printf("Copied NocoDB original %s to %s", nocoDBLocalFilePath, destinationPath)
			copiedFiles[nocoDBLocalFilePath] = true
		}
	}
	log.Println("NocoDB original image copy to out/images/ completed.")
	return nil
}

// CopyImagesToOutput copies all images from the images directory to the out/images directory.
// THIS FUNCTION MAY BE REDUNDANT or need to be re-evaluated if all user images come from NocoDB.
// If there are other static site assets (e.g. logos, favicons) in a local "images/" folder, this might still be needed.
func CopyImagesToOutput() error {
	log.Println("Starting image copy process (from local 'images/' to 'out/images/')")
	// This function now primarily serves to copy any *other* static images
	// that are not managed by NocoDB but are part of the site (e.g. logos, favicons in a local 'images' dir).

	sourceDir := "images" // The local directory to scan
	outputDir := filepath.Join("out", "images")

	// Check if the source 'images' directory exists. If not, it's not an error,
	// as it might not be used if all content is NocoDB-driven.
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		log.Printf("Local source directory '%s' not found. Skipping copy from this source.", sourceDir)
		return nil
	}

	// Create the output directory for images if it doesn't exist
	// (CopyNocoDBImagesToOutput might have already created out/images)
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output images directory %s: %w", outputDir, err)
	}

	// Walk through the source 'images' directory recursively
	copyCount := 0
	skippedCount := 0
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate the relative path within the sourceDir
		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		// Create the destination path in out/images
		destinationPath := filepath.Join(outputDir, relativePath)

		if info.IsDir() {
			// Create the directory in the output path
			if err := os.MkdirAll(destinationPath, info.Mode()); err != nil { // Use source mode for dirs
				return fmt.Errorf("failed to create destination directory %s: %w", destinationPath, err)
			}
			return nil
		}

		// Only copy image files (can be extended)
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico": // Added more common static types
			// Copy the file
			sourceFile, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open source image %s: %w", path, err)
			}
			defer sourceFile.Close()

			// Ensure destination directory exists (important for nested structures within images/)
			if err := os.MkdirAll(filepath.Dir(destinationPath), 0755); err != nil {
				return fmt.Errorf("failed to create destination sub-directory for %s: %w", destinationPath, err)
			}

			destinationFile, err := os.Create(destinationPath)
			if err != nil {
				return fmt.Errorf("failed to create destination image %s: %w", destinationPath, err)
			}
			defer destinationFile.Close()

			if _, err := io.Copy(destinationFile, sourceFile); err != nil {
				return fmt.Errorf("failed to copy image %s: %w", path, err)
			}

			copyCount++
			// log.Printf("Copied static asset %s to %s", path, destinationPath) // Reduce log noise
		default:
			skippedCount++
			// log.Printf("Skipped non-image file during static copy: %s", path)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking local '%s' directory: %w", sourceDir, err)
	}

	if copyCount > 0 || skippedCount > 0 {
		log.Printf("Successfully copied %d static assets from '%s' to '%s' (skipped %d files)", copyCount, sourceDir, outputDir, skippedCount)
	} else {
		log.Printf("No static assets found in local '%s' to copy to '%s'", sourceDir, outputDir)
	}
	return nil
}

// ProcessImages has been replaced by ProcessNocoDBImages
// func ProcessImages() error {
//  log.Println("Starting image processing process")

//  err := os.MkdirAll("images/processed", 0755)
//  if err != nil {
//      return fmt.Errorf("failed to create resized images directory: %w", err)
//  }

//  files, err := os.ReadDir("images")
//  if err != nil {
//      return fmt.Errorf("failed to read images directory: %w", err)
//  }

//  for _, file := range files {
//      if file.IsDir() {
//          continue
//      }

//      filename := file.Name()
//      ext := strings.ToLower(filepath.Ext(filename))

//      // Skip if it's already a WebP file
//      if ext == ".webp" {
//          log.Printf("Skipping WebP file: %s", filename)
//          continue
//      }

//      // Skip if the file is not a JPEG or PNG
//      if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
//          log.Printf("Skipping non-image file: %s", filename)
//          continue
//      }

//      // Process the image
//      err := compressImage(filepath.Join("images", filename)) // This would need baseOriginalFilename now
//      if err != nil {
//          return fmt.Errorf("failed to process image %s: %w", filename, err)
//      }
//  }

//  return nil
// }
