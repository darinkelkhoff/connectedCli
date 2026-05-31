package chapters

import (
	"os"
	"testing"
)

func TestParseID3Chapters(t *testing.T) {
	data, err := os.ReadFile("testdata/id3_head.bin")
	if err != nil {
		t.Fatal(err)
	}
	chs, err := ParseID3Chapters(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(chs) < 2 {
		t.Fatalf("expected several chapters, got %d", len(chs))
	}
	for i, c := range chs {
		if c.Title == "" {
			t.Errorf("chapter %d has no title: %+v", i, c)
		}
		if c.EndMs < c.StartMs {
			t.Errorf("chapter %d end<start: %+v", i, c)
		}
	}
	// Chapters should be in ascending start order.
	for i := 1; i < len(chs); i++ {
		if chs[i].StartMs < chs[i-1].StartMs {
			t.Errorf("chapters out of order at %d", i)
		}
	}
}
