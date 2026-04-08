// Package util contains helpful functions for working with text throughout the archive.
//
// This file has three main jobs:
//
// 1. Making dates look nice - turning "2025-03-10T10:00:00Z" into "Monday 10 March 2025"
// 2. Making sure text isn't too long - cutting it down when needed
// 3. Making URLs from titles - turning "Climate Change" into "climate-change"
//
// These are used all over the place whenever we need to format text for the website.
// By keeping them all here, everywhere in the archive formats things the same way.
package util

import (
	"html"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var (
	markdownHeading1Re = regexp.MustCompile(`^#\s+(.+)$`)
	markdownHeading2Re = regexp.MustCompile(`^##\s+(.+)$`)
	markdownHeading3Re = regexp.MustCompile(`^###\s+(.+)$`)
	htmlBreakTagRe     = regexp.MustCompile(`(?i)<br\s*/?>`)
	markdownCodeRe     = regexp.MustCompile("`([^`]+)`")
	markdownBoldARe    = regexp.MustCompile(`\*\*(.+?)\*\*`)
	markdownBoldBRe    = regexp.MustCompile(`__(.+?)__`)
	markdownItalicARe  = regexp.MustCompile(`\*(.+?)\*`)
	markdownItalicBRe  = regexp.MustCompile(`_(.+?)_`)
	markdownLinkRe     = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)
)

// FormatDate takes a date from the database and makes it nice to read.
//
// Databases store dates in a technical format like "2025-03-10T10:00:00Z", but we
// want to show visitors something friendly like "Monday 10 March 2025". That's what
// this function does.
//
// It's a bit flexible - it tries a few different date formats that NocoDB might
// send us, so it should work even if the database changes how it stores dates.
//
// If the date is empty or something goes wrong, it'll just return what it was given
// rather than breaking everything.
//
// Here's what happens inside:
// - First, try the standard RFC3339 format (the most common one)
// - If that doesn't work, try NocoDB's specific format
// - If that still doesn't work, log a warning and return the original
// - If it works, convert it to "Monday 10 March 2025" format
//
// A nice thing about Go's date formatting: you show it an example of what you want
// using a reference date, rather than remembering cryptic codes. Much easier!
func FormatDate(dateString string) string {
	if dateString == "" {
		return ""
	}

	// Try parsing with RFC3339 format first
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

// TruncateString cuts a string down if it's too long.
//
// Sometimes we need to make sure text isn't too long - like when creating URLs
// or displaying previews. This function just chops it at a maximum length.
//
// If the text is already short enough, it gets returned as-is. If it's too long,
// you get the first maxLength characters.
//
// Simple as that!
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	return s[:maxLength]
}

// Slugify turns any text into something safe to use in a URL.
//
// When we create web pages, we need URLs like "/stories/climate-change.html" but
// our story titles might be "Climate Change!" or "Art, Music & Dance". This function
// does the conversion.
//
// Here's what it does step by step:
// 1. Makes everything lowercase
// 2. Swaps special characters for normal ones (& becomes "and", £ becomes "gbp")
// 3. Removes quotes and trademark symbols that don't belong in URLs
// 4. Turns spaces and punctuation into hyphens
// 5. Removes any double hyphens (-- becomes -)
// 6. Trims hyphens from the start and end
// 7. Cuts it down to 100 characters max (URLs shouldn't be too long)
//
// So "Climate Change!" becomes "climate-change"
// And "Community & Nature" becomes "community-and-nature"
//
// This runs every time we create a URL from a title, theme name, or tag.
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

// MarkdownToHTML converts a limited subset of markdown to HTML.
//
// Supported blocks: headings (#, ##, ###), paragraphs, and unordered lists.
// Supported inline styles: links, bold, italic, and inline code.
func MarkdownToHTML(markdown string) string {
	// NocoDB sometimes stores manual HTML line breaks inside text.
	// Convert them to newlines so markdown paragraph handling can format them.
	markdown = htmlBreakTagRe.ReplaceAllString(markdown, "\n")
	markdown = strings.TrimSpace(strings.ReplaceAll(markdown, "\r\n", "\n"))
	if markdown == "" {
		return ""
	}

	lines := strings.Split(markdown, "\n")
	var out strings.Builder
	var paragraph []string
	inList := false

	flushParagraph := func() {
		if len(paragraph) == 0 {
			return
		}
		out.WriteString("<p>")
		out.WriteString(renderInlineMarkdown(strings.Join(paragraph, " ")))
		out.WriteString("</p>")
		paragraph = nil
	}

	closeList := func() {
		if inList {
			out.WriteString("</ul>")
			inList = false
		}
	}

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			flushParagraph()
			closeList()
			continue
		}

		if matches := markdownHeading3Re.FindStringSubmatch(line); len(matches) == 2 {
			flushParagraph()
			closeList()
			out.WriteString("<h3>")
			out.WriteString(renderInlineMarkdown(matches[1]))
			out.WriteString("</h3>")
			continue
		}

		if matches := markdownHeading2Re.FindStringSubmatch(line); len(matches) == 2 {
			flushParagraph()
			closeList()
			out.WriteString("<h2>")
			out.WriteString(renderInlineMarkdown(matches[1]))
			out.WriteString("</h2>")
			continue
		}

		if matches := markdownHeading1Re.FindStringSubmatch(line); len(matches) == 2 {
			flushParagraph()
			closeList()
			out.WriteString("<h1>")
			out.WriteString(renderInlineMarkdown(matches[1]))
			out.WriteString("</h1>")
			continue
		}

		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			flushParagraph()
			if !inList {
				out.WriteString("<ul>")
				inList = true
			}
			itemText := strings.TrimSpace(line[2:])
			out.WriteString("<li>")
			out.WriteString(renderInlineMarkdown(itemText))
			out.WriteString("</li>")
			continue
		}

		closeList()
		paragraph = append(paragraph, line)
	}

	flushParagraph()
	closeList()

	return out.String()
}

func renderInlineMarkdown(input string) string {
	escaped := html.EscapeString(input)

	escaped = markdownCodeRe.ReplaceAllString(escaped, "<code>$1</code>")
	escaped = markdownBoldARe.ReplaceAllString(escaped, "<strong>$1</strong>")
	escaped = markdownBoldBRe.ReplaceAllString(escaped, "<strong>$1</strong>")
	escaped = markdownItalicARe.ReplaceAllString(escaped, "<em>$1</em>")
	escaped = markdownItalicBRe.ReplaceAllString(escaped, "<em>$1</em>")

	escaped = markdownLinkRe.ReplaceAllStringFunc(escaped, func(match string) string {
		matches := markdownLinkRe.FindStringSubmatch(match)
		if len(matches) != 3 {
			return match
		}

		label := matches[1]
		rawURL := html.UnescapeString(matches[2])
		safeURL := sanitizeURL(rawURL)
		if safeURL == "" {
			return label
		}

		return `<a href="` + html.EscapeString(safeURL) + `" target="_blank" rel="noopener noreferrer">` + label + `</a>`
	})

	return escaped
}

func sanitizeURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return ""
	}

	if parsed.Scheme == "" {
		if strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "#") {
			return trimmed
		}
		return ""
	}

	switch strings.ToLower(parsed.Scheme) {
	case "http", "https", "mailto":
		return trimmed
	default:
		return ""
	}
}
