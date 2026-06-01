package feed

import (
	"os"
	"testing"
)

func loadFixture(t *testing.T) []Episode {
	t.Helper()
	data, err := os.ReadFile("testdata/feed_sample.xml")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	eps, err := Parse(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return eps
}

func TestParseExtractsEpisodes(t *testing.T) {
	eps := loadFixture(t)
	if len(eps) == 0 {
		t.Fatal("expected episodes")
	}
	e := eps[0]
	if e.Number == 0 {
		t.Errorf("episode number not parsed: %+v", e)
	}
	if e.Title == "" {
		t.Errorf("title not parsed: %+v", e)
	}
	if e.MP3URL == "" {
		t.Errorf("mp3 url not parsed: %+v", e)
	}
	if e.Date.IsZero() {
		t.Errorf("date not parsed (RSS uses GMT-style pubDate): %+v", e)
	}
}

func TestLatestIsHighestNumbered(t *testing.T) {
	eps := loadFixture(t)
	latest := Latest(eps)
	for _, e := range eps {
		if e.Number > latest.Number {
			t.Fatalf("Latest returned %d but %d exists", latest.Number, e.Number)
		}
	}
}

func TestByNumber(t *testing.T) {
	eps := loadFixture(t)
	target := eps[0].Number
	got, ok := ByNumber(eps, target)
	if !ok || got.Number != target {
		t.Fatalf("ByNumber(%d) failed: %+v ok=%v", target, got, ok)
	}
	if _, ok := ByNumber(eps, 999999); ok {
		t.Fatal("ByNumber should miss for nonexistent episode")
	}
}
