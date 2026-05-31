package podsearch

import (
	"os"
	"strings"
	"testing"
)

func TestParseTranscriptSegments(t *testing.T) {
	data, _ := os.ReadFile("testdata/episode_page.html")
	segs := parseTranscript(string(data))
	if len(segs) == 0 {
		t.Fatal("expected transcript segments")
	}
	if segs[0].Text == "" {
		t.Errorf("segment text empty: %+v", segs[0])
	}
	// Text should be clean prose, not play-control glyphs or raw entities.
	if strings.ContainsAny(segs[0].Text, "◼►") || strings.Contains(segs[0].Text, "&#") {
		t.Errorf("segment text not cleaned: %q", segs[0].Text)
	}
}

func TestTextInRange(t *testing.T) {
	segs := []Segment{
		{Seconds: 0, Text: "a"},
		{Seconds: 60, Text: "b"},
		{Seconds: 120, Text: "c"},
	}
	got := TextInRange(segs, 60, 119)
	if got != "b" {
		t.Errorf("expected 'b', got %q", got)
	}
}

func TestEpisodeNumberFromListing(t *testing.T) {
	data, _ := os.ReadFile("testdata/show_listing.html")
	m := parseShowListing(string(data))
	if len(m) == 0 {
		t.Fatal("expected ep#->id map")
	}
	// Verified pair from the fixture: episode 601 -> internal id 7950.
	if m[601] != 7950 {
		t.Errorf("expected 601->7950, got %d", m[601])
	}
}
