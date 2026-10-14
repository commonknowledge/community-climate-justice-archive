package generate

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"community-climate-justice-archive/data"
)

func TestStoryConnectionsResolveThumbnailsAfterImageProcessing(t *testing.T) {
	templateDir, err := filepath.Abs("../../templates")
	if err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	if err := os.CopyFS("templates", os.DirFS(templateDir)); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir("images", 0755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create("images/Connection target.png")
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, image.NewRGBA(image.Rect(0, 0, 16, 8))); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	// Cache the connection before processing, just as WarmCache does for a new
	// upload. Its saved URLs must still point at the original throughout the test.
	target := data.Story{ImageVideoSound: `[{"title":"Connection target.png","mimetype":"image/png"}]`}
	attachment := target.GetStoryImage()
	connection := data.StoryConnection{
		ID:                 "2",
		Finding:            "Connection target",
		URL:                "/stories/connection-target-2.html",
		Image:              attachment.URL,
		ThumbURL:           attachment.ThumbURL,
		AttachmentType:     attachment.FileType,
		AttachmentFilename: attachment.Filename,
	}
	const originalURL = "/images/Connection target.png"
	if connection.ThumbURL != originalURL {
		t.Fatalf("connection was not cached before processing: %q", connection.ThumbURL)
	}
	page := data.StoryPage{Story: data.Story{
		Finding:     "Source story",
		InspiredBy:  []data.StoryConnection{connection},
		HasInspired: []data.StoryConnection{connection},
	}}
	tmpl, err := loadTemplates()
	if err != nil {
		t.Fatal(err)
	}
	checkRenderedConnections := func(wantURL string) {
		t.Helper()
		var html bytes.Buffer
		if err := tmpl.ExecuteTemplate(&html, "story.html", page); err != nil {
			t.Fatal(err)
		}
		if count := strings.Count(html.String(), "background-image: url('"+wantURL+"')"); count != 2 {
			t.Errorf("Inspired by/Has inspired cards reference %q %d times, want 2", wantURL, count)
		}
		if wantURL != originalURL && strings.Contains(html.String(), originalURL) {
			t.Error("rendered connection cards still reference the cached original after processing")
		}
	}

	checkRenderedConnections(originalURL)
	if err := ProcessImages(); err != nil {
		t.Fatal(err)
	}
	checkRenderedConnections("/images/processed/Connection target_thumb.webp")

	// Losing the derivative must retain the original-image fallback, including
	// on builds that skip image processing.
	if err := os.Remove("images/processed/Connection target_thumb.webp"); err != nil {
		t.Fatal(err)
	}
	checkRenderedConnections(originalURL)
}
