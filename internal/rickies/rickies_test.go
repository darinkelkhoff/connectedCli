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
