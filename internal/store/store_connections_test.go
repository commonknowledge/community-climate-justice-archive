package store

import (
	"fmt"
	"testing"

	"community-climate-justice-archive/data"
)

func TestStoriesWithConnectionsIncludesAllConnectedStories(t *testing.T) {
	stories := make([]data.Story, 0, 52)
	for i := 1; i <= 25; i++ {
		stories = append(stories, data.Story{
			ID:         fmt.Sprintf("source-%d", i),
			InspiredBy: []data.StoryConnection{{ID: fmt.Sprintf("target-%d", i)}},
		})
	}
	for i := 1; i <= 25; i++ {
		stories = append(stories, data.Story{ID: fmt.Sprintf("target-%d", i)})
	}
	stories = append(stories, data.Story{ID: "unconnected"})

	got := storiesWithConnections(stories)
	if len(got) != 50 {
		t.Fatalf("got %d stories, want 50 connected stories and targets", len(got))
	}
	for i, story := range got {
		if story.ID == "unconnected" {
			t.Fatal("included a story without connections")
		}
		if i < 25 && story.ID != fmt.Sprintf("source-%d", i+1) {
			t.Errorf("story at index %d = %q, want source-%d", i, story.ID, i+1)
		}
	}
}
