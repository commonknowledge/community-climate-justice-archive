package util

import (
	"regexp"
	"strings"
	"unicode"
)

// Safe way to truncate a string to maxLength characters
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	return s[:maxLength]
}

// Slugify a string, downcasing it, removing special characters, and replacing spaces with hyphens.
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
