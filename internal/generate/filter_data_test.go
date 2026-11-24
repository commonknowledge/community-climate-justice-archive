package generate

import (
	"encoding/json"
	"testing"

	"community-climate-justice-archive/data"
	"community-climate-justice-archive/internal/nocodb"
)

func TestFilterDataPreservesCreationAndExperienceDates(t *testing.T) {
	story, err := nocodb.NocoDBRecordToStory(map[string]interface{}{
		"CreatedAt":                   "2026-03-01 10:00:00",
		"Dated created / experienced": "2020-05-01",
	})
	if err != nil {
		t.Fatal(err)
	}

	filterJSON, err := convertStoriesToFilterData([]data.Story{story}, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var result FilterData
	if err := json.Unmarshal([]byte(filterJSON), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Stories) != 1 {
		t.Fatalf("got %d stories, want 1", len(result.Stories))
	}
	if got := result.Stories[0].CreatedTime; got != "2026-03-01 10:00:00" {
		t.Errorf("createdTime = %q, want creation timestamp", got)
	}
	if got := result.Stories[0].StartDateTime; got != "2020-05-01" {
		t.Errorf("startDateTime = %q, want experience date", got)
	}
}
