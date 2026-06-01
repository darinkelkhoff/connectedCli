package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/darinkelkhoff/connectedCli/internal/podsearch"
)

func TestMakeHighlighter(t *testing.T) {
	hl := makeHighlighter("vision pro", true)
	got := hl("the Vision Pro thing")
	if !strings.Contains(got, "\033[1;93mVision Pro"+ansiReset) {
		t.Errorf("term not highlighted (case-insensitive): %q", got)
	}
	// No color => no-op.
	if plain := makeHighlighter("vision pro", false)("the Vision Pro thing"); plain != "the Vision Pro thing" {
		t.Errorf("expected no-op without color, got %q", plain)
	}
}

func TestPrintSearchHitsGroupsByEpisode(t *testing.T) {
	hits := []podsearch.Hit{
		{EpisodeInternalID: 1, EpisodeNumber: 601, EpisodeTitle: "I Love Wrists", Time: "00:01:01", Snippet: "a", URL: "u1"},
		{EpisodeInternalID: 1, EpisodeNumber: 601, EpisodeTitle: "I Love Wrists", Time: "00:02:02", Snippet: "b", URL: "u2"},
		{EpisodeInternalID: 2, EpisodeNumber: 600, EpisodeTitle: "Tommy Siri", Time: "00:03:03", Snippet: "c", URL: "u3"},
	}
	var buf bytes.Buffer // not a terminal => no color/links
	printSearchHits(&buf, "x", hits)
	got := buf.String()
	if !strings.Contains(got, "#601 I Love Wrists:") || !strings.Contains(got, "#600 Tommy Siri:") {
		t.Errorf("episode headers missing:\n%s", got)
	}
	// Episode 601 header should appear once, covering both of its hits.
	if strings.Count(got, "#601 I Love Wrists:") != 1 {
		t.Errorf("601 header should appear once:\n%s", got)
	}
	if !strings.Contains(got, "  00:01:01  … a …") {
		t.Errorf("hit line not formatted as expected:\n%s", got)
	}
}
