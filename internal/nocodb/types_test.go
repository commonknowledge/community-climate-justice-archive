package nocodb

import (
	"path/filepath"
	"testing"
)

func TestLocalAttachmentPathUsesDedicatedMediaDirectories(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		mimetype string
		want     string
	}{
		{name: "image mime", filename: "photo.jpg", mimetype: "image/jpeg", want: filepath.Join("images", "photo.jpg")},
		{name: "audio mime", filename: "song.mp3", mimetype: "audio/mpeg", want: filepath.Join("audio", "song.mp3")},
		{name: "video mime", filename: "clip.mp4", mimetype: "video/mp4", want: filepath.Join("video", "clip.mp4")},
		{name: "document mime", filename: "guide.pdf", mimetype: "application/pdf", want: filepath.Join("documents", "guide.pdf")},
		{name: "video extension fallback", filename: "clip.mp4", mimetype: "", want: filepath.Join("video", "clip.mp4")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := localAttachmentPath(tt.filename, tt.mimetype); got != tt.want {
				t.Fatalf("localAttachmentPath(%q, %q) = %q, want %q", tt.filename, tt.mimetype, got, tt.want)
			}
		})
	}
}
