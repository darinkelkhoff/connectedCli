package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/dkelkhoff/connectedCli/internal/chapters"
	"github.com/dkelkhoff/connectedCli/internal/podsearch"
	"github.com/spf13/cobra"
)

// intro/exit flag state, bound on the root command.
type introExitFlags struct {
	intro, exit       bool
	play, say, useLLM bool
	short             bool
	prompt            string
	episode           int
}

var ie introExitFlags

func bindIntroExit(root *cobra.Command) {
	f := root.PersistentFlags()
	f.BoolVar(&ie.intro, "intro", false, "render the first chapter (the cold open)")
	f.BoolVar(&ie.exit, "exit", false, "render the final chapter (the sign-off)")
	f.BoolVar(&ie.play, "play", false, "play the chapter audio (requires ffmpeg)")
	f.BoolVar(&ie.say, "say", false, "read the chapter aloud via macOS say")
	f.BoolVar(&ie.useLLM, "llm", false, "generate from the chapter via a local AI CLI")
	f.BoolVar(&ie.short, "short", false, "just the quick sign-offs")
	f.StringVar(&ie.prompt, "prompt", "", "LLM prompt preset (e.g. conclusion, cold-open, haiku)")
	f.IntVar(&ie.episode, "episode", 0, "target episode number (default: latest)")

	// Allow a positional episode number (e.g. `conctl --intro 601`) only when
	// --intro/--exit is set; otherwise keep cobra's "unknown command" behavior.
	root.Args = func(cmd *cobra.Command, args []string) error {
		if ie.intro || ie.exit {
			return nil
		}
		if len(args) > 0 {
			return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
		}
		return nil
	}

	// When --intro/--exit are present and no subcommand is given, run here.
	// With no flags at all, show the banner (NOT the full help).
	root.RunE = func(c *cobra.Command, args []string) error {
		if !ie.intro && !ie.exit {
			printBanner(c.OutOrStdout())
			return nil
		}
		return runIntroExit(c, args)
	}
}

// introExitWord names the targeted chapter for messages.
func introExitWord() string {
	if ie.exit {
		return "closing"
	}
	return "intro"
}

// introExitChapter selects the first chapter for --intro, the last for --exit.
func introExitChapter(chs []chapters.Chapter) chapters.Chapter {
	if ie.exit {
		return chapters.Last(chs)
	}
	return chapters.First(chs)
}

func runShortExit(w io.Writer) error {
	_, err := fmt.Fprintln(w, "Arrivederci. Cheerio. Bye, y'all.")
	return err
}

func runShortIntro(w io.Writer) error {
	_, err := fmt.Fprintln(w, "This is Connected.")
	return err
}

func runIntroExit(c *cobra.Command, args []string) error {
	if ie.short {
		if ie.exit {
			return runShortExit(c.OutOrStdout())
		}
		return runShortIntro(c.OutOrStdout())
	}

	// Episode may come from --episode or a positional number; both are "explicit".
	epNum, explicit := ie.episode, ie.episode > 0
	if !explicit && len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil {
			epNum, explicit = n, true
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()

	prompt := ie.prompt
	if ie.useLLM && prompt == "" {
		if ie.exit {
			prompt = "conclusion"
		} else {
			prompt = "cold-open"
		}
	}
	mode := renderMode{play: ie.play, say: ie.say, useLLM: ie.useLLM, prompt: prompt}

	err := emitIntroExit(ctx, c, epNum, mode)

	// Fallback: if the (defaulted, not explicitly chosen) latest episode has no
	// transcript yet for a text mode, show the same chapter from the newest
	// transcribed episode instead — with a note. --play needs no transcript.
	var nt *podsearch.NoTranscriptError
	if err != nil && !ie.play && !explicit && errors.As(err, &nt) {
		fmt.Fprintf(c.ErrOrStderr(),
			"note: episode %d isn't transcribed yet — showing the %s from %d instead.\n\n",
			nt.Episode, introExitWord(), nt.Newest)
		return emitIntroExit(ctx, c, nt.Newest, mode)
	}
	return err
}

// emitIntroExit resolves an episode (0 = latest), fetches its chapters, and
// emits the intro (first) or closing (last) chapter in the given mode.
func emitIntroExit(ctx context.Context, c *cobra.Command, epNum int, mode renderMode) error {
	ep, err := resolveEpisode(ctx, epNum)
	if err != nil {
		return err
	}
	chs, err := chapters.Fetch(ctx, ep.MP3URL)
	if err != nil {
		return err
	}
	if len(chs) == 0 {
		return errors.New("no chapters found")
	}
	return emitChapter(ctx, c, ep, introExitChapter(chs), mode)
}
