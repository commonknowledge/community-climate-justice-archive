package util

import (
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// FormatDate formats a date string to a human readable format.
//
// For example, "2025-03-10T10:00:00Z" becomes "Monday 10 March 2025".
//
// If the date string is empty, it returns an empty string.
//
// The database stores these in accordance with RFC3339 date format.
// See https://www.rfc-editor.org/rfc/rfc3339 for the details of this.
func FormatDate(dateString string) string {
	if dateString == "" {
		return ""
	}

	// Try parsing with RFC3339 format first (SQLite format)
	parsedTime, err := time.Parse(time.RFC3339, dateString)

	if err != nil {
		// Try parsing with NocoDB format: "2006-01-02 15:04:05-07:00"
		parsedTime, err = time.Parse("2006-01-02 15:04:05-07:00", dateString)
	}

	if err != nil {
		// Try parsing with NocoDB format with +00:00: "2006-01-02 15:04:05+00:00"
		parsedTime, err = time.Parse("2006-01-02 15:04:05+00:00", dateString)
	}

	if err != nil {
		// Handle the error
		log.Printf("Failed to parse date: %v", err)
		return dateString // Return original string if parsing fails
	}

	// Go has an interesting way of formatting dates.
	// https://pkg.go.dev/time#Time.Format
	//
	// You provide an example format of dates that you want to display, and it will format the date to match this.
	//
	// For example, "Monday 10 March 2006" as example format will display the date you give it in the same format.
	//
	// However, you need to use the reference time "Mon Jan 2 15:04:05 MST 2006" to structure this. Any other date won't work.
	//
	// This is a nice pragmatic language feature as remembering the formatting of dates in other languages is a pain.
	// For example, to do the same in PHP, you'd have to do this:
	//
	// $date = date('l jS F Y', strtotime($dateString));
	//
	// Which is a pain!
	return parsedTime.Format("Monday 2 January 2006")
}

// Safe way to truncate a string to maxLength characters
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	return s[:maxLength]
}

// Make a string into a slug for use in URLs: downcasing it, removing special characters, and replacing spaces with hyphens.
//
// For example, "Climate Change" becomes "climate-change".
func Slugify(s string) string {
	s = TruncateString(s, 100)

	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace common special characters
	replacements := map[string]string{
		"&":  "and",
		"@":  "at",
		"©":  "c",
		"®":  "r",
		"+":  "plus",
		"£":  "gbp",
		"$":  "usd",
		"€":  "eur",
		"™":  "",
		"'":  "",
		"\"": "",
		"–":  "-", // en dash
		"—":  "-", // em dash
	}

	for old, new := range replacements {
		s = strings.ReplaceAll(s, old, new)
	}

	// Replace any whitespace or punctuation with hyphens
	s = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case unicode.IsSpace(r):
			return '-'
		case r == '-':
			return r
		default:
			return '-'
		}
	}, s)

	// Replace multiple hyphens with single hyphen
	s = regexp.MustCompile(`-+`).ReplaceAllString(s, "-")

	// Trim hyphens from start and end
	return strings.Trim(s, "-")
}
