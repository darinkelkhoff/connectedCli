package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/darinkelkhoff/connectedCli/internal/chapters"
	"github.com/darinkelkhoff/connectedCli/internal/feed"
	"github.com/darinkelkhoff/connectedCli/internal/render"
	"github.com/spf13/cobra"
)

// resolveEpisode returns the target episode: explicit number or latest.
func resolveEpisode(ctx context.Context, number int) (feed.Episode, error) {
	eps, err := feed.Fetch(ctx)
	if err != nil {
		return feed.Episode{}, err
	}
	if number > 0 {
		e, ok := feed.ByNumber(eps, number)
		if !ok {
			return feed.Episode{}, fmt.Errorf("episode %d not found (latest is %d)", number, feed.Latest(eps).Number)
		}
		return e, nil
	}
	return feed.Latest(eps), nil
}

func msToClock(ms uint32) string {
	s := ms / 1000
	return fmt.Sprintf("%02d:%02d:%02d", s/3600, (s%3600)/60, s%60)
}

// printChapterList writes the indented "NN. HH:MM:SS  Title" lines for chapters.
func printChapterList(out io.Writer, chs []chapters.Chapter) {
	for _, ch := range chs {
		fmt.Fprintf(out, "  %2d. %s  %s\n", ch.Index, msToClock(ch.StartMs), ch.Title)
	}
}

func init() {
	cmd := &cobra.Command{
		Use:   "chapters [episode]",
		Short: "List an episode's chapters",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			num, err := parseOptionalEp(args)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
			if jsonOutput {
				return render.JSON(c.OutOrStdout(), chs)
			}
			fmt.Fprintf(c.OutOrStdout(), "Connected #%d — %s\n", ep.Number, ep.Title)
			printChapterList(c.OutOrStdout(), chs)
			return nil
		},
	}
	registerCommands = append(registerCommands, cmd)
}
