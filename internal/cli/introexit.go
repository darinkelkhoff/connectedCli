package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/darinkelkhoff/connectedCli/internal/audio"
	"github.com/darinkelkhoff/connectedCli/internal/chapters"
	"github.com/darinkelkhoff/connectedCli/internal/podsearch"
	"github.com/spf13/cobra"
)

// introExitFlags holds the hidden root-flag form (the `connected --exit` joke).
type introExitFlags struct {
	intro, exit       bool
	play, say, useLLM bool
	short             bool
	prompt            string
	episode           int
}

var ie introExitFlags

// introExitOpts is a resolved intro/exit request, shared by the `intro`/`exit`
// subcommands and the hidden --intro/--exit root flags.
type introExitOpts struct {
	isExit   bool
	epNum    int  // 0 = latest
	explicit bool // the episode was chosen explicitly (positional or --episode)
	short    bool
	mode     renderMode
}

func bindIntroExit(root *cobra.Command) {
	// The canonical interface: real subcommands with their own help + flags.
	root.AddCommand(newIntroExitCmd(false), newIntroExitCmd(true))

	// Hidden root-flag aliases so `conctl --exit` / `conctl --intro` still work
	// (the joke the project is named for). Hidden so they don't clutter help.
	f := root.Flags()
	f.BoolVar(&ie.intro, "intro", false, "")
	f.BoolVar(&ie.exit, "exit", false, "")
	f.BoolVar(&ie.play, "play", false, "")
	f.BoolVar(&ie.say, "say", false, "")
	f.BoolVar(&ie.useLLM, "llm", false, "")
	f.BoolVar(&ie.short, "short", false, "")
	f.StringVar(&ie.prompt, "prompt", "", "")
	f.IntVar(&ie.episode, "episode", 0, "")
	for _, n := range []string{"intro", "exit", "play", "say", "llm", "short", "prompt", "episode"} {
		_ = f.MarkHidden(n)
	}

	// Allow a positional episode (e.g. `conctl --intro 601`) only with the
	// --intro/--exit flag form; otherwise keep cobra's "unknown command" error.
	root.Args = func(cmd *cobra.Command, args []string) error {
		if ie.intro || ie.exit {
			return nil
		}
		if len(args) > 0 {
			return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
		}
		return nil
	}

	root.RunE = func(c *cobra.Command, args []string) error {
		if !ie.intro && !ie.exit {
			printBanner(c.OutOrStdout())
			return nil
		}
		if ie.intro && ie.exit {
			return errors.New("use either --intro or --exit, not both")
		}
		epNum, explicit := ie.episode, ie.episode > 0
		if !explicit && len(args) > 0 {
			if n, err := strconv.Atoi(args[0]); err == nil {
				epNum, explicit = n, true
			}
		}
		return runIntroExit(c, introExitOpts{
			isExit:   ie.exit,
			epNum:    epNum,
			explicit: explicit,
			short:    ie.short,
			mode:     renderMode{play: ie.play, say: ie.say, useLLM: ie.useLLM, prompt: ie.prompt},
		})
	}
}

// newIntroExitCmd builds the `intro` or `exit` subcommand.
func newIntroExitCmd(isExit bool) *cobra.Command {
	var play, say, useLLM, short bool
	var prompt string

	use, shortDesc := "intro [episode]", "Play the show's intro (an episode's first chapter)"
	if isExit {
		use, shortDesc = "exit [episode]", "Play the show's closing (an episode's last chapter)"
	}
	cmd := &cobra.Command{
		Use:   use,
		Short: shortDesc,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			epNum, explicit := 0, false
			if len(args) > 0 {
				n, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Errorf("episode must be a number, got %q", args[0])
				}
				epNum, explicit = n, true
			}
			return runIntroExit(c, introExitOpts{
				isExit:   isExit,
				epNum:    epNum,
				explicit: explicit,
				short:    short,
				mode:     renderMode{play: play, say: say, useLLM: useLLM, prompt: prompt},
			})
		},
	}
	f := cmd.Flags()
	f.BoolVar(&play, "play", false, "play the chapter audio (requires ffmpeg)")
	f.BoolVar(&say, "say", false, "read the chapter aloud via macOS say")
	f.BoolVar(&useLLM, "llm", false, "generate from the chapter via a local AI CLI")
	f.BoolVar(&short, "short", false, "just the first sentence (intro) or sign-off (exit)")
	f.StringVar(&prompt, "prompt", "", "LLM prompt preset (conclusion, cold-open, recap, style-myke/-stephen/-federico, haiku)")
	return cmd
}

func firstSentence(text string) string {
	if i := strings.IndexAny(text, ".!?"); i >= 0 {
		return strings.TrimSpace(text[:i+1])
	}
	return strings.TrimSpace(text)
}

func shortIntroText(ctx context.Context, epNum int) (string, error) {
	ep, err := resolveEpisode(ctx, epNum)
	if err != nil {
		return "", err
	}
	chs, err := chapters.Fetch(ctx, ep.MP3URL)
	if err != nil {
		return "", err
	}
	if len(chs) == 0 {
		return "", errors.New("no chapters found")
	}
	text, err := chapterText(ctx, ep, chapters.First(chs))
	if err != nil {
		return "", err
	}
	return firstSentence(text), nil
}

// runIntroExit validates and executes a resolved intro/exit request.
func runIntroExit(c *cobra.Command, o introExitOpts) error {
	// --short is a content modifier and may combine with --say, but not --play/--llm.
	if o.short && (o.mode.play || o.mode.useLLM) {
		return errors.New("--short cannot be combined with --play or --llm")
	}
	modes := 0
	for _, b := range []bool{o.mode.play, o.mode.say, o.mode.useLLM} {
		if b {
			modes++
		}
	}
	if modes > 1 {
		return errors.New("choose at most one of --play, --say, --llm")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
	defer cancel()

	if o.short {
		var text string
		if o.isExit {
			text = "Arrivederci. Cheerio. Bye, y'all."
		} else {
			var err error
			text, err = shortIntroText(ctx, o.epNum)
			if err != nil {
				return err
			}
		}
		if o.mode.say {
			return audio.Say(ctx, text)
		}
		_, err := fmt.Fprintln(c.OutOrStdout(), text)
		return err
	}

	if o.mode.useLLM && o.mode.prompt == "" {
		o.mode.prompt = "cold-open"
		if o.isExit {
			o.mode.prompt = "conclusion"
		}
	}

	err := emitIntroExitChapter(ctx, c, o.isExit, o.epNum, o.mode)

	// Fallback: a defaulted (not explicit) latest episode with no transcript yet
	// for a text mode → use the newest transcribed episode, noting it on stderr.
	var nt *podsearch.NoTranscriptError
	if err != nil && !o.mode.play && !o.explicit && errors.As(err, &nt) {
		word := "intro"
		if o.isExit {
			word = "closing"
		}
		fmt.Fprintf(c.ErrOrStderr(),
			"note: episode %d isn't transcribed yet — showing the %s from %d instead.\n\n",
			nt.Episode, word, nt.Newest)
		return emitIntroExitChapter(ctx, c, o.isExit, nt.Newest, o.mode)
	}
	return err
}

// emitIntroExitChapter resolves an episode (0 = latest), fetches its chapters,
// and emits the first (intro) or last (closing) chapter in the given mode.
func emitIntroExitChapter(ctx context.Context, c *cobra.Command, isExit bool, epNum int, mode renderMode) error {
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
	ch := chapters.First(chs)
	if isExit {
		ch = chapters.Last(chs)
	}
	return emitChapter(ctx, c, ep, ch, mode)
}
