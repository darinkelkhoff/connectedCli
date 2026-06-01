package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/darinkelkhoff/connectedCli/internal/audio"
	"github.com/darinkelkhoff/connectedCli/internal/chapters"
	"github.com/spf13/cobra"
)

func init() {
	var chapterIdx int
	var first, last bool
	cmd := &cobra.Command{
		Use:   "play [episode]",
		Short: "Play an episode (or a single chapter with --chapter/--first/--last)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			num, err := parseOptionalEp(args)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Hour)
			defer cancel()
			ep, err := resolveEpisode(ctx, num)
			if err != nil {
				return err
			}

			// No chapter selector → play the whole episode.
			if !c.Flags().Changed("chapter") && !first && !last {
				fmt.Fprintf(c.OutOrStdout(), "▶ Connected #%d — %s\n", ep.Number, ep.Title)
				return audio.Play(ctx, ep.MP3URL)
			}

			chs, err := chapters.Fetch(ctx, ep.MP3URL)
			if err != nil {
				return err
			}
			if len(chs) == 0 {
				return errors.New("no chapters found")
			}
			ch := chapters.First(chs)
			switch {
			case last:
				ch = chapters.Last(chs)
			case first:
				ch = chapters.First(chs)
			default:
				selected, ok := chapters.At(chs, chapterIdx)
				if !ok {
					return errors.New("chapter index out of range")
				}
				ch = selected
			}
			return emitChapter(ctx, c, ep, ch, renderMode{play: true})
		},
	}
	cmd.Flags().IntVar(&chapterIdx, "chapter", 0, "play this chapter index (0-based) instead of the whole episode")
	cmd.Flags().BoolVar(&first, "first", false, "play the first chapter")
	cmd.Flags().BoolVar(&last, "last", false, "play the last chapter")
	registerCommands = append(registerCommands, cmd)
}
