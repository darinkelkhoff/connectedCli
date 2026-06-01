package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/dkelkhoff/connectedCli/internal/render"
	"github.com/dkelkhoff/connectedCli/internal/rickies"
	"github.com/spf13/cobra"
)

// formatTitles renders a host's titles for text output: semicolon-delimited,
// or "Man of the People" when the host holds no titles.
func formatTitles(titles []string) string {
	if len(titles) == 0 {
		return "Man of the People"
	}
	return strings.Join(titles, "; ")
}

// lookupHost finds a host's titles case-insensitively, returning the canonical
// host name (e.g. "Federico" for "federico").
func lookupHost(titles map[string][]string, query string) (string, []string) {
	for h, ts := range titles {
		if strings.EqualFold(h, query) {
			return h, ts
		}
	}
	return query, nil
}

// rickiesRundown lists the rickies subcommands, shown under the bare-command banner.
const rickiesRundown = `Rounds:
  Titles           conctl rickies titles [host]  current chairmen & titles
  Bill             conctl rickies bill           the full current rules
  Games            conctl rickies games          every Rickies game
  Game             conctl rickies game <name>    one game's picks & scores`

// printStandings prints the current chairmen and per-host titles.
func printStandings(out io.Writer, w rickies.Winners) {
	fmt.Fprintf(out, "Annual Chairman:  %s (%s, %s)\n", w.Annual.Winner, w.Annual.GameName, w.Annual.Date)
	fmt.Fprintf(out, "Keynote Chairman: %s (%s, %s)\n", w.Keynote.Winner, w.Keynote.GameName, w.Keynote.Date)
	fmt.Fprintln(out, "\nTitles:")
	hosts := make([]string, 0, len(w.Titles))
	for h := range w.Titles {
		hosts = append(hosts, h)
	}
	sort.Strings(hosts)
	for _, h := range hosts {
		fmt.Fprintf(out, "  %s: %s\n", h, formatTitles(w.Titles[h]))
	}
}

func init() {
	root := &cobra.Command{
		Use:   "rickies",
		Short: "The Rickies: standings, titles, games, and the bill",
		RunE: func(c *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			if jsonOutput {
				w, err := rickies.FetchWinners(ctx)
				if err != nil {
					return err
				}
				return render.JSON(c.OutOrStdout(), w)
			}
			out := c.OutOrStdout()
			// Small banner: the opening line of the current Bill of Rickies.
			if entries, berr := rickies.FetchBill(ctx, time.Now().Unix()); berr == nil {
				if para := rickies.FirstParagraph(entries); para != "" {
					fmt.Fprintf(out, "%s\n\n", para)
				}
			}
			fmt.Fprintln(out, rickiesRundown)
			return nil
		},
	}

	titles := &cobra.Command{
		Use:   "titles [host]",
		Short: "Current chairmen and titles (or one host's titles)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			w, err := rickies.FetchWinners(ctx)
			if err != nil {
				return err
			}
			// One host: just that host's titles (case-insensitive).
			if len(args) == 1 {
				host, ts := lookupHost(w.Titles, args[0])
				if jsonOutput {
					return render.JSON(c.OutOrStdout(), map[string][]string{host: ts})
				}
				fmt.Fprintf(c.OutOrStdout(), "%s: %s\n", host, formatTitles(ts))
				return nil
			}
			// No host: the full standings (chairmen + every host's titles).
			if jsonOutput {
				return render.JSON(c.OutOrStdout(), w)
			}
			printStandings(c.OutOrStdout(), w)
			return nil
		},
	}

	games := &cobra.Command{
		Use:   "games",
		Short: "List every Rickies game",
		RunE: func(c *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()
			index, err := rickies.FetchGameIndex(ctx)
			if err != nil {
				return err
			}
			if jsonOutput {
				sort.Slice(index, func(i, j int) bool { return index[i].Name < index[j].Name })
				return render.JSON(c.OutOrStdout(), index)
			}
			names := make([]string, 0, len(index))
			for _, g := range index {
				names = append(names, g.Name)
			}
			sort.Strings(names)
			out := c.OutOrStdout()
			fmt.Fprintf(out, "%d Rickies games (use `conctl rickies game <name>` for one):\n", len(names))
			for _, n := range names {
				fmt.Fprintf(out, "  %s\n", n)
			}
			return nil
		},
	}

	game := &cobra.Command{
		Use:   "game <name>",
		Short: "Show one Rickies game: picks, who made them, and how they scored",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			query := strings.Join(args, " ")
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()
			index, err := rickies.FetchGameIndex(ctx)
			if err != nil {
				return err
			}
			ref, err := matchGame(index, query)
			if err != nil {
				return err
			}
			g, err := rickies.FetchGame(ctx, ref.Path)
			if err != nil {
				return err
			}
			if jsonOutput {
				return render.JSON(c.OutOrStdout(), g)
			}
			out := c.OutOrStdout()
			color := colorEnabled(out)
			fmt.Fprintf(out, "%s — %s\n", g.Name, g.GameType)
			fmt.Fprintf(out, "Picked %s · Graded %s\n", g.DatePicked, g.DateGraded)
			renderSubGame(out, "Main Game", g.MainGame, color)
			renderSubGame(out, "The Flexies", g.Flexies, color)
			return nil
		},
	}

	bill := &cobra.Command{
		Use:   "bill",
		Short: "Print the full Bill of Rickies (the current rules)",
		RunE: func(c *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			entries, err := rickies.FetchBill(ctx, time.Now().Unix())
			if err != nil {
				return err
			}
			if len(entries) == 0 {
				return errors.New("could not read the Bill of Rickies")
			}
			if jsonOutput {
				return render.JSON(c.OutOrStdout(), entries)
			}
			printBill(c.OutOrStdout(), entries, colorEnabled(c.OutOrStdout()))
			return nil
		},
	}

	root.AddCommand(titles, games, game, bill)
	registerCommands = append(registerCommands, root)
}

