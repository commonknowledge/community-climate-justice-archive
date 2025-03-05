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
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Image represents an image entry from the JSON data.
type Image struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

// StoryImage associates a story ID with its image data and finding.
type StoryImage struct {
	StoryID string
	Image   Image
	Finding string
}

// SaveImageToDB stores image binary data in the Stories table, maintaining an array of
// images for each story.
func SaveImageToDB(db *sql.DB, storyID string, filename string, data []byte) error {
	// First, get existing images for this story
	var existingData []byte
	err := db.QueryRow("SELECT ImageData FROM Stories WHERE _id = ?", storyID).Scan(&existingData)

	var images [][]byte
	if err == nil && len(existingData) > 0 {
		// Try to unmarshal as JSON array first
		err = json.Unmarshal(existingData, &images)
		if err != nil {
			// If unmarshal fails, treat existing data as a single image
			images = [][]byte{existingData}
		}
	}

	images = append(images, data)

	jsonData, err := json.Marshal(images)
	if err != nil {
		return fmt.Errorf("failed to marshal images to JSON: %w", err)
	}

	_, err = db.Exec(
		"UPDATE Stories SET ImageData = ? WHERE _id = ?",
		jsonData,
		storyID,
	)
	if err != nil {
		return fmt.Errorf("failed to save images to database: %w", err)
	}
	return nil
}

// FetchImagesFromDB retrieves all images from the database and returns them as StoryImage structs.
func FetchImagesFromDB(dbPath string) ([]StoryImage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT _id, Image, Finding FROM Stories")
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var results []StoryImage

	for rows.Next() {
		var (
			storyID  sql.NullString
			jsonBlob sql.NullString
			finding  sql.NullString
		)
		if err := rows.Scan(&storyID, &jsonBlob, &finding); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if !jsonBlob.Valid {
			fmt.Println("Skipping null JSON blob")
			continue
		}

		var imgArray []Image
		if err := json.Unmarshal([]byte(jsonBlob.String), &imgArray); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}

		if !finding.Valid {
			fmt.Println("No finding for story, with id", storyID)
		}

		fmt.Println("Processing story from finding", finding.String)

		for _, img := range imgArray {
			results = append(results, StoryImage{
				StoryID: storyID.String,
				Image:   img,
				Finding: finding.String,
			})
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// calculateHash returns a SHA-256 hash of the provided data as a hex string.
func calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// DownloadFile retrieves an image from the given URL and saves it to both the filesystem
// and database. It skips download if an identical image already exists.
func DownloadFile(db *sql.DB, storyID string, url, filename string) error {
	var newFilePath = "images/" + filename

	// Check if file exists on disk
	diskData, err := os.ReadFile(newFilePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error checking disk file: %w", err)
	}

	// Check database version
	var dbJsonData []byte
	err = db.QueryRow("SELECT ImageData FROM Stories WHERE _id = ?", storyID).Scan(&dbJsonData)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error checking database: %w", err)
	}

	// If we have database data, check if this image exists
	if len(dbJsonData) > 0 {
		var dbImages [][]byte
		if err := json.Unmarshal(dbJsonData, &dbImages); err != nil {
			// If unmarshal fails, treat existing data as a single image
			dbImages = [][]byte{dbJsonData}
		}

		// Check each stored image
		for _, dbImage := range dbImages {
			dbHash := calculateHash(dbImage)
			if len(diskData) > 0 {
				diskHash := calculateHash(diskData)
				if diskHash == dbHash {
					fmt.Printf("Skipping %s (identical hash: %s)\n", filename, diskHash)
					return nil
				}
			}
		}
	}

	fmt.Printf("Downloading %s\n", filename)

	// If we get here, either files don't exist or they're different - download fresh copy
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

	if err := SaveImageToDB(db, storyID, filename, data); err != nil {
		return fmt.Errorf("failed to save to database: %w", err)
	}

	fmt.Printf("Downloaded and saved %s to both DB and filesystem (hash: %s)\n",
		filename, calculateHash(data))
	return nil
}

func main() {
	dbPath := "airtable-export.db"

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	// First, add the ImageData column if it doesn't exist
	_, err = db.Exec("ALTER TABLE Stories ADD COLUMN ImageData BLOB")
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		fmt.Println("Error adding ImageData column:", err)
		return
	}

	results, err := FetchImagesFromDB(dbPath)
	if err != nil {
		fmt.Println("Error fetching images:", err)
		return
	}

	for _, result := range results {
		err := DownloadFile(db, result.StoryID, result.Image.URL, result.Image.Filename)
		if err != nil {
			fmt.Println("Error downloading file:", err)
		}
	}
}
