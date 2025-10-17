package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
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
	FieldName       string `json:"field_name"`
	FieldType       string `json:"field_type"`
	CurrentlyMapped bool   `json:"currently_mapped"`
	UsedInStory     bool   `json:"used_in_story"`
	UsedInTemplates bool   `json:"used_in_templates"`
	SampleValue     string `json:"sample_value"`
	Status          string `json:"status"`
	JSONTag         string `json:"json_tag,omitempty"`
	TemplateFiles   string `json:"template_files,omitempty"`
}

// APIFieldAnalyzer handles the analysis of API fields
type APIFieldAnalyzer struct {
	client             *nocodb.Client
	nocoDBFields       map[string]string   // field name -> json tag
	storyFields        map[string]bool     // field name -> exists
	templateFieldUsage map[string][]string // field name -> template files
}

func main() {
	fmt.Println("API Field Analysis Utility")
	fmt.Println("==========================")

	// Load configuration
	config.LoadConfig()

	if !config.AppConfig.UseNocoDB {
		log.Fatal("This utility requires NocoDB to be enabled. Set USE_NOCODB=true in your environment.")
	}

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
	printSummary(analyses)
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
		storyFields:        make(map[string]bool),
		templateFieldUsage: make(map[string][]string),
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

// AnalyzeFields performs the main field analysis
func (a *APIFieldAnalyzer) AnalyzeFields() ([]FieldAnalysis, error) {
	// Fetch sample records from API
	records, err := a.client.GetAllRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no records found in API response")
	}

	// Use first record as sample for field analysis
	sampleRecord := records[0]

	var analyses []FieldAnalysis

	// Analyze each field in the API response
	for fieldName, fieldValue := range sampleRecord {
		// Skip internal cache fields
		if strings.HasPrefix(fieldName, "__cached_") {
			continue
		}

		analysis := FieldAnalysis{
			FieldName:       fieldName,
			FieldType:       getFieldType(fieldValue),
			CurrentlyMapped: a.isFieldMappedInNocoDB(fieldName),
			UsedInStory:     a.isFieldUsedInStory(fieldName),
			UsedInTemplates: len(a.templateFieldUsage[fieldName]) > 0,
			SampleValue:     truncateValue(fieldValue, 50),
			JSONTag:         a.nocoDBFields[fieldName],
		}

		if templateFiles, exists := a.templateFieldUsage[fieldName]; exists {
			analysis.TemplateFiles = strings.Join(templateFiles, ", ")
		}

		// Determine status
		analysis.Status = a.determineFieldStatus(analysis)

		analyses = append(analyses, analysis)
	}

	// Sort by field name for consistent output
	sort.Slice(analyses, func(i, j int) bool {
		return analyses[i].FieldName < analyses[j].FieldName
	})

	return analyses, nil
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
	fmt.Fprintln(w, "FIELD NAME\tTYPE\tMAPPED\tIN STORY\tIN TEMPLATES\tSTATUS\tSAMPLE VALUE")
	fmt.Fprintln(w, "----------\t----\t------\t--------\t------------\t------\t------------")

	// Data rows
	for _, analysis := range analyses {
		fmt.Fprintf(w, "%s\t%s\t%v\t%v\t%v\t%s\t%s\n",
			analysis.FieldName,
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
		"Field Name", "Type", "Currently Mapped", "Used in Story",
		"Used in Templates", "Status", "JSON Tag", "Template Files", "Sample Value",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Data rows
	for _, analysis := range analyses {
		row := []string{
			analysis.FieldName,
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
func printSummary(analyses []FieldAnalysis) {
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
	for _, analysis := range analyses {
		if analysis.Status == "NEW" {
			newFields = append(newFields, analysis.FieldName)
		}
	}

	if len(newFields) > 0 {
		fmt.Printf("\nNew fields that need attention:\n")
		for _, field := range newFields {
			fmt.Printf("  - %s\n", field)
		}
	}

	fmt.Println("\nRecommendations:")
	fmt.Println("  1. Review NEW fields and add them to NocoDBStoryDTO if needed")
	fmt.Println("  2. Check MAPPED_NOT_DISPLAYED fields for potential template usage")
	fmt.Println("  3. Consider removing MAPPED_NOT_USED fields if they're truly unused")
}
