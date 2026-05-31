package cli

import "testing"

func TestChapterDurationSec(t *testing.T) {
	if d := chapterDurationSec(1000, 4000); d != 3 {
		t.Fatalf("expected 3, got %d", d)
	}
	if d := chapterDurationSec(4000, 1000); d != 0 {
		t.Fatalf("negative range should be 0, got %d", d)
	}
}
