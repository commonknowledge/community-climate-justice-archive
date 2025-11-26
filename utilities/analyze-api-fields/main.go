package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/config"
	"community-climate-justice-archive/internal/nocodb"
)

// FieldAnalysis represents the analysis of a single API field
type FieldAnalysis struct {
	FieldName          string `json:"field_name"`
	FieldType          string `json:"field_type"`
	NocoDBType         string `json:"nocodb_type"`
	NocoDBOptions      string `json:"nocodb_options,omitempty"`
	CurrentlyMapped    bool   `json:"currently_mapped"`
	UsedInStory        bool   `json:"used_in_story"`
	UsedInTemplates    bool   `json:"used_in_templates"`
	SampleValue        string `json:"sample_value"`
	Status             string `json:"status"`
	JSONTag            string `json:"json_tag,omitempty"`
	TemplateFiles      string `json:"template_files,omitempty"`
	ProcessingCategory string `json:"processing_category,omitempty"`
}

// NocoDBField represents a field definition from NocoDB schema
type NocoDBField struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"uidt"`
	Show  *bool  `json:"show,omitempty"`
	Meta  struct {
		Options []struct {
			Title string `json:"title"`
			Color string `json:"color"`
		} `json:"options"`
	} `json:"meta"`
}

// NocoDBTableSchema represents the table schema from NocoDB
type NocoDBTableSchema struct {
	ID      string        `json:"id"`
	Title   string        `json:"title"`
	Columns []NocoDBField `json:"columns"`
	Views   []NocoDBView  `json:"views"`
}

// NocoDBView represents a view configuration
type NocoDBView struct {
	ID      string             `json:"id"`
	Title   string             `json:"title"`
	Type    int                `json:"type"`
	Columns []NocoDBViewColumn `json:"columns"`
}

// NocoDBViewColumn represents column visibility in a view
type NocoDBViewColumn struct {
	ID   string `json:"fk_column_id"`
	Show *bool  `json:"show,omitempty"`
}

// APIFieldAnalyzer handles the analysis of API fields
type APIFieldAnalyzer struct {
	client             *nocodb.Client
	nocoDBFields       map[string]string      // field name -> json tag
	nocoDBSchema       map[string]NocoDBField // field name -> schema info
	hiddenFields       map[string]bool        // field name -> is hidden in view
	internalFields     map[string]bool        // field name -> is internal system field
	unusedFields       map[string]bool        // field name -> is intentionally unused
	storyFields        map[string]bool        // field name -> exists
	templateFieldUsage map[string][]string    // field name -> template files
}

func main() {
	fmt.Println("API Field Analysis Utility")
	fmt.Println("==========================")

	// Load configuration
	config.LoadConfig()

	log.Println("Analyzing NocoDB API fields...")

	// Create analyzer
	analyzer, err := NewAPIFieldAnalyzer()
	if err != nil {
		log.Fatalf("Failed to create analyzer: %v", err)
	}

	// Run analysis
	analyses, err := analyzer.AnalyzeFields()
	if err != nil {
		log.Fatalf("Failed to analyze fields: %v", err)
	}

	// Output results
	fmt.Printf("\nFound %d fields in API response\n\n", len(analyses))

	// Console output
	printConsoleTable(analyses)

	// CSV output
	if err := writeCSV(analyses); err != nil {
		log.Printf("Warning: Failed to write CSV: %v", err)
	} else {
		fmt.Println("\nCSV file written to: api-field-analysis.csv")
	}

	// JSON output
	if err := writeJSON(analyses); err != nil {
		log.Printf("Warning: Failed to write JSON: %v", err)
	} else {
		fmt.Println("JSON file written to: api-field-analysis.json")
	}

	// Summary
	printSummary(analyses, analyzer)
}

// NewAPIFieldAnalyzer creates a new field analyzer
func NewAPIFieldAnalyzer() (*APIFieldAnalyzer, error) {
	// Create NocoDB client
	client, err := nocodb.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create NocoDB client: %w", err)
	}

	analyzer := &APIFieldAnalyzer{
		client:             client,
		nocoDBFields:       make(map[string]string),
		nocoDBSchema:       make(map[string]NocoDBField),
		hiddenFields:       make(map[string]bool),
		internalFields:     make(map[string]bool),
		unusedFields:       make(map[string]bool),
		storyFields:        make(map[string]bool),
		templateFieldUsage: make(map[string][]string),
	}

	// Fetch NocoDB schema
	if err := analyzer.fetchNocoDBSchema(); err != nil {
		return nil, fmt.Errorf("failed to fetch NocoDB schema: %w", err)
	}

	// Analyze current struct definitions
	if err := analyzer.analyzeCurrentStructs(); err != nil {
		return nil, fmt.Errorf("failed to analyze current structs: %w", err)
	}

	// Analyze template usage
	if err := analyzer.analyzeTemplateUsage(); err != nil {
		return nil, fmt.Errorf("failed to analyze template usage: %w", err)
	}

	return analyzer, nil
}

// fetchNocoDBSchema fetches the table schema from NocoDB API
func (a *APIFieldAnalyzer) fetchNocoDBSchema() error {
	url := fmt.Sprintf("%s/api/v1/db/meta/tables/%s",
		config.AppConfig.NocoDBEndpoint,
		config.AppConfig.NocoDBTableID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create schema request: %w", err)
	}

	req.Header.Set("xc-token", config.AppConfig.NocoDBAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch schema: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("schema API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read schema response: %w", err)
	}

	var schema NocoDBTableSchema
	if err := json.Unmarshal(body, &schema); err != nil {
		return fmt.Errorf("failed to parse schema response: %w", err)
	}

	log.Printf("Fetched schema for table '%s' with %d columns", schema.Title, len(schema.Columns))

	// Build schema map
	for _, column := range schema.Columns {
		a.nocoDBSchema[column.Title] = column
	}

	// Fetch view information to detect hidden fields
	if err := a.fetchViewConfiguration(); err != nil {
		log.Printf("Warning: Failed to fetch view configuration: %v", err)
		// Continue without view info - we'll still show all fields
	}

	return nil
}

// fetchViewConfiguration fetches view configuration to detect hidden fields
func (a *APIFieldAnalyzer) fetchViewConfiguration() error {
	// First, get the list of views for this table
	viewsURL := fmt.Sprintf("%s/api/v1/db/meta/tables/%s/views",
		config.AppConfig.NocoDBEndpoint,
		config.AppConfig.NocoDBTableID)

	req, err := http.NewRequest("GET", viewsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create views request: %w", err)
	}

	req.Header.Set("xc-token", config.AppConfig.NocoDBAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch views: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("views API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read views response: %w", err)
	}

	var views struct {
		List []NocoDBView `json:"list"`
	}
	if err := json.Unmarshal(body, &views); err != nil {
		return fmt.Errorf("failed to parse views response: %w", err)
	}

	log.Printf("Found %d views for table", len(views.List))

	// Use the first view (usually the default Grid view) to determine field visibility
	if len(views.List) > 0 {
		defaultView := views.List[0]
		log.Printf("Using view '%s' to determine field visibility", defaultView.Title)

		// Get detailed view configuration
		if err := a.fetchViewColumns(defaultView.ID); err != nil {
			return fmt.Errorf("failed to fetch view columns: %w", err)
		}
	}

	return nil
}

// fetchViewColumns fetches column visibility for a specific view
func (a *APIFieldAnalyzer) fetchViewColumns(viewID string) error {
	url := fmt.Sprintf("%s/api/v1/db/meta/views/%s/columns",
		config.AppConfig.NocoDBEndpoint,
		viewID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create view columns request: %w", err)
	}

	req.Header.Set("xc-token", config.AppConfig.NocoDBAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch view columns: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("view columns API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read view columns response: %w", err)
	}

	var viewColumns struct {
		List []struct {
			ID          string      `json:"id"`
			ColumnID    string      `json:"fk_column_id"`
			Show        interface{} `json:"show"`
			ColumnTitle string      `json:"title"`
		} `json:"list"`
	}
	if err := json.Unmarshal(body, &viewColumns); err != nil {
		return fmt.Errorf("failed to parse view columns response: %w", err)
	}

	// Map column visibility
	hiddenCount := 0
	for _, viewCol := range viewColumns.List {
		// Find the corresponding schema column
		for _, schemaCol := range a.nocoDBSchema {
			if schemaCol.ID == viewCol.ColumnID {
				// Check if field is hidden - handle different types for Show field
				isHidden := false
				switch v := viewCol.Show.(type) {
				case bool:
					isHidden = !v
				case float64:
					isHidden = v == 0
				case int:
					isHidden = v == 0
				case nil:
					isHidden = false // nil defaults to visible
				default:
					log.Printf("Warning: Unknown type for show field: %T", v)
				}

				if isHidden {
					// Only mark as hidden if it's not already categorized as internal
					if !a.isInternalField(schemaCol.Title) {
						a.hiddenFields[schemaCol.Title] = true
						hiddenCount++
					}
				}
				break
			}
		}
	}

	log.Printf("Detected %d hidden fields in the default view", hiddenCount)
	return nil
}

// AnalyzeFields performs the main field analysis
func (a *APIFieldAnalyzer) AnalyzeFields() ([]FieldAnalysis, error) {
	// Fetch just one sample record for field values (much more efficient)
	sampleRecord, err := a.fetchSampleRecord()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sample record: %w", err)
	}

	var analyses []FieldAnalysis

	// Analyze each field from the schema (more comprehensive than just sample record)
	for fieldName := range a.nocoDBSchema {
		// Categorize and exclude internal system fields
		if a.isInternalField(fieldName) {
			a.internalFields[fieldName] = true
			continue
		}

		// Categorize and exclude intentionally unused fields
		if a.isUnusedField(fieldName) {
			a.unusedFields[fieldName] = true
			continue
		}

		// Categorize and exclude hidden fields (turned off in NocoDB view)
		if a.hiddenFields[fieldName] {
			continue
		}

		// Regular fields that go through standard processing
		processingCategory := "standard"

		// Get sample value from the record if available
		var sampleValue interface{}
		var fieldType string
		if value, exists := sampleRecord[fieldName]; exists {
			sampleValue = value
			fieldType = getFieldType(value)
		} else {
			sampleValue = nil
			fieldType = "null"
		}

		// Get NocoDB schema info for this field
		nocoDBType, nocoDBOptions := a.getNocoDBFieldInfo(fieldName)

		analysis := FieldAnalysis{
			FieldName:          fieldName,
			FieldType:          fieldType,
			NocoDBType:         nocoDBType,
			NocoDBOptions:      nocoDBOptions,
			CurrentlyMapped:    a.isFieldMappedInNocoDB(fieldName),
			UsedInStory:        a.isFieldUsedInStory(fieldName),
			UsedInTemplates:    len(a.templateFieldUsage[fieldName]) > 0,
			SampleValue:        truncateValue(sampleValue, 50),
			JSONTag:            a.nocoDBFields[fieldName],
			ProcessingCategory: processingCategory,
		}

		if templateFiles, exists := a.templateFieldUsage[fieldName]; exists {
			analysis.TemplateFiles = strings.Join(templateFiles, ", ")
		}

		// Determine status
		analysis.Status = a.determineFieldStatus(analysis)

		analyses = append(analyses, analysis)
	}

	// Also check for any fields in the sample record that aren't in schema (shouldn't happen but just in case)
	for fieldName, fieldValue := range sampleRecord {
		// Skip if we already analyzed this field from schema
		if _, exists := a.nocoDBSchema[fieldName]; exists {
			continue
		}

		// Skip internal cache fields
		if strings.HasPrefix(fieldName, "__cached_") {
			continue
		}

		log.Printf("Warning: Found field '%s' in sample record but not in schema", fieldName)

		analysis := FieldAnalysis{
			FieldName:       fieldName,
			FieldType:       getFieldType(fieldValue),
			NocoDBType:      "Unknown",
			NocoDBOptions:   "",
			CurrentlyMapped: a.isFieldMappedInNocoDB(fieldName),
			UsedInStory:     a.isFieldUsedInStory(fieldName),
			UsedInTemplates: len(a.templateFieldUsage[fieldName]) > 0,
			SampleValue:     truncateValue(fieldValue, 50),
			JSONTag:         a.nocoDBFields[fieldName],
		}

		if templateFiles, exists := a.templateFieldUsage[fieldName]; exists {
			analysis.TemplateFiles = strings.Join(templateFiles, ", ")
		}

		analysis.Status = a.determineFieldStatus(analysis)
		analyses = append(analyses, analysis)
	}

	// Sort by field name for consistent output
	sort.Slice(analyses, func(i, j int) bool {
		return analyses[i].FieldName < analyses[j].FieldName
	})

	return analyses, nil
}

// fetchSampleRecord fetches just one record for sample values (much more efficient than getting all records)
func (a *APIFieldAnalyzer) fetchSampleRecord() (map[string]interface{}, error) {
	if a.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	log.Printf("Fetching single sample record for field analysis...")

	// Use direct HTTP request to get just one record
	url := fmt.Sprintf("%s/api/v2/tables/%s/records?limit=1",
		config.AppConfig.NocoDBEndpoint,
		config.AppConfig.NocoDBTableID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create sample record request: %w", err)
	}

	req.Header.Set("xc-token", config.AppConfig.NocoDBAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sample record: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("sample record API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sample record response: %w", err)
	}

	var response struct {
		List []map[string]interface{} `json:"list"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse sample record response: %w", err)
	}

	if len(response.List) == 0 {
		return nil, fmt.Errorf("no records found in table")
	}

	log.Printf("Successfully fetched sample record with %d fields", len(response.List[0]))
	return response.List[0], nil
}

