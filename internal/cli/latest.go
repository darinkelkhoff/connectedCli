package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/dkelkhoff/connectedCli/internal/chapters"
	"github.com/dkelkhoff/connectedCli/internal/feed"
	"github.com/dkelkhoff/connectedCli/internal/render"
	"github.com/spf13/cobra"
)

func init() {
	var full bool
	cmd := &cobra.Command{
		Use:   "latest",
		Short: "Show the most recent Connected episode",
		RunE: func(c *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			eps, err := feed.Fetch(ctx)
			if err != nil {
				return err
			}
			e := feed.Latest(eps)

			if !full {
				if jsonOutput {
					return render.JSON(c.OutOrStdout(), e)
				}
				fmt.Fprintf(c.OutOrStdout(), "Connected #%d — %s\n%s\n", e.Number, e.Title, e.Link)
				return nil
			}

			// --full: also pull chapters (from the MP3) and show notes (from the feed).
			chs, _ := chapters.Fetch(ctx, e.MP3URL) // best-effort; absent on error
			links := parseNoteLinks(e.NotesHTML)

			if jsonOutput {
				return render.JSON(c.OutOrStdout(), struct {
					feed.Episode
					Chapters []chapters.Chapter `json:"chapters"`
					Links    []NoteLink         `json:"links"`
				}{e, chs, links})
			}

			out := c.OutOrStdout()
			fmt.Fprintf(out, "Connected #%d — %s\n", e.Number, e.Title)
			if !e.Date.IsZero() {
				fmt.Fprintf(out, "%s\n", e.Date.Format("Mon, 2 Jan 2006"))
			}
			fmt.Fprintln(out, e.Link)
			if e.Summary != "" {
				fmt.Fprintf(out, "\n%s\n", e.Summary)
			}
			if len(chs) > 0 {
				fmt.Fprintln(out, "\nChapters:")
				printChapterList(out, chs)
			}
			if len(links) > 0 {
				fmt.Fprintln(out, "\nShow notes:")
				printNoteLinks(out, links, colorEnabled(out))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&full, "full", false, "also print chapters and show notes")
	registerCommands = append(registerCommands, cmd)
}
