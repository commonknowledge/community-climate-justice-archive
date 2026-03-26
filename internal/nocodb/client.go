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
// Important scope note:
// - This cache stores raw NocoDB records.
// - It does NOT cache converted Story structs or parsed templates.
// - Story conversion caching is handled separately in internal/generate during each build run.
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
	"strings"
	"time"
	"unicode"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/config"
	"community-climate-justice-archive/internal/util"

	"github.com/eduardolat/nocodbgo"
)

const defaultRecordsPageSize = 100
const (
	relationshipCountInspiredByField         = "Inspired by"
	relationshipCountHasInspiredField        = "Has inspired"
	relationshipCountContributorsField       = "Contributors"
	relationshipCountPublicContributorsField = "Public Contributors"
)

// Client wraps the NocoDB client with our configuration.
type Client struct {
	client *nocodbgo.Client
	table  *nocodbgo.Table
	// httpClient is reused for all outbound requests to avoid recreating transports.
	httpClient *http.Client
	// cachedRecords stores raw API records, not converted data.Story structs.
	cachedRecords []map[string]interface{}
	// cachedRecordByID provides fast lookup by record ID for cachedRecords.
	cachedRecordByID map[string]map[string]interface{}
	// cacheLoaded indicates whether cachedRecords currently represents a valid complete dataset.
	cacheLoaded bool
	// cacheOnlyMode forces reads to come from disk/in-memory cache and blocks API requests.
	cacheOnlyMode bool
	diskCacheMode bool // If true, allow reading/writing debug-cache-nocodb.json
	fieldIDByTitle map[string]string
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
		client:     client,
		table:      table,
		httpClient: &http.Client{Timeout: 30 * time.Second},
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

	// Try loading from disk cache first (debug mode only)
	if c.diskCacheMode && c.IsDiskCacheAvailable() {
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
	limit := defaultRecordsPageSize
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
	c.rebuildRecordIndex()

	// Now fetch relationships for all records and add to cache
	log.Printf("Fetching relationships for all %d records...", len(allRecords))
	c.fetchAndCacheRelationships(allRecords)

	c.cacheLoaded = true
	log.Printf("Successfully retrieved and cached all %d records from NocoDB with relationships", len(allRecords))

	// Save cache to disk for faster future debugging (debug mode only)
	if c.diskCacheMode {
		if err := c.SaveCacheToDisk(); err != nil {
			log.Printf("Warning: Failed to save cache to disk: %v", err)
		}
	}

	return allRecords, nil
}

// GetRecordByID retrieves a single record by its ID
func (c *Client) GetRecordByID(id string) (map[string]interface{}, error) {
	if c.table == nil {
		return nil, fmt.Errorf("table not initialized")
	}

	// First try to find it in cache if available.
	if c.cacheLoaded {
		if record, found := c.getCachedRecordByID(id); found {
			return record, nil
		}
		return nil, fmt.Errorf("record with ID %s not found in cache", id)
	}

	// If not cached, we can't fetch individual records from the API easily
	// Force cache load and then search in cache
	_, err := c.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to load cache for record search: %w", err)
	}

	// Search in cache again.
	if record, found := c.getCachedRecordByID(id); found {
		return record, nil
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
	c.rebuildRecordIndex()
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
		// Cache-only mode requires disk cache to be enabled.
		c.diskCacheMode = true
		log.Println("NocoDB client set to cache-only mode (disk cache enabled)")
	}
}

