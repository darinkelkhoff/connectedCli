package cli

import (
	"context"
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/darinkelkhoff/connectedCli/internal/render"
	"github.com/spf13/cobra"
)

// printNoteLinks writes "• text / url" lines, hyperlinking the text via OSC 8
// when color (a terminal) is enabled.
func printNoteLinks(out io.Writer, links []NoteLink, color bool) {
	for _, l := range links {
		text := l.Text
		if color {
			text = osc8(l.URL, l.Text)
		}
		fmt.Fprintf(out, "  • %s\n    %s\n", text, l.URL)
	}
}

// NoteLink is one link from an episode's show notes.
type NoteLink struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

var noteLinkRe = regexp.MustCompile(`(?is)<a[^>]+href=['"]([^'"]+)['"][^>]*>(.*?)</a>`)
var noteTagRe = regexp.MustCompile(`<[^>]+>`)
var noteWsRe = regexp.MustCompile(`\s+`)

// parseNoteLinks extracts {text, url} links from show-notes HTML, in order,
// skipping any with empty text.
func parseNoteLinks(notesHTML string) []NoteLink {
	var links []NoteLink
	for _, m := range noteLinkRe.FindAllStringSubmatch(notesHTML, -1) {
		text := strings.TrimSpace(noteWsRe.ReplaceAllString(html.UnescapeString(noteTagRe.ReplaceAllString(m[2], "")), " "))
		if text == "" {
			continue
		}
		links = append(links, NoteLink{Text: text, URL: m[1]})
	}
	return links
}

func init() {
	cmd := &cobra.Command{
		Use:   "notes [episode]",
		Short: "Show an episode's show notes and links (from the RSS feed)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			num, err := parseOptionalEp(args)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			ep, err := resolveEpisode(ctx, num)
			if err != nil {
				return err
			}
			links := parseNoteLinks(ep.NotesHTML)

			if jsonOutput {
				return render.JSON(c.OutOrStdout(), struct {
					Episode int        `json:"episode"`
					Title   string     `json:"title"`
					Summary string     `json:"summary,omitempty"`
					Links   []NoteLink `json:"links"`
				}{ep.Number, ep.Title, ep.Summary, links})
			}

			out := c.OutOrStdout()
			color := colorEnabled(out)
			fmt.Fprintf(out, "Connected #%d — %s\n", ep.Number, ep.Title)
			if ep.Summary != "" {
				fmt.Fprintf(out, "%s\n", ep.Summary)
			}
			if len(links) == 0 {
				return nil
			}
			fmt.Fprintln(out, "\nShow notes:")
			printNoteLinks(out, links, color)
			return nil
		},
	}
	registerCommands = append(registerCommands, cmd)
}