// getNocoDBFieldInfo returns the NocoDB field type and options for a field
func (a *APIFieldAnalyzer) getNocoDBFieldInfo(fieldName string) (string, string) {
	if field, exists := a.nocoDBSchema[fieldName]; exists {
		fieldType := field.Type

		// For multi-select fields, include the options
		if fieldType == "MultiSelect" && len(field.Meta.Options) > 0 {
			var options []string
			for _, option := range field.Meta.Options {
				options = append(options, option.Title)
			}
			return fieldType, strings.Join(options, ", ")
		}

		// For single select fields, include the options
		if fieldType == "SingleSelect" && len(field.Meta.Options) > 0 {
			var options []string
			for _, option := range field.Meta.Options {
				options = append(options, option.Title)
			}
			return fieldType, strings.Join(options, ", ")
		}

		return fieldType, ""
	}

	return "Unknown", ""
}

// isInternalField determines if a field is an internal NocoDB system field
func (a *APIFieldAnalyzer) isInternalField(fieldName string) bool {
	// NocoDB internal fields that should be excluded from analysis
	internalPrefixes := []string{
		"nc_", // Most NocoDB internal fields
	}

	// Specific internal fields to exclude
	internalFields := []string{
		"ncRecordHash",
		"ncRecordId",
		// Note: CreatedAt and UpdatedAt are NOT excluded because they are used in templates
	}

	// Check prefixes
	for _, prefix := range internalPrefixes {
		if strings.HasPrefix(fieldName, prefix) {
			return true
		}
	}

	// Check specific fields
	for _, internal := range internalFields {
		if fieldName == internal {
			return true
		}
	}

	return false
}

