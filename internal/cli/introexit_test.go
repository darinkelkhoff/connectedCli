package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestExitShort(t *testing.T) {
	var out bytes.Buffer
	if err := runShortExit(&out); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{"Arrivederci", "Cheerio", "Bye"} {
		if !strings.Contains(got, want) {
			t.Errorf("short exit missing %q: %s", want, got)
		}
	}
}
