package podsearch

import (
	"os"
	"testing"
)

func TestParseSearchResults(t *testing.T) {
	data, err := os.ReadFile("testdata/search_visionpro.html")
	if err != nil {
		t.Fatal(err)
	}
	hits := parseSearchResults(string(data))
	if len(hits) == 0 {
		t.Fatal("expected hits")
	}
	h := hits[0]
	if h.EpisodeInternalID == 0 {
		t.Errorf("internal id not parsed: %+v", h)
	}
	if h.Seconds == 0 && h.Time == "" {
		t.Errorf("timestamp not parsed: %+v", h)
	}
	if h.URL == "" {
		t.Errorf("url not built: %+v", h)
	}
	if h.Snippet == "" {
		t.Errorf("snippet not parsed: %+v", h)
	}
	if h.EpisodeNumber == 0 {
		t.Errorf("episode number not parsed: %+v", h)
	}
}