// isUnusedField determines if a field is programmatically excluded
func (a *APIFieldAnalyzer) isUnusedField(fieldName string) bool {
	// Fields that are programmatically excluded from analysis
	unusedFields := []string{
		"Stories",  // Programmatically excluded - not being used
		"Stories1", // Programmatically excluded - not being used
	}

	for _, unused := range unusedFields {
		if fieldName == unused {
			return true
		}
	}

	return false
}

// analyzeCurrentStructs analyzes the current Go struct definitions
func (a *APIFieldAnalyzer) analyzeCurrentStructs() error {
	// Analyze NocoDBStoryDTO struct using reflection
	nocoDBType := reflect.TypeOf(nocodb.NocoDBStoryDTO{})
	for i := 0; i < nocoDBType.NumField(); i++ {
		field := nocoDBType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			// Remove omitempty and other tag options
			jsonTag = strings.Split(jsonTag, ",")[0]
			a.nocoDBFields[jsonTag] = jsonTag
		}
	}

	// Analyze Story struct
	storyType := reflect.TypeOf(data.Story{})
	for i := 0; i < storyType.NumField(); i++ {
		field := storyType.Field(i)
		a.storyFields[field.Name] = true
	}

	return nil
}

// analyzeTemplateUsage scans template files for field usage
func (a *APIFieldAnalyzer) analyzeTemplateUsage() error {
	templatesDir := "templates"

	return filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process HTML files
		if !strings.HasSuffix(path, ".html") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Look for template variable patterns like {{.FieldName}} or {{.Story.FieldName}}
		patterns := []string{
			`\{\{\s*\.(\w+)\s*\}\}`,         // {{.FieldName}}
			`\{\{\s*\.Story\.(\w+)\s*\}\}`,  // {{.Story.FieldName}}
			`\{\{\s*range\s+\.(\w+)\s*\}\}`, // {{range .FieldName}}
			`\{\{\s*if\s+\.(\w+)\s*\}\}`,    // {{if .FieldName}}
			`\{\{\s*with\s+\.(\w+)\s*\}\}`,  // {{with .FieldName}}
		}

		relPath, _ := filepath.Rel("templates", path)

		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindAllStringSubmatch(string(content), -1)

			for _, match := range matches {
				if len(match) > 1 {
					fieldName := match[1]
					if a.templateFieldUsage[fieldName] == nil {
						a.templateFieldUsage[fieldName] = []string{}
					}
					// Avoid duplicates
					found := false
					for _, existing := range a.templateFieldUsage[fieldName] {
						if existing == relPath {
							found = true
							break
						}
					}
					if !found {
						a.templateFieldUsage[fieldName] = append(a.templateFieldUsage[fieldName], relPath)
					}
				}
			}
		}

		return nil
	})
}

