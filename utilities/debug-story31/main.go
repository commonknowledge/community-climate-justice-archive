package main

import (
	"encoding/json"
	"fmt"
	"log"

	"community-climate-justice-archive/internal/config"
	"community-climate-justice-archive/internal/store"
)

func main() {
	config.LoadConfig()
	if err := store.Initialize(); err != nil {
		log.Fatal(err)
	}

	stories := store.GetAllStories()
	for _, story := range stories {
		if story.ID != "31" {
			continue
		}

		fmt.Printf("story_id=%s\n", story.ID)
		fmt.Printf("finding=%s\n", story.Finding)
		fmt.Printf("contributors_len=%d public_contributors_len=%d\n", len(story.Contributors), len(story.PublicContributors))

		contributorsJSON, _ := json.MarshalIndent(story.Contributors, "", "  ")
		publicContributorsJSON, _ := json.MarshalIndent(story.PublicContributors, "", "  ")
		fmt.Printf("contributors=%s\n", string(contributorsJSON))
		fmt.Printf("public_contributors=%s\n", string(publicContributorsJSON))
		return
	}

	fmt.Println("story 31 not found")
}