// SetDiskCacheMode enables/disables reading and writing the on-disk debug cache.
func (c *Client) SetDiskCacheMode(enabled bool) {
	c.diskCacheMode = enabled
	if enabled {
		log.Println("NocoDB disk cache mode enabled")
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
	resp, err := c.httpClient.Do(req)
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
	c.cachedRecordByID = nil
	c.cacheLoaded = false
	log.Println("NocoDB cache dropped")
}

func (c *Client) rebuildRecordIndex() {
	c.cachedRecordByID = make(map[string]map[string]interface{}, len(c.cachedRecords))
	for _, record := range c.cachedRecords {
		recordID := toString(record["Id"])
		if recordID == "" {
			continue
		}
		c.cachedRecordByID[recordID] = record
	}
}

func (c *Client) getCachedRecordByID(id string) (map[string]interface{}, bool) {
	if c.cachedRecordByID == nil {
		c.rebuildRecordIndex()
	}
	record, found := c.cachedRecordByID[id]
	return record, found
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
	attachmentLookup := buildConnectionAttachmentLookup(records)
	contributorsFieldID := c.getRelationshipFieldIDByTitle(relationshipCountContributorsField)
	publicContributorsFieldID := c.getRelationshipFieldIDByTitle(relationshipCountPublicContributorsField)
	skippedInspiredBy := 0
	skippedHasInspired := 0
	fetchedInspiredBy := 0
	fetchedHasInspired := 0
	skippedContributors := 0
	skippedPublicContributors := 0
	fetchedContributors := 0
	fetchedPublicContributors := 0

	for i, record := range records {
		recordID := toString(record["Id"])
		if recordID == "" {
			continue
		}

		// Only fetch links when count is unknown or greater than zero.
		inspiredBy := []data.StoryConnection{}
		if !isKnownZeroRelationshipCount(record[relationshipCountInspiredByField]) {
			inspiredBy = c.fetchRelationshipDataWithCache(recordID, relationshipFieldInspiredByID, attachmentLookup)
			fetchedInspiredBy++
		} else {
			skippedInspiredBy++
		}
		records[i]["__cached_inspired_by"] = inspiredBy

		// Only fetch links when count is unknown or greater than zero.
		hasInspired := []data.StoryConnection{}
		if !isKnownZeroRelationshipCount(record[relationshipCountHasInspiredField]) {
			hasInspired = c.fetchRelationshipDataWithCache(recordID, relationshipFieldHasInspiredID, attachmentLookup)
			fetchedHasInspired++
		} else {
			skippedHasInspired++
		}
		records[i]["__cached_has_inspired"] = hasInspired

		contributors := []data.Contributor{}
		if contributorsFieldID != "" && !isKnownZeroRelationshipCount(record[relationshipCountContributorsField]) {
			contributors = c.fetchContributorDataWithCache(recordID, contributorsFieldID)
			fetchedContributors++
		} else {
			skippedContributors++
		}
		records[i][cachedContributorsKey] = contributors

		publicContributors := []data.Contributor{}
		if publicContributorsFieldID != "" && !isKnownZeroRelationshipCount(record[relationshipCountPublicContributorsField]) {
			publicContributors = c.fetchContributorDataWithCache(recordID, publicContributorsFieldID)
			fetchedPublicContributors++
		} else {
			skippedPublicContributors++
		}
		records[i][cachedPublicContributorsKey] = publicContributors
	}

	log.Printf(
		"Completed relationship fetch for %d records (inspiredBy: fetched=%d skipped_zero=%d, hasInspired: fetched=%d skipped_zero=%d, contributors: fetched=%d skipped=%d, publicContributors: fetched=%d skipped=%d)",
		len(records),
		fetchedInspiredBy,
		skippedInspiredBy,
		fetchedHasInspired,
		skippedHasInspired,
		fetchedContributors,
		skippedContributors,
		fetchedPublicContributors,
		skippedPublicContributors,
	)
}

func (c *Client) getRelationshipFieldIDByTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return ""
	}
	normalizedTitle := normalizeRelationshipFieldTitle(title)

	if c.fieldIDByTitle == nil {
		c.fieldIDByTitle = make(map[string]string)
	}

	if fieldID, ok := c.fieldIDByTitle[normalizedTitle]; ok {
		return fieldID
	}

	if err := c.populateRelationshipFieldIDMapFromV2(); err != nil {
		if !strings.Contains(err.Error(), "status 404") {
			log.Printf("Warning: Failed loading relationship metadata from v2 endpoint: %v", err)
		}
		if fallbackErr := c.populateRelationshipFieldIDMapFromV1(); fallbackErr != nil {
			log.Printf("Warning: Failed loading relationship metadata from v1 endpoint: %v", fallbackErr)
		}
	}

	fieldID := c.fieldIDByTitle[normalizedTitle]
	if fieldID == "" {
		log.Printf("Warning: Relationship field ID not found for title %q", title)
	}

	return fieldID
}

