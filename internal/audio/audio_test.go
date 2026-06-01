package audio

import (
	"testing"
)

func TestFfmpegAudioOutputArgs(t *testing.T) {
	args := ffmpegAudioOutputArgs()
	if len(args) < 2 {
		t.Fatalf("expected at least 2 args, got %v", args)
	}
	if args[0] != "-f" {
		t.Errorf("expected first arg to be -f, got %q", args[0])
	}
}
