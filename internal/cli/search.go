package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
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
			printSearchHits(c.OutOrStdout(), term, hits)
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 10, "max results")
	registerCommands = append(registerCommands, cmd)
}

// printSearchHits renders hits grouped by episode, with the search term
// highlighted and each timestamp linked (OSC 8) to its deep-link when the
// output is a terminal.
func printSearchHits(w io.Writer, term string, hits []podsearch.Hit) {
	color := colorEnabled(w)
	highlight := makeHighlighter(term, color)

	lastEp := -1
	for _, h := range hits {
		if h.EpisodeInternalID != lastEp {
			lastEp = h.EpisodeInternalID
			header := h.EpisodeTitle
			if h.EpisodeNumber > 0 {
				header = fmt.Sprintf("#%d %s", h.EpisodeNumber, h.EpisodeTitle)
			}
			fmt.Fprintf(w, "\n%s:\n", strings.TrimSpace(header))
		}
		ts := h.Time
		if color {
			ts = osc8(h.URL, h.Time)
		}
		fmt.Fprintf(w, "  %s  … %s …\n", ts, highlight(h.Snippet))
	}
}

// makeHighlighter returns a function that wraps case-insensitive occurrences of
// term in a highlight color. It is a no-op when color is disabled.
func makeHighlighter(term string, color bool) func(string) string {
	term = strings.TrimSpace(term)
	if !color || term == "" {
		return func(s string) string { return s }
	}
	re, err := regexp.Compile("(?i)" + regexp.QuoteMeta(term))
	if err != nil {
		return func(s string) string { return s }
	}
	return func(s string) string {
		return re.ReplaceAllStringFunc(s, func(m string) string {
			return "\033[1;93m" + m + ansiReset
		})
	}
}

// osc8 wraps text in an OSC 8 terminal hyperlink to url.
func osc8(url, text string) string {
	return "\033]8;;" + url + "\033\\" + text + "\033]8;;\033\\"
}
