// Command approve-all sets the Approved field to "Yes Live" for all records in NocoDB.
//
// This is a one-time migration tool to mark all existing archive items as approved.
// Run with: go run ./cmd/approve-all
package main

import (
	"fmt"
	"log"

	"community-climate-justice-archive/internal/config"
	"community-climate-justice-archive/internal/nocodb"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Create NocoDB client
	client, err := nocodb.NewClient()
	if err != nil {
		log.Fatalf("Failed to create NocoDB client: %v", err)
	}

	// Get all records
	log.Println("Fetching all records from NocoDB...")
	records, err := client.GetAllRecords()
	if err != nil {
		log.Fatalf("Failed to get records: %v", err)
	}

	log.Printf("Found %d records to update", len(records))

	// Update each record
	successCount := 0
	failCount := 0

	for i, record := range records {
		// Get the record ID
		id := getRecordID(record)
		if id == "" {
			log.Printf("Skipping record %d: no ID found", i)
			failCount++
			continue
		}

		// Update the Approved field
		updateData := map[string]interface{}{
			"Approved": "Yes-Live",
		}

		err := client.UpdateRecord(id, updateData)
		if err != nil {
			log.Printf("Failed to update record %s: %v", id, err)
			failCount++
			continue
		}

		successCount++

		// Log progress every 50 records
		if successCount%50 == 0 {
			log.Printf("Updated %d/%d records...", successCount, len(records))
		}
	}

	log.Printf("Complete. Updated %d records successfully, %d failed.", successCount, failCount)
}

// getRecordID extracts the ID from a record as a string
func getRecordID(record map[string]interface{}) string {
	id := record["Id"]
	if id == nil {
		return ""
	}

	switch v := id.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
