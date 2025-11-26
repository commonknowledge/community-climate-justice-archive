// Command to consolidate SourceImage field into ImageVideoSound field in NocoDB
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"community-climate-justice-archive/internal/config"
	"community-climate-justice-archive/internal/nocodb"
)

// ConsolidationStats tracks overall statistics for the consolidation operation
type ConsolidationStats struct {
	TotalStories       int
	ProcessedStories   int
	SkippedStories     int
	ErrorStories       int
	StoriesWithChanges int
	TotalAttachments   int
	AddedAttachments   int
	// Checksum mode specific
	StoriesWithMissingFiles int
	TotalMissingFiles       int
}

func main() {
	// Parse command line flags
	var dryRun = flag.Bool("dry-run", false, "Analyze what would be changed without making any writes to NocoDB")
	var testMode = flag.Bool("test", false, "Process one story at a time with user confirmation (writes to NocoDB)")
	var checksumMode = flag.Bool("checksum", false, "Verify consolidation is complete - check all SourceImage files are in ImageVideoSound")
	var verbose = flag.Bool("verbose", false, "Enable verbose logging of all operations")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Consolidates SourceImage field into ImageVideoSound field in NocoDB.\n\n")
		fmt.Fprintf(os.Stderr, "Operation Modes:\n")
		fmt.Fprintf(os.Stderr, "  Default    - Process all stories with batch progress reporting (writes to NocoDB)\n")
		fmt.Fprintf(os.Stderr, "  --dry-run  - Analyze what would be changed without any writes\n")
		fmt.Fprintf(os.Stderr, "  --test     - Process one story at a time with confirmation (writes to NocoDB)\n")
		fmt.Fprintf(os.Stderr, "  --checksum - Verify consolidation is complete and safe to delete SourceImage field\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Validate flags
	modeCount := 0
	if *dryRun {
		modeCount++
	}
	if *testMode {
		modeCount++
	}
	if *checksumMode {
		modeCount++
	}
	if modeCount > 1 {
		log.Fatal("Cannot use multiple operation modes simultaneously")
	}

	// Initialize configuration
	config.LoadConfig()

	// Create NocoDB client
	client, err := nocodb.NewClient()
	if err != nil {
		log.Fatalf("Failed to create NocoDB client: %v", err)
	}

	// Determine operation mode
	var mode string
	var writesToDB bool
	switch {
	case *dryRun:
		mode = "DRY-RUN"
		writesToDB = false
	case *testMode:
		mode = "TEST"
		writesToDB = true
	case *checksumMode:
		mode = "CHECKSUM"
		writesToDB = false
	default:
		mode = "PRODUCTION"
		writesToDB = true
	}

	fmt.Printf("\n=== Image Field Consolidation Tool ===\n")
	fmt.Printf("Mode: %s\n", mode)
	if writesToDB {
		fmt.Printf("WARNING: This will make changes to your NocoDB database!\n")
	} else {
		fmt.Printf("INFO: Running in analysis mode - no changes will be made.\n")
	}
	fmt.Printf("Target: Merge SourceImage → ImageVideoSound\n")
	fmt.Printf("Verbose: %t\n\n", *verbose)

	// Get all records from NocoDB
	fmt.Print("Loading stories from NocoDB... ")
	records, err := client.GetAllRecords()
	if err != nil {
		log.Fatalf("Failed to get records from NocoDB: %v", err)
	}
	fmt.Printf("Loaded %d stories\n\n", len(records))

	// Initialize statistics
	stats := &ConsolidationStats{
		TotalStories: len(records),
	}

	// Process records based on mode
	if *testMode {
		processTestMode(client, records, stats, *verbose)
	} else if *checksumMode {
		processChecksumMode(client, records, stats, *verbose)
	} else {
		processProductionMode(client, records, stats, *dryRun, *verbose)
	}

	// Print final summary
	printFinalSummary(stats, mode)
}

// processTestMode processes stories one at a time with user confirmation
func processTestMode(client *nocodb.Client, records []map[string]interface{}, stats *ConsolidationStats, verbose bool) {
	fmt.Printf("=== TEST MODE: Processing one story at a time ===\n\n")
	reader := bufio.NewReader(os.Stdin)

	for i, record := range records {
		fmt.Printf("--- Story %d/%d ---\n", i+1, len(records))
		
		result, err := nocodb.ConsolidateImageFields(record, client, false) // dryRun=false in test mode
		if err != nil {
			fmt.Printf("ERROR: Error processing story: %v\n", err)
			stats.ErrorStories++
			continue
		}

		printStoryResult(result, verbose)
		updateStats(stats, result)

		if result.HasChanges {
			fmt.Printf("\nThis story has changes to make. Proceed? (y/N/q): ")
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			switch response {
			case "q", "quit":
				fmt.Println("\nQuitting at user request.")
				return
			case "y", "yes":
				fmt.Println("Processing...")
				// Changes already applied since dryRun=false
			default:
				fmt.Println("Skipping...")
				continue
			}
		}

		fmt.Printf("\nContinue to next story? (Y/n/q): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "q" || response == "quit" {
			fmt.Println("\nQuitting at user request.")
			return
		}
		if response == "n" || response == "no" {
			fmt.Println("\nStopping at user request.")
			return
		}

		fmt.Println() // Add spacing between stories
	}
}

// processProductionMode processes all stories in batch mode
func processProductionMode(client *nocodb.Client, records []map[string]interface{}, stats *ConsolidationStats, dryRun, verbose bool) {
	fmt.Printf("=== %s MODE: Processing all stories ===\n\n", map[bool]string{true: "DRY-RUN", false: "PRODUCTION"}[dryRun])

	for i, record := range records {
		// Show progress every 10 stories or for verbose mode
		if verbose || i%10 == 0 || i == len(records)-1 {
			fmt.Printf("Processing story %d/%d...\n", i+1, len(records))
		}

		result, err := nocodb.ConsolidateImageFields(record, client, dryRun)
		if err != nil {
			if verbose {
				fmt.Printf("ERROR: Error processing story %s: %v\n", result.StoryID, err)
			}
			stats.ErrorStories++
			continue
		}

		if verbose || result.HasChanges {
			printStoryResult(result, verbose)
		}
		
		updateStats(stats, result)
	}
}

// processChecksumMode verifies that all SourceImage files are present in ImageVideoSound
func processChecksumMode(client *nocodb.Client, records []map[string]interface{}, stats *ConsolidationStats, verbose bool) {
	fmt.Printf("=== CHECKSUM MODE: Verifying consolidation completion ===\n\n")

	for i, record := range records {
		// Show progress every 10 stories or for verbose mode
		if verbose || i%10 == 0 || i == len(records)-1 {
			fmt.Printf("Checking story %d/%d...\n", i+1, len(records))
		}

		result, err := nocodb.ConsolidateImageFields(record, client, true) // Always dry-run for checksum
		if err != nil {
			if verbose {
				fmt.Printf("ERROR: Error checking story %s: %v\n", result.StoryID, err)
			}
			stats.ErrorStories++
			continue
		}

		// For checksum mode, we care about missing files
		if len(result.SourceAttachments) > 0 && result.HasChanges {
			// This means there are SourceImage files not yet in ImageVideoSound
			stats.StoriesWithMissingFiles++
			stats.TotalMissingFiles += len(result.AddedAttachments)
			
			if verbose {
				fmt.Printf("Story ID: %s\n", result.StoryID)
				fmt.Printf("Title: %s\n", result.StoryTitle)
				fmt.Printf("MISSING: %d files not yet consolidated\n", len(result.AddedAttachments))
				for _, attachment := range result.AddedAttachments {
					fmt.Printf("  - %s\n", attachment.Filename)
				}
				fmt.Println()
			}
		}
		
		updateStats(stats, result)
	}
}

// printStoryResult prints the result of processing a single story
func printStoryResult(result *nocodb.ConsolidationResult, verbose bool) {
	fmt.Printf("Story ID: %s\n", result.StoryID)
	if result.StoryTitle != "" {
		fmt.Printf("Title: %s\n", result.StoryTitle)
	}

	fmt.Printf("SourceImage attachments: %d\n", len(result.SourceAttachments))
	fmt.Printf("ImageVideoSound attachments: %d\n", len(result.TargetAttachments))

	if result.SkipReason != "" {
		fmt.Printf("SKIPPED: %s\n", result.SkipReason)
	} else if result.HasChanges {
		fmt.Printf("CHANGES: Added %d attachments to ImageVideoSound\n", len(result.AddedAttachments))
		if verbose {
			for _, attachment := range result.AddedAttachments {
				fmt.Printf("  + %s\n", attachment.Filename)
			}
		}
	} else {
		fmt.Printf("No changes needed\n")
	}
}

// updateStats updates the consolidation statistics
func updateStats(stats *ConsolidationStats, result *nocodb.ConsolidationResult) {
	stats.ProcessedStories++
	
	if result.SkipReason != "" {
		stats.SkippedStories++
	}
	
	if result.HasChanges {
		stats.StoriesWithChanges++
		stats.AddedAttachments += len(result.AddedAttachments)
	}

	stats.TotalAttachments += len(result.SourceAttachments) + len(result.TargetAttachments)
}

// printFinalSummary prints the final consolidation summary
func printFinalSummary(stats *ConsolidationStats, mode string) {
	fmt.Printf("\n=== CONSOLIDATION SUMMARY (%s MODE) ===\n", mode)
	fmt.Printf("Total stories: %d\n", stats.TotalStories)
	fmt.Printf("Processed stories: %d\n", stats.ProcessedStories)
	fmt.Printf("Skipped stories: %d\n", stats.SkippedStories)
	fmt.Printf("Stories with changes: %d\n", stats.StoriesWithChanges)
	fmt.Printf("Stories with errors: %d\n", stats.ErrorStories)
	fmt.Printf("Total attachments processed: %d\n", stats.TotalAttachments)
	fmt.Printf("Attachments moved from SourceImage to ImageVideoSound: %d\n", stats.AddedAttachments)

	if mode == "DRY-RUN" {
		fmt.Printf("\nNOTE: This was a dry run. To apply changes, run without --dry-run flag.\n")
	} else if mode == "CHECKSUM" {
		fmt.Printf("\n=== CHECKSUM VERIFICATION RESULTS ===\n")
		if stats.StoriesWithMissingFiles == 0 && stats.TotalMissingFiles == 0 {
			fmt.Printf("SUCCESS: All SourceImage files have been consolidated into ImageVideoSound!\n")
			fmt.Printf("SAFE TO DELETE: You can now safely delete the SourceImage field from NocoDB.\n")
		} else {
			fmt.Printf("WARNING: Consolidation is INCOMPLETE!\n")
			fmt.Printf("Stories with missing files: %d\n", stats.StoriesWithMissingFiles)
			fmt.Printf("Total missing files: %d\n", stats.TotalMissingFiles)
			fmt.Printf("DO NOT DELETE: The SourceImage field still contains files not in ImageVideoSound.\n")
			fmt.Printf("Run the consolidation tool without --checksum to complete the process.\n")
		}
	} else if stats.StoriesWithChanges > 0 {
		fmt.Printf("\nSUCCESS: Consolidation completed! %d stories were updated.\n", stats.StoriesWithChanges)
	} else {
		fmt.Printf("\nSUCCESS: All stories were already properly consolidated.\n")
	}
}
