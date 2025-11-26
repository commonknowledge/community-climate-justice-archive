// Package nocodb provides the NocoDB API client for accessing the archive database.
//
// This file implements the Client struct, which wraps the NocoDB REST API and
// provides methods for reading (and occasionally writing) story data.
//
// The Client handles:
// - Connecting to NocoDB with authentication
// - Fetching all records with pagination
// - Retrieving individual records by ID
// - Filtering records by field values
// - Caching responses for performance
// - Downloading attachment files
//
// Caching Strategy:
// To minimize API calls and improve performance, the client implements two levels
// of caching:
// 1. In-memory cache: Stores fetched records for the duration of the program
// 2. Disk cache: Saves records to a JSON file for faster startup on subsequent runs
//
// The cache includes not just the basic record data, but also relationship
// information (like "inspired by" connections between stories), making the
// application fast even with hundreds of stories.
package nocodb

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/config"
	"community-climate-justice-archive/internal/util"

	"github.com/eduardolat/nocodbgo"
)

// Client wraps the NocoDB client with our configuration
type Client struct {
	client        *nocodbgo.Client
	table         *nocodbgo.Table
	cachedRecords []map[string]interface{}
	cacheLoaded   bool
	cacheOnlyMode bool // If true, only use cache, never hit API
}

// NewClient creates a new NocoDB client with configuration from environment variables
func NewClient() (*Client, error) {

	// Create NocoDB client
	client, err := nocodbgo.NewClient().
		WithBaseURL(config.AppConfig.NocoDBEndpoint).
		WithAPIToken(config.AppConfig.NocoDBAPIKey).
		Create()
	if err != nil {
		return nil, fmt.Errorf("failed to create NocoDB client: %w", err)
	}

	// Get the configured table
	table := client.Table(config.AppConfig.NocoDBTableID)

	log.Printf("NocoDB client created successfully - Endpoint: %s, TableID: %s",
		config.AppConfig.NocoDBEndpoint, config.AppConfig.NocoDBTableID)

	return &Client{
		client: client,
		table:  table,
	}, nil
}

// GetAllRecords retrieves all records from the configured table using pagination
// Uses caching to avoid repeated API calls. Tries disk cache first for faster debugging.
func (c *Client) GetAllRecords() ([]map[string]interface{}, error) {
	if c.table == nil {
		return nil, fmt.Errorf("table not initialized")
	}

	// Return cached records if available
	if c.cacheLoaded {
		return c.cachedRecords, nil
	}

	// Try loading from disk cache first
	if c.IsDiskCacheAvailable() {
		if err := c.LoadCacheFromDisk(); err != nil {
			log.Printf("Warning: Failed to load disk cache: %v", err)
			if c.cacheOnlyMode {
				return nil, fmt.Errorf("cache-only mode enabled but disk cache failed to load: %w", err)
			}
		} else {
			return c.cachedRecords, nil
		}
	}

	// If cache-only mode is enabled and no cache available, fail
	if c.cacheOnlyMode {
		return nil, fmt.Errorf("cache-only mode enabled but no disk cache available")
	}

	// Fetch records from API and cache them
	var allRecords []map[string]interface{}
	limit := 100 // Records per page
	offset := 0

	log.Printf("Starting paginated retrieval of all records from NocoDB (cache miss)...")

	for {
		log.Printf("Fetching records with limit=%d, offset=%d", limit, offset)

		response, err := c.table.ListRecords().
			Limit(limit).
			Offset(offset).
			Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch records from NocoDB at offset %d: %w", offset, err)
		}

		// Add this batch to our collection
		allRecords = append(allRecords, response.List...)

		log.Printf("Retrieved %d records in this batch (total so far: %d)",
			len(response.List), len(allRecords))

		// If we got fewer records than the limit, we've reached the end
		if len(response.List) < limit {
			log.Printf("Reached end of records (batch had %d < %d limit)",
				len(response.List), limit)
			break
		}

		// Move to the next page
		offset += limit
	}

	// Cache the results
	c.cachedRecords = allRecords

	// Now fetch relationships for all records and add to cache
	log.Printf("Fetching relationships for all %d records...", len(allRecords))
	c.fetchAndCacheRelationships(allRecords)

	c.cacheLoaded = true
	log.Printf("Successfully retrieved and cached all %d records from NocoDB with relationships", len(allRecords))

	// Save cache to disk for faster future loading
	if err := c.SaveCacheToDisk(); err != nil {
		log.Printf("Warning: Failed to save cache to disk: %v", err)
	}

	return allRecords, nil
}

