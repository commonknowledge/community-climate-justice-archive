package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Image represents an image entry from the JSON data.
type Image struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

// StoryImage associates a story ID with its image data and finding.
// Finding represents the context or description of the image within the story.
type StoryImage struct {
	StoryID string
	Image   Image
	Finding string
}

// FetchImagesFromDB retrieves all images from the database and returns them as StoryImage structs.
// It queries both the "Image" and "Source Image" fields from the Stories table.
// Returns a slice of StoryImage structs and any error encountered during the process.
func FetchImagesFromDB(dbPath string) ([]StoryImage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT _id, Image, \"Source Image\", Finding FROM Stories;")
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var results []StoryImage

	for rows.Next() {
		var (
			storyID             sql.NullString
			imageJSONBlob       sql.NullString
			sourceImageJSONBlob sql.NullString
			finding             sql.NullString
		)

		if err := rows.Scan(&storyID, &imageJSONBlob, &sourceImageJSONBlob, &finding); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		fmt.Println("Processing story with id", storyID.String)

		var imgArray []Image

		if !imageJSONBlob.Valid {
			fmt.Println("Images JSON blob is null for story", storyID.String)
		} else {
			if err := json.Unmarshal([]byte(imageJSONBlob.String), &imgArray); err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON from Image blob: %w", err)
			}

			if len(imgArray) == 0 {
				fmt.Println("No images found for story from Image", storyID.String)
			}

			if len(imgArray) > 1 {
				fmt.Println("Multiple images found for story from Image", storyID.String, "in total", len(imgArray))
			}

			if !finding.Valid {
				fmt.Println("No finding for story from Image, with id", storyID.String)
			}

			fmt.Println("Processing images in Image for story with id", storyID.String)

			for _, img := range imgArray {
				results = append(results, StoryImage{
					StoryID: storyID.String,
					Image:   img,
					Finding: finding.String,
				})
			}

			fmt.Println("Done processing images in Image for story with id", storyID.String)
		}

		if !sourceImageJSONBlob.Valid {
			fmt.Println("No images found for story from Source Image, for story with id", storyID.String, "skipping")
			continue
		}

		if err := json.Unmarshal([]byte(sourceImageJSONBlob.String), &imgArray); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON from Source Image blob: %w, storyID: %s", err, storyID.String)
		}

		if len(imgArray) == 0 {
			fmt.Println("No images found for story from Source Image", storyID.String)
		}

		if len(imgArray) > 1 {
			fmt.Println("Multiple images found for story from Source Image", storyID.String, "in total", len(imgArray))
		} else {
			fmt.Println("Single image found for story from Source Image", storyID.String)
		}

		if !finding.Valid {
			fmt.Println("No finding for story from Source Image, with id", storyID.String)
		}

		for _, img := range imgArray {
			results = append(results, StoryImage{
				StoryID: storyID.String,
				Image:   img,
				Finding: finding.String,
			})
		}

		fmt.Println("Story with id", storyID.String, "done, completed from Source Image")
	}

	fmt.Println("Found", len(results), "images in total")

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// calculateHash returns a SHA-256 hash of the provided data as a hex string.
// This is used to verify the integrity of downloaded images.
func calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// DownloadFile retrieves an image from the given URL and saves it to the filesystem.
// It downloads the image from the provided URL and saves it to the "images/" directory
// with the specified filename. The storyID is used for logging purposes.
func DownloadFile(storyID string, url, filename string) error {
	var newFilePath = "images/" + filename

	fmt.Printf("Downloading %s\n", filename)

	// Download files
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := os.WriteFile(newFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	fmt.Printf("Downloaded and saved %s to filesystem (hash: %s)\n",
		filename, calculateHash(data))

	return nil
}

func main() {
	dbPath := "airtable-export-new.db"

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Println("Error: airtable-export.db does not exist. Please create it by exporting the Airtable database using the instructions in the README.")
		return
	}

	results, err := FetchImagesFromDB(dbPath)
	if err != nil {
		fmt.Println("Error fetching images:", err)
		return
	}

	fmt.Println("At end of processing, found", len(results), "images in total – beginning download process")

	for _, result := range results {
		err := DownloadFile(result.StoryID, result.Image.URL, result.Image.Filename)

		if err != nil {
			fmt.Println("Error downloading file:", err)
		}
	}
}