// isFieldMappedInNocoDB checks if field exists in NocoDBStoryDTO
func (a *APIFieldAnalyzer) isFieldMappedInNocoDB(fieldName string) bool {
	_, exists := a.nocoDBFields[fieldName]
	return exists
}

// isFieldUsedInStory checks if field is used in final Story struct
func (a *APIFieldAnalyzer) isFieldUsedInStory(fieldName string) bool {
	// Check direct field name match
	if a.storyFields[fieldName] {
		return true
	}

	// Check common field name variations
	variations := []string{
		strings.Title(fieldName),
		strings.ToLower(fieldName),
		strings.ReplaceAll(fieldName, " ", ""),
		strings.ReplaceAll(strings.Title(fieldName), " ", ""),
	}

	for _, variation := range variations {
		if a.storyFields[variation] {
			return true
		}
	}

	return false
}

// determineFieldStatus determines the status of a field
func (a *APIFieldAnalyzer) determineFieldStatus(analysis FieldAnalysis) string {
	if !analysis.CurrentlyMapped {
		return "NEW"
	}
	if analysis.CurrentlyMapped && analysis.UsedInStory && analysis.UsedInTemplates {
		return "FULLY_MAPPED"
	}
	if analysis.CurrentlyMapped && analysis.UsedInStory {
		return "MAPPED_NOT_DISPLAYED"
	}
	if analysis.CurrentlyMapped {
		return "MAPPED_NOT_USED"
	}
	return "UNKNOWN"
}

