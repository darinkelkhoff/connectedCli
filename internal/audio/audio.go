package audio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
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
// Tries ffplay, then ffmpeg, then downloads to a temp file and plays with afplay.
func Play(ctx context.Context, url string) error {
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

	if _, ok := lookPath("afplay"); ok {
		return downloadAndPlay(ctx, url)
	}

	return errors.New("ffplay, ffmpeg, or afplay not found — install ffmpeg (brew install ffmpeg) to use --play")
}

// downloadAndPlay downloads url to a temp file and plays it with afplay.
// afplay cannot stream from podcast CDN URLs directly, so we fetch first.
func downloadAndPlay(ctx context.Context, url string) error {
	// Catch SIGINT so the context is cancelled instead of the process being killed
	// abruptly — this lets defers run and the temp file get cleaned up.
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Fprintln(os.Stderr, "ffplay/ffmpeg not found — downloading episode to tmp file (which will be auto-removed) for afplay...")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tmp, err := os.CreateTemp("", "conctl-*.mp3")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return fmt.Errorf("download: %w", err)
	}
	tmp.Close()

	cmd := exec.CommandContext(ctx, "afplay", tmp.Name())
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("say: %w", err)
	}
	return nil
}