func (c *Client) populateRelationshipFieldIDMapFromV2() error {
	url := fmt.Sprintf("%s/api/v2/meta/tables/%s/columns", config.AppConfig.NocoDBEndpoint, config.AppConfig.NocoDBTableID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create v2 columns metadata request: %w", err)
	}

	req.Header.Set("xc-token", config.AppConfig.NocoDBAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch v2 columns metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("v2 columns metadata API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read v2 columns metadata response: %w", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to parse v2 columns metadata response: %w", err)
	}

	columnsRaw, ok := payload["list"].([]interface{})
	if !ok {
		columnsRaw, ok = payload["columns"].([]interface{})
	}
	if !ok {
		return fmt.Errorf("unexpected v2 columns metadata payload shape")
	}

	c.storeRelationshipFieldIDs(columnsRaw)
	return nil
}

func (c *Client) populateRelationshipFieldIDMapFromV1() error {
	url := fmt.Sprintf("%s/api/v1/db/meta/tables/%s", config.AppConfig.NocoDBEndpoint, config.AppConfig.NocoDBTableID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create v1 table metadata request: %w", err)
	}

	req.Header.Set("xc-token", config.AppConfig.NocoDBAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch v1 table metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("v1 table metadata API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read v1 table metadata response: %w", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to parse v1 table metadata response: %w", err)
	}

	columnsRaw, ok := payload["columns"].([]interface{})
	if !ok {
		return fmt.Errorf("unexpected v1 table metadata payload shape")
	}

	c.storeRelationshipFieldIDs(columnsRaw)
	return nil
}

func (c *Client) storeRelationshipFieldIDs(columnsRaw []interface{}) {
	for _, rawColumn := range columnsRaw {
		columnMap, ok := rawColumn.(map[string]interface{})
		if !ok {
			continue
		}

		columnTitle := strings.TrimSpace(toString(columnMap["title"]))
		if columnTitle == "" {
			columnTitle = strings.TrimSpace(toString(columnMap["column_name"]))
		}
		if columnTitle == "" {
			columnTitle = strings.TrimSpace(toString(columnMap["columnName"]))
		}

		columnID := strings.TrimSpace(toString(columnMap["fk_column_id"]))
		if columnID == "" {
			columnID = strings.TrimSpace(toString(columnMap["id"]))
		}
		if columnID == "" {
			columnID = strings.TrimSpace(toString(columnMap["Id"]))
		}

		if columnTitle != "" && columnID != "" {
			c.fieldIDByTitle[normalizeRelationshipFieldTitle(columnTitle)] = columnID
		}
	}
}

func normalizeRelationshipFieldTitle(value string) string {
	if value == "" {
		return ""
	}

	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	builder.Grow(len(value))

	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

func (c *Client) fetchContributorDataWithCache(recordID, fieldID string) []data.Contributor {
	url := fmt.Sprintf("%s/api/v2/tables/%s/links/%s/records/%s",
		config.AppConfig.NocoDBEndpoint,
		config.AppConfig.NocoDBTableID,
		fieldID,
		recordID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Warning: Failed to create contributor request for record %s field %s: %v", recordID, fieldID, err)
		return []data.Contributor{}
	}

	req.Header.Set("xc-token", config.AppConfig.NocoDBAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("Warning: Failed to fetch contributor links for record %s field %s: %v", recordID, fieldID, err)
		return []data.Contributor{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Warning: Contributor links API returned status %d for record %s field %s", resp.StatusCode, recordID, fieldID)
		return []data.Contributor{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Warning: Failed to read contributor links response for record %s field %s: %v", recordID, fieldID, err)
		return []data.Contributor{}
	}

	var linkResponse struct {
		List []map[string]interface{} `json:"list"`
	}

	if err := json.Unmarshal(body, &linkResponse); err != nil {
		log.Printf("Warning: Failed to parse contributor links response for record %s field %s: %v", recordID, fieldID, err)
		return []data.Contributor{}
	}

	contributors := make([]data.Contributor, 0, len(linkResponse.List))
	for _, item := range linkResponse.List {
		name := strings.TrimSpace(toString(item["name"]))
		if name == "" {
			name = strings.TrimSpace(toString(item["Name"]))
		}
		if name == "" {
			name = strings.TrimSpace(toString(item["title"]))
		}
		if name == "" {
			name = strings.TrimSpace(toString(item["Title"]))
		}
		if name == "" {
			continue
		}

		email := strings.TrimSpace(toString(item["email"]))
		if email == "" {
			email = strings.TrimSpace(toString(item["Email"]))
		}

		approved := strings.TrimSpace(toString(item["approved"]))
		if approved == "" {
			approved = strings.TrimSpace(toString(item["Approved"]))
		}

		contributors = append(contributors, data.Contributor{
			Name:     name,
			Email:    email,
			Approved: approved,
		})
	}

	return contributors
}

func isKnownZeroRelationshipCount(value interface{}) bool {
	switch v := value.(type) {
	case nil:
		return false
	case int:
		return v == 0
	case int64:
		return v == 0
	case float64:
		return int(v) == 0
	case string:
		if v == "" {
			return false
		}
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return false
		}
		return parsed == 0
	default:
		return false
	}
}

type connectionAttachmentInfo struct {
	image              string
	thumbURL           string
	attachmentType     string
	attachmentFilename string
}

func buildConnectionAttachmentLookup(records []map[string]interface{}) map[string]connectionAttachmentInfo {
	lookup := make(map[string]connectionAttachmentInfo, len(records))

	for _, record := range records {
		storyID := toString(record["Id"])
		if storyID == "" {
			continue
		}

		// Convert each record once to reuse attachment info across all link lookups.
		story, err := NocoDBRecordToStoryWithClient(record, nil)
		if err != nil {
			log.Printf("Warning: Failed to convert record to story %s: %v", storyID, err)
			lookup[storyID] = connectionAttachmentInfo{attachmentType: "none"}
			continue
		}

		info := connectionAttachmentInfo{attachmentType: "none"}
		storyImage := story.GetStoryImage()
		if storyImage.URL != "" {
			info.image = storyImage.URL
			info.thumbURL = storyImage.ThumbURL
			info.attachmentType = "image"
			info.attachmentFilename = storyImage.Filename
			lookup[storyID] = info
			continue
		}

		attachment := story.GetFirstNonImageAttachment()
		if attachment.URL != "" && attachment.IsAudio() {
			info.attachmentType = "audio"
			info.attachmentFilename = attachment.Filename
		} else if attachment.URL != "" && attachment.IsDocument() {
			info.attachmentType = "document"
			info.attachmentFilename = attachment.Filename
		}

		lookup[storyID] = info
	}

	return lookup
}

// fetchRelationshipDataWithCache makes the HTTP call to get relationship data for a single record/field
// and uses a precomputed attachment lookup for connected stories.
func (c *Client) fetchRelationshipDataWithCache(recordID, fieldID string, attachmentLookup map[string]connectionAttachmentInfo) []data.StoryConnection {
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

	resp, err := c.httpClient.Do(req)
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
			storyID := strconv.Itoa(item.Id)
			connection := data.StoryConnection{
				ID:      strconv.Itoa(item.Id),
				Title:   item.Title,
				Finding: item.Title,
				URL:     "/stories/" + util.Slugify(item.Title) + "-" + storyID + ".html",
			}

			if info, found := attachmentLookup[storyID]; found {
				connection.Image = info.image
				connection.ThumbURL = info.thumbURL
				connection.AttachmentType = info.attachmentType
				connection.AttachmentFilename = info.attachmentFilename
			} else {
				connection.AttachmentType = "none"
			}

			connections = append(connections, connection)
		}
	}

	return connections
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
