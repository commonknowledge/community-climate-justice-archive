package generate

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"community-climate-justice-archive/data"

	"golang.org/x/image/webp"
)

func TestProcessImagesDetectsFileContents(t *testing.T) {
	for _, tc := range []struct {
		name     string
		filename string
		format   string
	}{
		{"PNG", "upload.png", "png"},
		{"JPEG", "upload.jpg", "jpeg"},
		{"PNG named JPEG", "upload.jpg", "png"},
		{"JPEG named PNG", "upload.png", "jpeg"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Chdir(t.TempDir())

			if err := os.Mkdir("images", 0755); err != nil {
				t.Fatal(err)
			}

			img := image.NewRGBA(image.Rect(0, 0, 16, 8))

			for y := 0; y < 8; y++ {
				for x := 0; x < 16; x++ {
					img.Set(x, y, color.RGBA{R: uint8(x * 16), G: uint8(y * 32), A: 255})
				}
			}

			file, err := os.Create(filepath.Join("images", tc.filename))

			if err != nil {
				t.Fatal(err)
			}

			if tc.format == "png" {
				err = png.Encode(file, img)
			} else {
				err = jpeg.Encode(file, img, nil)
			}

			file.Close()

			if err != nil {
				t.Fatal(err)
			}

			if err := ProcessImages(); err != nil {
				t.Fatal(err)
			}

			if err := CopyImagesToOutput(); err != nil {
				t.Fatal(err)
			}

			for _, variant := range []struct {
				filename string
				width    int
			}{
				{"upload.webp", 16},
				{"upload_thumb.webp", 300},
				{"upload_medium.webp", 800},
				{"upload_large.webp", 1200},
			} {
				file, err := os.Open(filepath.Join("out/images/processed", variant.filename))

				if err != nil {
					t.Errorf("missing published variant %s: %v", variant.filename, err)
					continue
				}

				processed, err := webp.Decode(file)
				file.Close()

				if err != nil {
					t.Errorf("invalid published variant %s: %v", variant.filename, err)
				} else if processed.Bounds().Dx() != variant.width {
					t.Errorf("%s width = %d, want %d", variant.filename, processed.Bounds().Dx(), variant.width)
				}
			}
		})
	}
}

// Missing derivatives must not turn a successfully uploaded image into a broken
// archive thumbnail, hover preview, or story image.
func TestImageAttachmentsStayAvailableWithoutProcessedImages(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.MkdirAll("images", 0755); err != nil {
		t.Fatal(err)
	}

	const filename = "New upload.png"

	file, err := os.Create(filepath.Join("images", filename))

	if err != nil {
		t.Fatal(err)
	}

	if err := png.Encode(file, image.NewRGBA(image.Rect(0, 0, 16, 8))); err != nil {
		t.Fatal(err)
	}

	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	story := data.Story{
		Finding:         "New story",
		ImageVideoSound: `[{"title":"New upload.png","mimetype":"image/png"}]`,
	}

	checkImages := func(wantURLs []string) {
		t.Helper()

		if err := CopyImagesToOutput(); err != nil {
			t.Fatal(err)
		}

		attachment := story.GetStoryAttachment()
		urls := []string{attachment.URL, attachment.ThumbURL, attachment.MediumURL, attachment.LargeURL}

		for i, url := range urls {
			if url != wantURLs[i] {
				t.Errorf("image URL %d = %q, want %q", i, url, wantURLs[i])
			}

			file, err := os.Open(filepath.Join("out", strings.TrimPrefix(url, "/")))

			if err != nil {
				t.Errorf("published image %q is unavailable: %v", url, err)
				continue
			}

			if strings.HasSuffix(url, ".webp") {
				_, err = webp.Decode(file)
			} else {
				_, err = png.Decode(file)
			}

			file.Close()

			if err != nil {
				t.Errorf("published image %q cannot be decoded: %v", url, err)
			}
		}
	}

	// A build that skips image processing must still publish usable images.
	const originalURL = "/images/New upload.png"

	checkImages([]string{originalURL, originalURL, originalURL, originalURL})

	if err := ProcessImages(); err != nil {
		t.Fatal(err)
	}

	processedURLs := []string{
		"/images/processed/New upload.webp",
		"/images/processed/New upload_thumb.webp",
		"/images/processed/New upload_medium.webp",
		"/images/processed/New upload_large.webp",
	}

	checkImages(processedURLs)

	// A partially failed conversion should preserve successful variants and use
	// the original for missing or empty outputs.
	if err := os.Remove("images/processed/New upload_thumb.webp"); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("images/processed/New upload_large.webp", nil, 0644); err != nil {
		t.Fatal(err)
	}

	checkImages([]string{processedURLs[0], originalURL, processedURLs[2], originalURL})
}
