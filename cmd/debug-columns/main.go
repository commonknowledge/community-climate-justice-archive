package main

import (
	"fmt"
	"log"
	"os"

	"community-climate-justice-archive/internal/config"
	"community-climate-justice-archive/internal/nocodb"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Check if NocoDB is enabled
	if !config.AppConfig.UseNocoDB {
		fmt.Println("NocoDB is not enabled (USE_NOCODB=false)")
		fmt.Println("Set USE_NOCODB=true and configure your NocoDB settings to use this tool")
		os.Exit(1)
	}

	// Create NocoDB client
	client, err := nocodb.NewClient()
	if err != nil {
		log.Fatalf("Failed to create NocoDB client: %v", err)
	}

	fmt.Println("NocoDB Table Column Inspector")
	fmt.Println("=============================")
	fmt.Printf("Endpoint: %s\n", config.AppConfig.NocoDBEndpoint)
	fmt.Printf("Table ID: %s\n", config.AppConfig.NocoDBTableID)
	fmt.Println()

	// Get detailed column information
	columns, err := client.GetTableColumns()
	if err != nil {
		log.Fatalf("Failed to get table columns: %v", err)
	}

	fmt.Printf("Found %d columns in the table:\n\n", len(columns))

	// Print table header
	fmt.Printf("%-20s %-40s %s\n", "ID", "Column Name", "Type")
	fmt.Printf("%-20s %-40s %s\n", "--------------------", "----------------------------------------", "--------------------")

	// Print each column as a table row
	for _, col := range columns {
		fmt.Printf("%-20s %-40s %s\n", col.ID, col.Title, col.Type)
	}
}
