package audio

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func lookPath(names ...string) (string, bool) {
	for _, name := range names {
		if p, err := exec.LookPath(name); err == nil {
			return p, true
		}
	}
	return "", false
}

// ffmpegAudioOutputArgs returns the -f <format> <device> args for direct audio
// output via ffmpeg on the current OS.
func ffmpegAudioOutputArgs() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"-f", "audiotoolbox", "default"}
	default:
		return []string{"-f", "alsa", "default"}
	}
}

// PlaySegment streams [startSec, startSec+durSec) of url.
// Tries ffplay first, then ffmpeg.
func PlaySegment(ctx context.Context, url string, startSec, durSec int) error {
	start, dur := strconv.Itoa(startSec), strconv.Itoa(durSec)

	if _, ok := lookPath("ffplay"); ok {
		cmd := exec.CommandContext(ctx, "ffplay",
			"-nodisp", "-autoexit", "-loglevel", "error",
			"-ss", start, "-t", dur, url)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		return cmd.Run()
	}

	if _, ok := lookPath("ffmpeg"); ok {
		args := append([]string{"-loglevel", "error", "-ss", start, "-i", url, "-t", dur}, ffmpegAudioOutputArgs()...)
		cmd := exec.CommandContext(ctx, "ffmpeg", args...)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		return cmd.Run()
	}

	return errors.New("ffplay/ffmpeg not found — install ffmpeg (brew install ffmpeg) to use --play")
}

// Play streams an entire URL from the start.
// Tries afplay first (macOS built-in), then ffplay, then ffmpeg.
func Play(ctx context.Context, url string) error {
	if _, ok := lookPath("afplay"); ok {
		cmd := exec.CommandContext(ctx, "afplay", url)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		return cmd.Run()
	}

	if _, ok := lookPath("ffplay"); ok {
		cmd := exec.CommandContext(ctx, "ffplay",
			"-nodisp", "-autoexit", "-loglevel", "error", url)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		return cmd.Run()
	}

	if _, ok := lookPath("ffmpeg"); ok {
		args := append([]string{"-loglevel", "error", "-i", url}, ffmpegAudioOutputArgs()...)
		cmd := exec.CommandContext(ctx, "ffmpeg", args...)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		return cmd.Run()
	}

	return errors.New("afplay/ffplay/ffmpeg not found — install ffmpeg (brew install ffmpeg) to use --play")
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
