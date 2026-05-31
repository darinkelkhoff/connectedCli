package rickies

import (
	"os"
	"testing"
)

func TestParseWinners(t *testing.T) {
	data, err := os.ReadFile("testdata/winners.json")
	if err != nil {
		t.Fatal(err)
	}
	w, err := ParseWinners(data)
	if err != nil {
		t.Fatal(err)
	}
	if w.Annual.Winner == "" {
		t.Errorf("annual winner missing: %+v", w.Annual)
	}
	if w.Keynote.Winner == "" {
		t.Errorf("keynote winner missing: %+v", w.Keynote)
	}
	if _, ok := w.Titles["Stephen"]; !ok {
		t.Errorf("expected Stephen in titles: %+v", w.Titles)
	}
}

func TestParseEpisodes(t *testing.T) {
	data, err := os.ReadFile("testdata/episodes.json")
	if err != nil {
		t.Fatal(err)
	}
	eps, err := ParseEpisodes(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(eps) == 0 {
		t.Fatal("expected episodes")
	}
	if eps[0].Episode == 0 {
		t.Errorf("episode number missing: %+v", eps[0])
	}
}
