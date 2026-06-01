package rickies

import (
	"os"
	"strings"
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

func TestParseGameIndexFiltersToGames(t *testing.T) {
	data, err := os.ReadFile("testdata/index_small.json")
	if err != nil {
		t.Fatal(err)
	}
	games, err := ParseGameIndex(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 2 {
		t.Fatalf("expected 2 games (picks/topics/coinflips filtered), got %d: %+v", len(games), games)
	}
	for _, g := range games {
		if !strings.HasPrefix(g.Path, "/game/") {
			t.Errorf("non-game entry leaked: %+v", g)
		}
	}
}

func TestParseGame(t *testing.T) {
	data, err := os.ReadFile("testdata/game_39.json")
	if err != nil {
		t.Fatal(err)
	}
	g, err := ParseGame(data)
	if err != nil {
		t.Fatal(err)
	}
	if g.Name == "" || g.GameType == "" {
		t.Errorf("game header not parsed: %+v", g.Name)
	}
	if len(g.MainGame.Picks) == 0 {
		t.Fatal("main game has no picks")
	}
	if g.MainGame.Winner == "" {
		t.Error("main game winner missing")
	}
	p := g.MainGame.Picks[0]
	if p.Host == "" || p.Text == "" {
		t.Errorf("pick not parsed: %+v", p)
	}
	if len(g.MainGame.Scores) == 0 {
		t.Error("main game scores missing")
	}
}

func TestParseBill(t *testing.T) {
	data, err := os.ReadFile("testdata/bill_sample.html")
	if err != nil {
		t.Fatal(err)
	}
	// now=5000 is past rule2's end (2000) but within the others.
	entries := ParseBill(string(data), 5000)
	if len(entries) != 5 {
		t.Fatalf("expected 5 entries (rule2 filtered), got %d: %+v", len(entries), entries)
	}
	if !entries[0].Heading || entries[0].Text != "The Rickies" {
		t.Errorf("first entry should be 'The Rickies' heading: %+v", entries[0])
	}
	for _, e := range entries {
		if strings.Contains(e.Text, "OLD expired") {
			t.Errorf("expired rule leaked: %+v", e)
		}
	}
	if got := FirstParagraph(entries); got != "The Rickies is a game Connected hosts play." {
		t.Errorf("FirstParagraph wrong: %q", got)
	}
	// Entities decoded.
	found := false
	for _, e := range entries {
		if e.Text == "There are two types: Annual & Keynote." {
			found = true
		}
	}
	if !found {
		t.Error("entity-decoded rule not found")
	}
}

func TestParseBillVersions(t *testing.T) {
	data, err := os.ReadFile("testdata/bill_sample.html")
	if err != nil {
		t.Fatal(err)
	}
	versions := ParseBillVersions(string(data))
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d: %+v", len(versions), versions)
	}
	v0 := versions[0]
	if v0.Slug != "annual-2017" || v0.Datetime != "2017-01-04" || v0.Name != "Annual Predictions 2017" {
		t.Errorf("version 0 parsed wrong: %+v", v0)
	}
	if v0.Date != "January 4, 2017" {
		t.Errorf("display date wrong: %q", v0.Date)
	}
	if v0.Unix() <= 0 {
		t.Errorf("Unix() should be positive: %d", v0.Unix())
	}

	// Match by slug, index, and unique substring.
	if v, ok := MatchBillVersion(versions, "keynote-mar-2019"); !ok || v.Index != 1 {
		t.Errorf("slug match failed: %+v ok=%v", v, ok)
	}
	if v, ok := MatchBillVersion(versions, "1"); !ok || v.Slug != "keynote-mar-2019" {
		t.Errorf("index match failed: %+v ok=%v", v, ok)
	}
	if v, ok := MatchBillVersion(versions, "march"); !ok || v.Index != 1 {
		t.Errorf("substring match failed: %+v ok=%v", v, ok)
	}
	if _, ok := MatchBillVersion(versions, "nope"); ok {
		t.Error("expected no match for 'nope'")
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