// GetRecordByID retrieves a single record by its ID
func (c *Client) GetRecordByID(id string) (map[string]interface{}, error) {
	if c.table == nil {
		return nil, fmt.Errorf("table not initialized")
	}

	// First try to find it in cache if available
	if c.cacheLoaded {
		for _, record := range c.cachedRecords {
			// Try both string and numeric ID comparisons
			if recordID, ok := record["Id"].(string); ok && recordID == id {
				return record, nil
			}
			if recordID, ok := record["Id"].(float64); ok && fmt.Sprintf("%.0f", recordID) == id {
				return record, nil
			}
			if recordID, ok := record["Id"].(int); ok && fmt.Sprintf("%d", recordID) == id {
				return record, nil
			}
		}
		return nil, fmt.Errorf("record with ID %s not found in cache", id)
	}

	// If not cached, we can't fetch individual records from the API easily
	// Force cache load and then search in cache
	_, err := c.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to load cache for record search: %w", err)
	}

	// Search in cache again
	for _, record := range c.cachedRecords {
		// Try both string and numeric ID comparisons
		if recordID, ok := record["Id"].(string); ok && recordID == id {
			return record, nil
		}
		if recordID, ok := record["Id"].(float64); ok && fmt.Sprintf("%.0f", recordID) == id {
			return record, nil
		}
		if recordID, ok := record["Id"].(int); ok && fmt.Sprintf("%d", recordID) == id {
			return record, nil
		}
	}

	return nil, fmt.Errorf("record with ID %s not found", id)
}

// UpdateRecord updates a record by ID in NocoDB with the provided field data
func (c *Client) UpdateRecord(id string, fieldData map[string]interface{}) error {
	if c.table == nil {
		return fmt.Errorf("table not initialized")
	}

	if c.cacheOnlyMode {
		return fmt.Errorf("cannot update record in cache-only mode")
	}

	log.Printf("Updating record %s with fields: %v", id, fieldData)

	// The NocoDB library requires the ID to be included in the data for updates
	updateData := make(map[string]interface{})
	for k, v := range fieldData {
		updateData[k] = v
	}
	updateData["Id"] = id

	// Use the NocoDB library to update the record
	err := c.table.UpdateRecord(updateData).Execute()
	if err != nil {
		return fmt.Errorf("failed to update record %s: %w", id, err)
	}

	log.Printf("Successfully updated record %s", id)
	
	// Invalidate cache since we've made changes
	c.DropCache()
	
	return nil
}

const diskCacheFile = "debug-cache-nocodb.json"

// DiskCacheData represents the structure of the disk cache
type DiskCacheData struct {
	Records   []map[string]interface{} `json:"records"`
	Timestamp time.Time                `json:"timestamp"`
	Count     int                      `json:"count"`
}

// SaveCacheToDisk saves the current cache to disk for faster debugging
func (c *Client) SaveCacheToDisk() error {
	if !c.cacheLoaded || len(c.cachedRecords) == 0 {
		return fmt.Errorf("no cache loaded to save")
	}

	cacheData := DiskCacheData{
		Records:   c.cachedRecords,
		Timestamp: time.Now(),
		Count:     len(c.cachedRecords),
	}

	file, err := os.Create(diskCacheFile)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cacheData); err != nil {
		return fmt.Errorf("failed to write cache data: %w", err)
	}

	log.Printf("Cache saved to disk: %s (%d records)", diskCacheFile, len(c.cachedRecords))
	return nil
}

// LoadCacheFromDisk loads cache from disk if available
func (c *Client) LoadCacheFromDisk() error {
	file, err := os.Open(diskCacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("disk cache file not found: %s", diskCacheFile)
		}
		return fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	var cacheData DiskCacheData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cacheData); err != nil {
		return fmt.Errorf("failed to decode cache data: %w", err)
	}

	c.cachedRecords = cacheData.Records
	c.cacheLoaded = true

	age := time.Since(cacheData.Timestamp)
	log.Printf("Cache loaded from disk: %s (%d records, age: %v)",
		diskCacheFile, len(c.cachedRecords), age.Truncate(time.Second))

	return nil
}

// IsDiskCacheAvailable checks if a disk cache file exists
func (c *Client) IsDiskCacheAvailable() bool {
	_, err := os.Stat(diskCacheFile)
	return !os.IsNotExist(err)
}