// getFieldType determines the type of a field value
func getFieldType(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		return "string"
	case int, int32, int64, float32, float64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		if len(v) > 0 {
			return fmt.Sprintf("array[%s]", getFieldType(v[0]))
		}
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return fmt.Sprintf("unknown(%T)", value)
	}
}

// truncateValue truncates a value for display
func truncateValue(value interface{}, maxLen int) string {
	str := fmt.Sprintf("%v", value)
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-3] + "..."
}

// printConsoleTable prints results as a formatted table
func printConsoleTable(analyses []FieldAnalysis) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Header
	fmt.Fprintln(w, "FIELD NAME\tNOCODB TYPE\tDATA TYPE\tMAPPED\tIN STORY\tIN TEMPLATES\tSTATUS\tSAMPLE VALUE")
	fmt.Fprintln(w, "----------\t-----------\t---------\t------\t--------\t------------\t------\t------------")

	// Data rows
	for _, analysis := range analyses {
		nocoDBDisplay := analysis.NocoDBType
		if analysis.NocoDBOptions != "" {
			nocoDBDisplay += " [" + truncateValue(analysis.NocoDBOptions, 20) + "]"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%v\t%v\t%s\t%s\n",
			analysis.FieldName,
			nocoDBDisplay,
			analysis.FieldType,
			analysis.CurrentlyMapped,
			analysis.UsedInStory,
			analysis.UsedInTemplates,
			analysis.Status,
			analysis.SampleValue,
		)
	}

	w.Flush()
}

