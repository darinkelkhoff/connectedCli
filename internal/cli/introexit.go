package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/dkelkhoff/connectedCli/internal/chapters"
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

	// When --intro/--exit are present and no subcommand is given, run here.
	root.RunE = func(c *cobra.Command, args []string) error {
		if !ie.intro && !ie.exit {
			return c.Help()
		}
		return runIntroExit(c)
	}
}

func runShortExit(w io.Writer) error {
	_, err := fmt.Fprintln(w, "Arrivederci. Cheerio. Bye, y'all.")
	return err
}

func runShortIntro(w io.Writer) error {
	_, err := fmt.Fprintln(w, "This is Connected.")
	return err
}

func runIntroExit(c *cobra.Command) error {
	if ie.short {
		if ie.exit {
			return runShortExit(c.OutOrStdout())
		}
		return runShortIntro(c.OutOrStdout())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()
	ep, err := resolveEpisode(ctx, ie.episode)
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
	ch := chapters.First(chs)
	if ie.exit {
		ch = chapters.Last(chs)
	}
	prompt := ie.prompt
	if ie.useLLM && prompt == "" {
		if ie.exit {
			prompt = "conclusion"
		} else {
			prompt = "cold-open"
		}
	}
	return emitChapter(ctx, c, ep, ch, renderMode{
		play: ie.play, say: ie.say, useLLM: ie.useLLM, prompt: prompt,
	})
}
