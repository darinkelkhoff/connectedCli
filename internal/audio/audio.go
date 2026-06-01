package audio

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ffplayArgs builds args to stream a segment: from startSec for durSec seconds.
func ffplayArgs(url string, startSec, durSec int) []string {
	return []string{
		"-nodisp", "-autoexit", "-loglevel", "error",
		"-ss", strconv.Itoa(startSec),
		"-t", strconv.Itoa(durSec),
		url,
	}
}

// PlaySegment streams [startSec, startSec+durSec) of url via ffplay.
func PlaySegment(ctx context.Context, url string, startSec, durSec int) error {
	if _, err := exec.LookPath("ffplay"); err != nil {
		return errors.New("ffplay not found — install ffmpeg (brew install ffmpeg) to use --play")
	}
	cmd := exec.CommandContext(ctx, "ffplay", ffplayArgs(url, startSec, durSec)...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

// Play streams an entire URL from the start via ffplay.
func Play(ctx context.Context, url string) error {
	if _, err := exec.LookPath("ffplay"); err != nil {
		return errors.New("ffplay not found — install ffmpeg (brew install ffmpeg) to use --play")
	}
	cmd := exec.CommandContext(ctx, "ffplay", "-nodisp", "-autoexit", "-loglevel", "error", url)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

// Say speaks text via the macOS `say` command.
func Say(ctx context.Context, text string) error {
	if _, err := exec.LookPath("say"); err != nil {
		return errors.New("`say` not found — --say requires macOS")
	}
	cmd := exec.CommandContext(ctx, "say")
	cmd.Stdin = strings.NewReader(text)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("say: %w", err)
	}
	return nil
}
