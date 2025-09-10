// Package nocodb provides NocoDB API client functionality
package nocodb

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"community-climate-justice-archive/internal/config"

	"github.com/eduardolat/nocodbgo"
)

// Client wraps the NocoDB client with our configuration
type Client struct {
	client        *nocodbgo.Client
	table         *nocodbgo.Table
	cachedRecords []map[string]interface{} // Cache for all records
	cacheLoaded   bool                     // Flag to track if cache is loaded
}

// NewClient creates a new NocoDB client with configuration from environment variables
func NewClient() (*Client, error) {
	if !config.AppConfig.UseNocoDB {
		return nil, fmt.Errorf("NocoDB is not enabled in configuration")
	}

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
// Uses caching to avoid repeated API calls
func (c *Client) GetAllRecords() ([]map[string]interface{}, error) {
	if c.table == nil {
		return nil, fmt.Errorf("table not initialized")
	}

	// Return cached records if available
	if c.cacheLoaded {
		log.Printf("Returning %d cached records", len(c.cachedRecords))
		return c.cachedRecords, nil
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
	c.cacheLoaded = true

	log.Printf("Successfully retrieved and cached all %d records from NocoDB", len(allRecords))
	return allRecords, nil
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