// printBill renders the bill: headings (bold, spaced) and bulleted rules.
func printBill(out io.Writer, entries []rickies.BillEntry, color bool) {
	for i, e := range entries {
		if e.Heading {
			if i > 0 {
				fmt.Fprintln(out)
			}
			h := strings.ToUpper(e.Text)
			if color {
				h = ansiBold + h + ansiReset
			}
			fmt.Fprintf(out, "%s\n", h)
		} else {
			fmt.Fprintf(out, "  • %s\n", e.Text)
		}
	}
}

// matchGame resolves a query to a single game: exact (case-insensitive) match
// wins; otherwise a unique case-insensitive substring match; otherwise an error.
func matchGame(index []rickies.GameRef, query string) (rickies.GameRef, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	var partial []rickies.GameRef
	for _, g := range index {
		name := strings.ToLower(g.Name)
		if name == q {
			return g, nil
		}
		if strings.Contains(name, q) {
			partial = append(partial, g)
		}
	}
	switch len(partial) {
	case 0:
		return rickies.GameRef{}, fmt.Errorf("no Rickies game matching %q (see `conctl rickies games`)", query)
	case 1:
		return partial[0], nil
	default:
		names := make([]string, 0, len(partial))
		for _, g := range partial {
			names = append(names, g.Name)
		}
		sort.Strings(names)
		return rickies.GameRef{}, fmt.Errorf("%q matches %d games — be more specific: %s", query, len(partial), strings.Join(names, "; "))
	}
}

// mark returns a colored ✓/✗ for a hit/miss.
func mark(hit, color bool) string {
	if hit {
		if color {
			return "\033[92m✓" + ansiReset
		}
		return "✓"
	}
	if color {
		return "\033[91m✗" + ansiReset
	}
	return "✗"
}

// scoreLine formats a sub-game's per-host scores (main games show points; the
// flexies show correct/total).
func scoreLine(scores []rickies.HostScore) string {
	parts := make([]string, 0, len(scores))
	for _, s := range scores {
		if s.Total > 0 {
			parts = append(parts, fmt.Sprintf("%s %d/%d", s.Host, s.Correct, s.Total))
		} else {
			parts = append(parts, fmt.Sprintf("%s %d", s.Host, s.Score))
		}
	}
	return strings.Join(parts, " · ")
}

// renderSubGame prints one round (main game or flexies): winner, scores, and the
// picks grouped by host with a ✓/✗ for each.
func renderSubGame(out io.Writer, title string, sg rickies.SubGame, color bool) {
	if len(sg.Picks) == 0 && len(sg.Scores) == 0 {
		return
	}
	fmt.Fprintf(out, "\n%s", title)
	if sg.Winner != "" {
		fmt.Fprintf(out, " — Winner: %s", sg.Winner)
	}
	fmt.Fprintln(out)
	if len(sg.Scores) > 0 {
		fmt.Fprintf(out, "  %s\n", scoreLine(sg.Scores))
	}

	// Group picks by host, ordered by the scores list (ranked), then any extras.
	byHost := map[string][]rickies.Pick{}
	for _, p := range sg.Picks {
		byHost[p.Host] = append(byHost[p.Host], p)
	}
	var order []string
	seen := map[string]bool{}
	for _, s := range sg.Scores {
		if !seen[s.Host] {
			order = append(order, s.Host)
			seen[s.Host] = true
		}
	}
	for _, p := range sg.Picks {
		if !seen[p.Host] {
			order = append(order, p.Host)
			seen[p.Host] = true
		}
	}
	for _, h := range order {
		ps := byHost[h]
		if len(ps) == 0 {
			continue
		}
		fmt.Fprintf(out, "\n  %s\n", h)
		for _, p := range ps {
			fmt.Fprintf(out, "    %s %s\n", mark(p.Hit(), color), p.Text)
		}
	}

	if sg.CharityName != "" {
		line := "Charity: " + sg.CharityName
		if sg.Donation > 0 {
			line = fmt.Sprintf("Charity: $%d to %s", sg.Donation, sg.CharityName)
		}
		if sg.Donator != "" {
			line += fmt.Sprintf(" (donated by %s)", sg.Donator)
		}
		fmt.Fprintf(out, "\n  %s\n", line)
	}
}
