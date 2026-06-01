package cli

import (
	"testing"
)

func TestFirstSentence(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"Hello and welcome to Connected episode 601.", "Hello and welcome to Connected episode 601."},
		{"First sentence. Second sentence.", "First sentence."},
		{"No punctuation", "No punctuation"},
		{"  Leading spaces. More.", "Leading spaces."},
		{"Ends with question? More text.", "Ends with question?"},
		{"Exclaim! More.", "Exclaim!"},
	}
	for _, c := range cases {
		if got := firstSentence(c.input); got != c.want {
			t.Errorf("firstSentence(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
