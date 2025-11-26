// Package nocodb provides consolidation functionality for merging image fields in NocoDB.
//
// This file contains the logic for the one-time migration that consolidated the
// separate "Image" and "SourceImage" fields into a single "ImageVideoSound" field
// in the NocoDB database.
//
// Historical Context:
// The archive originally stored images in two separate fields, which created
// complexity and confusion. This consolidation tool was created to merge those
// fields into one unified attachment field.
//
// Status: This code is now primarily for reference. The migration has been completed,
// though the tool remains idempotent and can be run safely if needed.
//
// How It Works:
// The consolidation process reads both source fields and the target field,
// intelligently merges their attachments (avoiding duplicates based on filename),
// and writes the combined result back to the ImageVideoSound field. It preserves
// all NocoDB metadata like file paths, dimensions, and MIME types.
package nocodb

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"community-climate-justice-archive/data"
)

// NocoDBAttachment represents an attachment in NocoDB format
type NocoDBAttachment struct {
	URL        string                         `json:"url"`
	Title      string                         `json:"title"`
	MimeType   string                         `json:"mimetype"`
	Size       int                            `json:"size"`
	Width      int                            `json:"width"`
	Height     int                            `json:"height"`
	Path       string                         `json:"path"`
	Thumbnails map[string]data.Thumbnail      `json:"thumbnails"`
}

// ConsolidationResult represents the result of attempting to consolidate image fields for a story
type ConsolidationResult struct {
	StoryID           string
	StoryTitle        string
	SkipReason        string          // If skipped, why?
	SourceAttachments []data.StoryAttachment // Attachments from SourceImage field
	TargetAttachments []data.StoryAttachment // Existing attachments in ImageVideoSound field
	MergedAttachments []data.StoryAttachment // Final merged attachments
	AddedAttachments  []data.StoryAttachment // New attachments that would be added
	HasChanges        bool            // Whether any changes need to be made
}

// ConsolidateImageFields analyzes and potentially consolidates the SourceImage field into ImageVideoSound
// for a single story record. Returns detailed information about what was or would be done.
func ConsolidateImageFields(record map[string]interface{}, client *Client, dryRun bool) (*ConsolidationResult, error) {
	// Extract raw field data directly from the record to preserve NocoDB format
	storyID := toString(record["Id"])
	storyTitle := toString(record["Title"])
	
	result := &ConsolidationResult{
		StoryID:    storyID,
		StoryTitle: storyTitle,
	}

	// Get raw SourceImage data (this comes from NocoDB as well, so preserve it)
	sourceImageRaw := record["Source image"]
	var sourceAttachmentsRaw []interface{}
	if sourceImageRaw != nil {
		if sourceArray, ok := sourceImageRaw.([]interface{}); ok {
			sourceAttachmentsRaw = sourceArray
		}
	}

	// Get raw ImageVideoSound data 
	imageVideoSoundRaw := record["Image / video / sound"]
	var targetAttachmentsRaw []interface{}
	if imageVideoSoundRaw != nil {
		if targetArray, ok := imageVideoSoundRaw.([]interface{}); ok {
			targetAttachmentsRaw = targetArray
		}
	}

	// For display purposes, still parse to StoryAttachment format
	sourceAttachments := parseRawAttachmentsForDisplay(sourceAttachmentsRaw)
	targetAttachments := parseRawAttachmentsForDisplay(targetAttachmentsRaw)
	
	result.SourceAttachments = sourceAttachments
	result.TargetAttachments = targetAttachments

	// If no source attachments, nothing to consolidate
	if len(sourceAttachmentsRaw) == 0 {
		result.SkipReason = "No attachments in SourceImage field"
		result.MergedAttachments = targetAttachments
		result.HasChanges = false
		return result, nil
	}

	// Merge raw attachments, avoiding duplicates by filename
	mergedAttachmentsRaw, addedAttachmentsRaw := mergeRawAttachments(targetAttachmentsRaw, sourceAttachmentsRaw)
	
	// Convert for display
	result.MergedAttachments = parseRawAttachmentsForDisplay(mergedAttachmentsRaw)
	result.AddedAttachments = parseRawAttachmentsForDisplay(addedAttachmentsRaw)
	result.HasChanges = len(addedAttachmentsRaw) > 0

	// If no changes needed, skip
	if !result.HasChanges {
		result.SkipReason = "All SourceImage attachments already present in ImageVideoSound"
		return result, nil
	}

	// If not dry run, perform the actual update using raw data
	if !dryRun {
		err := updateImageVideoSoundFieldRaw(client, storyID, mergedAttachmentsRaw)
		if err != nil {
			return result, fmt.Errorf("failed to update ImageVideoSound field: %w", err)
		}
	}

	return result, nil
}

