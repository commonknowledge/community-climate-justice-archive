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

func TestNormalizeCollageDate(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		required bool
		want     string
		wantErr  bool
	}{
		{name: "date only", value: "2026-10-26", required: true, want: "2026-10-26"},
		{name: "date time", value: "2026-10-26T00:00:00.000Z", required: true, want: "2026-10-26"},
		{name: "optional empty date", value: nil, required: false, want: ""},
		{name: "missing required date", value: nil, required: true, wantErr: true},
		{name: "invalid date", value: "26-10-2026", required: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeCollageDate(tt.value, tt.required)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeCollageDate(%v, %t) error = %v, wantErr %t", tt.value, tt.required, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("normalizeCollageDate(%v, %t) = %q, want %q", tt.value, tt.required, got, tt.want)
			}
		})
	}
}

func TestIsPublishedCollage(t *testing.T) {
	tests := []struct {
		value interface{}
		want  bool
	}{
		{value: true, want: true},
		{value: "true", want: true},
		{value: "Yes", want: true},
		{value: "1", want: true},
		{value: float64(1), want: true},
		{value: int(1), want: true},
		{value: false, want: false},
		{value: "false", want: false},
		{value: float64(0), want: false},
	}

	for _, tt := range tests {
		if got := isPublishedCollage(tt.value); got != tt.want {
			t.Errorf("isPublishedCollage(%v) = %t, want %t", tt.value, got, tt.want)
		}
	}
}
