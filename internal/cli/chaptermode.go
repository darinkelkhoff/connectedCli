package cli

import (
	"context"
	"fmt"

	"github.com/dkelkhoff/connectedCli/internal/audio"
	"github.com/dkelkhoff/connectedCli/internal/chapters"
	"github.com/dkelkhoff/connectedCli/internal/feed"
	"github.com/dkelkhoff/connectedCli/internal/llm"
	"github.com/dkelkhoff/connectedCli/internal/podsearch"
	"github.com/dkelkhoff/connectedCli/internal/render"
	"github.com/spf13/cobra"
)

// renderMode captures which output a chapter command should produce.
type renderMode struct {
	play   bool
	say    bool
	useLLM bool
	prompt string
}

func chapterDurationSec(startMs, endMs uint32) int {
	if endMs <= startMs {
		return 0
	}
	return int((endMs - startMs) / 1000)
}

// chapterText fetches the transcript text for a chapter's time range.
func chapterText(ctx context.Context, ep feed.Episode, ch chapters.Chapter) (string, error) {
	id, err := podsearch.InternalID(ctx, ep.Number)
	if err != nil {
		return "", err
	}
	segs, err := podsearch.FetchTranscript(ctx, id)
	if err != nil {
		return "", err
	}
	return podsearch.TextInRange(segs, int(ch.StartMs/1000), int(ch.EndMs/1000)), nil
}

// emitChapter renders one chapter according to mode. Default (no flags) prints transcript text.
func emitChapter(ctx context.Context, c *cobra.Command, ep feed.Episode, ch chapters.Chapter, mode renderMode) error {
	out := c.OutOrStdout()

	if mode.play {
		dur := chapterDurationSec(ch.StartMs, ch.EndMs)
		fmt.Fprintf(out, "▶ Connected #%d — %s (%s)\n", ep.Number, ch.Title, msToClock(ch.StartMs))
		return audio.PlaySegment(ctx, ep.MP3URL, int(ch.StartMs/1000), dur)
	}

	// Text-producing modes need the chapter transcript text.
	text, err := chapterText(ctx, ep, ch)
	if err != nil {
		return err
	}

	if mode.useLLM {
		stop := startSpinner(fmt.Sprintf("Generating %s with your local AI…", mode.prompt))
		gen, provider, err := llm.Generate(ctx, mode.prompt, text, "")
		stop()
		if err != nil {
			return err
		}
		text = gen
		if !jsonOutput && !mode.say {
			fmt.Fprintf(out, "(%s · %s)\n", provider, mode.prompt)
		}
	}

	if mode.say {
		return audio.Say(ctx, text)
	}
	if jsonOutput {
		return render.JSON(out, map[string]any{
			"episode": ep.Number, "chapter": ch, "text": text,
		})
	}
	fmt.Fprintf(out, "Connected #%d — %s\n\n%s\n", ep.Number, ch.Title, text)
	return nil
}