// ClearDiskCache removes the disk cache file
func (c *Client) ClearDiskCache() error {
	err := os.Remove(diskCacheFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove disk cache: %w", err)
	}
	log.Printf("Disk cache cleared: %s", diskCacheFile)
	return nil
}

// SetCacheOnlyMode enables cache-only mode where no API calls will be made
func (c *Client) SetCacheOnlyMode(enabled bool) {
	c.cacheOnlyMode = enabled
	if enabled {
		log.Println("NocoDB client set to cache-only mode")
	}
}

// GetFilteredRecords retrieves records filtered by a field containing a value using client-side filtering
// This uses the cached records from GetAllRecords() for fast filtering
func (c *Client) GetFilteredRecords(field, value string) ([]map[string]interface{}, error) {
	log.Printf("Starting client-side filtering for field %s containing %s...", field, value)

	// Get all records from cache
	allRecords, err := c.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to get all records for client-side filtering: %w", err)
	}

	// Filter records client-side
	var filteredRecords []map[string]interface{}
	for _, record := range allRecords {
		if recordContainsValue(record, field, value) {
			filteredRecords = append(filteredRecords, record)
		}
	}

	log.Printf("Client-side filtering found %d records for field %s containing %s",
		len(filteredRecords), field, value)
	return filteredRecords, nil
}

// DownloadAttachment downloads a file from NocoDB using the path from attachment data
func (c *Client) DownloadAttachment(imagePath, outputPath string) error {
	// Construct the full URL
	downloadURL := config.AppConfig.NocoDBEndpoint + "/" + imagePath

	log.Printf("Downloading image from NocoDB: %s -> %s", downloadURL, outputPath)

	// Create HTTP request
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header
	req.Header.Set("xc-token", config.AppConfig.NocoDBAPIKey)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error downloading %s: %d", downloadURL, resp.StatusCode)
	}

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", outputDir, err)
	}

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", outputPath, err)
	}
	defer file.Close()

	// Copy the downloaded content
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", outputPath, err)
	}

	log.Printf("Successfully downloaded image: %s", outputPath)
	return nil
}

// DropCache clears the cached records, forcing fresh retrieval on next call
func (c *Client) DropCache() {
	c.cachedRecords = nil
	c.cacheLoaded = false
	log.Println("NocoDB cache dropped")
}

// recordContainsValue checks if a record's field contains the specified value
// Handles both string fields and JSON array fields
func recordContainsValue(record map[string]interface{}, field, value string) bool {
	fieldValue, exists := record[field]
	if !exists {
		return false
	}

	switch v := fieldValue.(type) {
	case string:
		// For string fields, check if it contains the value
		return containsIgnoreCase(v, value)
	case []interface{}:
		// For array fields, check if any element matches
		for _, item := range v {
			if str, ok := item.(string); ok && containsIgnoreCase(str, value) {
				return true
			}
		}
	case interface{}:
		// Handle other types by converting to string
		str := fmt.Sprintf("%v", v)
		return containsIgnoreCase(str, value)
	}

	return false
}

// containsIgnoreCase checks if a string contains another string (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains check
	// Convert both to lowercase for comparison
	sLower := toLower(s)
	substrLower := toLower(substr)
	return contains(sLower, substrLower)
}

// toLower converts string to lowercase (simplified)
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// fetchAndCacheRelationships fetches relationship data for all records and stores it in the cache
func (c *Client) fetchAndCacheRelationships(records []map[string]interface{}) {
	// NocoDB field IDs for relationships
	inspiredByFieldID := "ccsugv6du8wnisr"
	hasInspiredFieldID := "cilfzk65ypiw6o4"

	for i, record := range records {
		recordID := toString(record["Id"])
		if recordID == "" {
			continue
		}

		// Fetch "Inspired by" relationships
		inspiredBy := c.fetchRelationshipDataWithCache(recordID, inspiredByFieldID, records)
		records[i]["__cached_inspired_by"] = inspiredBy

		// Fetch "Has inspired" relationships
		hasInspired := c.fetchRelationshipDataWithCache(recordID, hasInspiredFieldID, records)
		records[i]["__cached_has_inspired"] = hasInspired

		// Log progress every 50 records
		if (i+1)%50 == 0 {
			log.Printf("Fetched relationships for %d/%d records", i+1, len(records))
		}
	}

	log.Printf("Completed fetching relationships for all %d records", len(records))
}