// writeCSV writes results to CSV file
func writeCSV(analyses []FieldAnalysis) error {
	file, err := os.Create("api-field-analysis.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	header := []string{
		"Field Name", "NocoDB Type", "NocoDB Options", "Data Type", "Currently Mapped", "Used in Story",
		"Used in Templates", "Status", "JSON Tag", "Template Files", "Sample Value",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Data rows
	for _, analysis := range analyses {
		row := []string{
			analysis.FieldName,
			analysis.NocoDBType,
			analysis.NocoDBOptions,
			analysis.FieldType,
			fmt.Sprintf("%v", analysis.CurrentlyMapped),
			fmt.Sprintf("%v", analysis.UsedInStory),
			fmt.Sprintf("%v", analysis.UsedInTemplates),
			analysis.Status,
			analysis.JSONTag,
			analysis.TemplateFiles,
			analysis.SampleValue,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// writeJSON writes results to JSON file
func writeJSON(analyses []FieldAnalysis) error {
	file, err := os.Create("api-field-analysis.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(analyses)
}

// printSummary prints a summary of the analysis
func printSummary(analyses []FieldAnalysis, analyzer *APIFieldAnalyzer) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("SUMMARY")
	fmt.Println(strings.Repeat("=", 50))

	statusCounts := make(map[string]int)
	for _, analysis := range analyses {
		statusCounts[analysis.Status]++
	}

	fmt.Printf("Total fields found: %d\n", len(analyses))
	for status, count := range statusCounts {
		fmt.Printf("  %s: %d\n", status, count)
	}

	// Show new fields
	var newFields []string
	var multiSelectFields []string
	for _, analysis := range analyses {
		if analysis.Status == "NEW" {
			newFields = append(newFields, analysis.FieldName)
		}
		if analysis.NocoDBType == "MultiSelect" {
			fieldDesc := analysis.FieldName
			if analysis.NocoDBOptions != "" {
				fieldDesc += " [" + analysis.NocoDBOptions + "]"
			}
			multiSelectFields = append(multiSelectFields, fieldDesc)
		}
	}

	if len(newFields) > 0 {
		fmt.Printf("\nNew fields that need attention:\n")
		for _, field := range newFields {
			fmt.Printf("  - %s\n", field)
		}
	}

	if len(multiSelectFields) > 0 {
		fmt.Printf("\nMulti-Select fields (used for tagging):\n")
		for _, field := range multiSelectFields {
			fmt.Printf("  - %s\n", field)
		}
	}

	// Show intentionally hidden fields (user-configured)
	var hiddenFieldsList []string
	for fieldName := range analyzer.hiddenFields {
		hiddenFieldsList = append(hiddenFieldsList, fieldName)
	}

	if len(hiddenFieldsList) > 0 {
		sort.Strings(hiddenFieldsList)
		fmt.Printf("\nFields intentionally hidden in NocoDB view (excluded from analysis):\n")
		for _, field := range hiddenFieldsList {
			fmt.Printf("  - %s\n", field)
		}
	}

	// Show intentionally unused fields
	var unusedFieldsList []string
	for fieldName := range analyzer.unusedFields {
		unusedFieldsList = append(unusedFieldsList, fieldName)
	}

	if len(unusedFieldsList) > 0 {
		sort.Strings(unusedFieldsList)
		fmt.Printf("\nFields programmatically excluded from analysis:\n")
		for _, field := range unusedFieldsList {
			fmt.Printf("  - %s\n", field)
		}
	}

	// Show internal system fields that were excluded
	var internalFieldsList []string
	for fieldName := range analyzer.internalFields {
		internalFieldsList = append(internalFieldsList, fieldName)
	}

	if len(internalFieldsList) > 0 {
		sort.Strings(internalFieldsList)
		fmt.Printf("\nInternal NocoDB system fields (excluded from analysis):\n")
		for _, field := range internalFieldsList {
			fmt.Printf("  - %s\n", field)
		}
	}

	totalExcluded := len(hiddenFieldsList) + len(unusedFieldsList) + len(internalFieldsList)
	if totalExcluded > 0 {
		fmt.Printf("\nNote: %d fields were excluded from analysis (%d hidden, %d programmatically excluded, %d internal).\n",
			totalExcluded, len(hiddenFieldsList), len(unusedFieldsList), len(internalFieldsList))
	}

}
