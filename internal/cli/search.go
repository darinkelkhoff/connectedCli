package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dkelkhoff/connectedCli/internal/podsearch"
	"github.com/dkelkhoff/connectedCli/internal/render"
	"github.com/spf13/cobra"
)

func init() {
	var limit int
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the Connected transcripts (via David Smith's Podcast Search)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			term := strings.Join(args, " ")
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()
			hits, err := podsearch.Search(ctx, term, limit)
			if err != nil {
				return err
			}
			if len(hits) == 0 {
				return errors.New("no results")
			}
			if jsonOutput {
				return render.JSON(c.OutOrStdout(), hits)
			}
			for _, h := range hits {
				ep := ""
				if h.EpisodeNumber > 0 {
					ep = fmt.Sprintf("#%d ", h.EpisodeNumber)
				}
				fmt.Fprintf(c.OutOrStdout(), "%s%s  …%s…\n  %s\n", ep, h.Time, h.Snippet, h.URL)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 10, "max results")
	registerCommands = append(registerCommands, cmd)
}