// fetchRelationshipDataWithCache makes the HTTP call to get relationship data for a single record/field
// and uses the provided records cache to look up images
func (c *Client) fetchRelationshipDataWithCache(recordID, fieldID string, allRecords []map[string]interface{}) []data.StoryConnection {
	url := fmt.Sprintf("%s/api/v2/tables/%s/links/%s/records/%s",
		config.AppConfig.NocoDBEndpoint,
		config.AppConfig.NocoDBTableID,
		fieldID,
		recordID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Warning: Failed to create request for record %s field %s: %v", recordID, fieldID, err)
		return []data.StoryConnection{}
	}

	req.Header.Set("xc-token", config.AppConfig.NocoDBAPIKey)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Warning: Failed to fetch links for record %s field %s: %v", recordID, fieldID, err)
		return []data.StoryConnection{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Warning: Links API returned status %d for record %s field %s", resp.StatusCode, recordID, fieldID)
		return []data.StoryConnection{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Warning: Failed to read response for record %s field %s: %v", recordID, fieldID, err)
		return []data.StoryConnection{}
	}

	var linkResponse struct {
		List []struct {
			Id    int    `json:"Id"`
			Title string `json:"Title"`
		} `json:"list"`
	}

	if err := json.Unmarshal(body, &linkResponse); err != nil {
		log.Printf("Warning: Failed to parse response for record %s field %s: %v", recordID, fieldID, err)
		return []data.StoryConnection{}
	}

	var connections []data.StoryConnection
	for _, item := range linkResponse.List {
		if item.Title != "" {
			connection := data.StoryConnection{
				ID:      strconv.Itoa(item.Id),
				Title:   item.Title,
				Finding: item.Title,
				URL:     "/stories/" + util.Slugify(item.Title) + "-" + strconv.Itoa(item.Id) + ".html",
			}

			// Look up the full story data from provided records to get proper attachment info
			storyIDStr := strconv.Itoa(item.Id)
			if cachedStory, found := c.getStoryFromRecords(storyIDStr, allRecords); found {
				// First try to get an image attachment
				storyImage := cachedStory.GetStoryImage()
				if storyImage.URL != "" {
					connection.Image = storyImage.URL
					connection.ThumbURL = storyImage.ThumbURL
					connection.AttachmentType = "image"
					connection.AttachmentFilename = storyImage.Filename
				} else {
					// No image, check for audio attachment
					audioAttachment := cachedStory.GetFirstNonImageAttachment()
					if audioAttachment.URL != "" && audioAttachment.IsAudio() {
						connection.AttachmentType = "audio"
						connection.AttachmentFilename = audioAttachment.Filename
					} else if audioAttachment.URL != "" && audioAttachment.IsDocument() {
						connection.AttachmentType = "document"
						connection.AttachmentFilename = audioAttachment.Filename
					} else {
						connection.AttachmentType = "none"
					}
				}
			} else {
				connection.AttachmentType = "none"
			}

			connections = append(connections, connection)
		}
	}

	return connections
}

// getStoryImageFromRecords looks up a story by ID in the provided records and returns its image URL
func (c *Client) getStoryImageFromRecords(storyID string, allRecords []map[string]interface{}) string {
	// Use the Story object from provided records to get the proper image URL
	if story, found := c.getStoryFromRecords(storyID, allRecords); found {
		storyImage := story.GetStoryImage()
		return storyImage.URL
	}
	return ""
}

// getStoryFromRecords looks up a story by ID in the provided records and returns a fully converted Story object
func (c *Client) getStoryFromRecords(storyID string, allRecords []map[string]interface{}) (data.Story, bool) {
	// Find the record with matching ID
	for _, record := range allRecords {
		recordID := toString(record["Id"])
		if recordID == storyID {
			// Convert the raw record to a full Story object
			// Note: We pass nil for client to avoid infinite recursion on relationships
			story, err := NocoDBRecordToStoryWithClient(record, nil)
			if err != nil {
				log.Printf("Warning: Failed to convert record to story %s: %v", storyID, err)
				return data.Story{}, false
			}
			return story, true
		}
	}

	return data.Story{}, false
}

// getStoryFromCache looks up a story by ID in the cached records and returns a fully converted Story object
func (c *Client) getStoryFromCache(storyID string) (data.Story, bool) {
	allRecords, err := c.GetAllRecords()
	if err != nil {
		log.Printf("Warning: Failed to get cached records for story %s: %v", storyID, err)
		return data.Story{}, false
	}

	return c.getStoryFromRecords(storyID, allRecords)
}

// toString converts interface{} to string safely
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}
