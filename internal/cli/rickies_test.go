package cli

import (
	"testing"

	"github.com/dkelkhoff/connectedCli/internal/rickies"
)

func TestMatchGame(t *testing.T) {
	index := []rickies.GameRef{
		{Name: "2026 March Rickies", Path: "/game/39.json"},
		{Name: "2026 Annual Rickies", Path: "/game/38.json"},
		{Name: "The EUies", Path: "/game/30.json"},
	}
	// Exact (case-insensitive) wins even though it's also a substring of nothing else.
	if g, err := matchGame(index, "the euies"); err != nil || g.Path != "/game/30.json" {
		t.Errorf("exact match failed: %+v err=%v", g, err)
	}
	// Unique substring.
	if g, err := matchGame(index, "march"); err != nil || g.Path != "/game/39.json" {
		t.Errorf("substring match failed: %+v err=%v", g, err)
	}
	// Ambiguous substring.
	if _, err := matchGame(index, "2026"); err == nil {
		t.Error("expected ambiguity error for '2026'")
	}
	// No match.
	if _, err := matchGame(index, "zzz"); err == nil {
		t.Error("expected no-match error")
	}
}

func TestFormatTitles(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{nil, "Man of the People"},
		{[]string{}, "Man of the People"},
		{[]string{"Annual Chairman"}, "Annual Chairman"},
		{[]string{"Annual Chairman", "Attorney General Flexie"}, "Annual Chairman; Attorney General Flexie"},
	}
	for _, c := range cases {
		if got := formatTitles(c.in); got != c.want {
			t.Errorf("formatTitles(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}
