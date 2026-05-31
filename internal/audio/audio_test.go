package audio

import (
	"strings"
	"testing"
)

func TestFfplayArgs(t *testing.T) {
	args := ffplayArgs("http://x/a.mp3", 90, 30)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-ss 90") || !strings.Contains(joined, "-t 30") {
		t.Fatalf("args missing seek/duration: %v", args)
	}
	if !strings.Contains(joined, "-nodisp") {
		t.Fatalf("expected -nodisp: %v", args)
	}
}