// parseAttachmentsFromField parses attachments from a JSON field (like Image or SourceImage)
func parseAttachmentsFromField(fieldContent string) ([]data.StoryAttachment, error) {
	if fieldContent == "" {
		return []data.StoryAttachment{}, nil
	}

	var attachments []data.StoryAttachment
	if err := json.Unmarshal([]byte(fieldContent), &attachments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attachment field: %w", err)
	}

	return attachments, nil
}

// parseAttachmentsFromNocoDBField parses attachments from the NocoDB ImageVideoSound field format
func parseAttachmentsFromNocoDBField(fieldContent string) ([]data.StoryAttachment, error) {
	if fieldContent == "" {
		return []data.StoryAttachment{}, nil
	}

	// Parse the NocoDB attachment format
	var nocoAttachments []NocoDBAttachment
	if err := json.Unmarshal([]byte(fieldContent), &nocoAttachments); err != nil {
		return nil, fmt.Errorf("failed to unmarshal NocoDB attachment field: %w", err)
	}

	// Convert to StoryAttachment format
	var attachments []data.StoryAttachment
	for _, nocoAttachment := range nocoAttachments {
		attachment := convertNocoDBAttachment(nocoAttachment)
		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

// convertNocoDBAttachment converts a NocoDB attachment to StoryAttachment format
func convertNocoDBAttachment(noco NocoDBAttachment) data.StoryAttachment {
	return data.StoryAttachment{
		Filename:        noco.Title,
		AlternativeText: "",
		Type:            noco.MimeType,
		FileType:        determineFileType(noco.MimeType),
		Size:            noco.Size,
		Width:           noco.Width,
		Height:          noco.Height,
		URL:             noco.URL,
		ThumbURL:        "",
		MediumURL:       "",
		LargeURL:        "",
		Thumbnails:      noco.Thumbnails,
	}
}

// determineFileType determines the file type category from MIME type
func determineFileType(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	default:
		return "document"
	}
}

// mergeAttachments combines target and source attachments, avoiding duplicates based on filename similarity
func mergeAttachments(targetAttachments, sourceAttachments []data.StoryAttachment) (merged, added []data.StoryAttachment) {
	// Start with all target attachments
	merged = make([]data.StoryAttachment, len(targetAttachments))
	copy(merged, targetAttachments)

	// Add source attachments that don't already exist
	for _, sourceAttachment := range sourceAttachments {
		if !isDuplicateAttachment(sourceAttachment, targetAttachments) {
			merged = append(merged, sourceAttachment)
			added = append(added, sourceAttachment)
		}
	}

	return merged, added
}

// isDuplicateAttachment checks if an attachment already exists in the target list based on filename similarity
func isDuplicateAttachment(attachment data.StoryAttachment, existingAttachments []data.StoryAttachment) bool {
	for _, existing := range existingAttachments {
		if areFilenamesSimilar(attachment.Filename, existing.Filename) {
			return true
		}
	}
	return false
}

// areFilenamesSimilar determines if two filenames are similar enough to be considered the same file
func areFilenamesSimilar(filename1, filename2 string) bool {
	if filename1 == "" || filename2 == "" {
		return false
	}

	// Exact match
	if filename1 == filename2 {
		return true
	}

	// Extract base names without extensions
	base1 := strings.TrimSuffix(filepath.Base(filename1), filepath.Ext(filename1))
	base2 := strings.TrimSuffix(filepath.Base(filename2), filepath.Ext(filename2))

	// Compare base names (case insensitive)
	if strings.EqualFold(base1, base2) {
		return true
	}

	// Check if one filename contains the other (for cases where one might have extra suffixes)
	lower1 := strings.ToLower(base1)
	lower2 := strings.ToLower(base2)
	
	if strings.Contains(lower1, lower2) || strings.Contains(lower2, lower1) {
		// Only consider similar if the difference is reasonable (not too different in length)
		lengthDiff := abs(len(lower1) - len(lower2))
		maxLength := max(len(lower1), len(lower2))
		if maxLength > 0 && lengthDiff <= maxLength/3 { // Allow up to 33% difference
			return true
		}
	}

	return false
}

// updateImageVideoSoundFieldRaw updates the ImageVideoSound field in NocoDB with raw attachment data
func updateImageVideoSoundFieldRaw(client *Client, storyID string, mergedAttachmentsRaw []interface{}) error {
	// Update the record with raw attachment data (preserves all NocoDB-specific fields)
	fieldData := map[string]interface{}{
		"Image / video / sound": mergedAttachmentsRaw,
	}

	return client.UpdateRecord(storyID, fieldData)
}

// parseRawAttachmentsForDisplay converts raw NocoDB attachments to StoryAttachment format for display
func parseRawAttachmentsForDisplay(rawAttachments []interface{}) []data.StoryAttachment {
	var attachments []data.StoryAttachment
	
	for _, rawAttachment := range rawAttachments {
		if attachmentMap, ok := rawAttachment.(map[string]interface{}); ok {
			attachment := data.StoryAttachment{
				Filename: toString(attachmentMap["title"]),
				Type:     toString(attachmentMap["mimetype"]),
				Size:     toInt(attachmentMap["size"]),
				Width:    toInt(attachmentMap["width"]),
				Height:   toInt(attachmentMap["height"]),
				URL:      toString(attachmentMap["url"]),
			}
			
			if attachment.Type == "" {
				attachment.Type = "application/octet-stream"
			}
			
			// Set FileType based on mimetype
			if strings.HasPrefix(attachment.Type, "image/") {
				attachment.FileType = "image"
			} else if strings.HasPrefix(attachment.Type, "audio/") {
				attachment.FileType = "audio"
			} else {
				attachment.FileType = "document"
			}
			
			attachments = append(attachments, attachment)
		}
	}
	
	return attachments
}

// mergeRawAttachments merges two raw attachment arrays, avoiding duplicates by filename
func mergeRawAttachments(targetRaw, sourceRaw []interface{}) (merged, added []interface{}) {
	// Start with all target attachments
	merged = make([]interface{}, len(targetRaw))
	copy(merged, targetRaw)
	
	// Extract existing filenames for duplicate checking
	existingFilenames := make(map[string]bool)
	for _, rawAttachment := range targetRaw {
		if attachmentMap, ok := rawAttachment.(map[string]interface{}); ok {
			filename := toString(attachmentMap["title"])
			if filename != "" {
				existingFilenames[strings.ToLower(filename)] = true
			}
		}
	}
	
	// Add source attachments that don't already exist
	for _, rawAttachment := range sourceRaw {
		if attachmentMap, ok := rawAttachment.(map[string]interface{}); ok {
			filename := toString(attachmentMap["title"])
			if filename != "" && !existingFilenames[strings.ToLower(filename)] {
				merged = append(merged, rawAttachment)
				added = append(added, rawAttachment)
				existingFilenames[strings.ToLower(filename)] = true
			}
		}
	}
	
	return merged, added
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// toInt safely converts interface{} to int
func toInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case string:
		if i, err := fmt.Sscanf(val, "%d", new(int)); err == nil && i == 1 {
			var result int
			fmt.Sscanf(val, "%d", &result)
			return result
		}
	}
	return 0
}
