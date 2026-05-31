package cli

import (
	"context"
	"errors"
	"time"

	"github.com/dkelkhoff/connectedCli/internal/chapters"
	"github.com/spf13/cobra"
)

func init() {
	var chapterIdx int
	var first, last bool
	cmd := &cobra.Command{
		Use:   "play [episode]",
		Short: "Play a chapter of an episode (requires ffmpeg)",
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
	cmd.Flags().IntVar(&chapterIdx, "chapter", 0, "chapter index (0-based)")
	cmd.Flags().BoolVar(&first, "first", false, "first chapter")
	cmd.Flags().BoolVar(&last, "last", false, "last chapter")
	registerCommands = append(registerCommands, cmd)
}
