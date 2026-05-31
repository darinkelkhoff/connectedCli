package cli

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/dkelkhoff/connectedCli/internal/render"
	"github.com/dkelkhoff/connectedCli/internal/rickies"
	"github.com/spf13/cobra"
)

func init() {
	root := &cobra.Command{
		Use:   "rickies",
		Short: "Current Rickies chairmen, titles, and standings",
		RunE: func(c *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			w, err := rickies.FetchWinners(ctx)
			if err != nil {
				return err
			}
			if jsonOutput {
				return render.JSON(c.OutOrStdout(), w)
			}
			out := c.OutOrStdout()
			fmt.Fprintf(out, "Annual Chairman:  %s (%s, %s)\n", w.Annual.Winner, w.Annual.GameName, w.Annual.Date)
			fmt.Fprintf(out, "Keynote Chairman: %s (%s, %s)\n", w.Keynote.Winner, w.Keynote.GameName, w.Keynote.Date)
			fmt.Fprintln(out, "\nTitles:")
			hosts := make([]string, 0, len(w.Titles))
			for h := range w.Titles {
				hosts = append(hosts, h)
			}
			sort.Strings(hosts)
			for _, h := range hosts {
				fmt.Fprintf(out, "  %s: %v\n", h, w.Titles[h])
			}
			return nil
		},
	}

	titles := &cobra.Command{
		Use:   "titles [host]",
		Short: "List titles per host",
		RunE: func(c *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			w, err := rickies.FetchWinners(ctx)
			if err != nil {
				return err
			}
			data := w.Titles
			if len(args) == 1 {
				data = map[string][]string{args[0]: w.Titles[args[0]]}
			}
			if jsonOutput {
				return render.JSON(c.OutOrStdout(), data)
			}
			for h, ts := range data {
				fmt.Fprintf(c.OutOrStdout(), "%s: %v\n", h, ts)
			}
			return nil
		},
	}

	games := &cobra.Command{
		Use:   "games",
		Short: "List every Rickies game",
		RunE: func(c *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()
			m, err := rickies.FetchGames(ctx)
			if err != nil {
				return err
			}
			names := make([]string, 0, len(m))
			for k := range m {
				names = append(names, k)
			}
			sort.Strings(names)
			if jsonOutput {
				return render.JSON(c.OutOrStdout(), names)
			}
			for _, n := range names {
				fmt.Fprintln(c.OutOrStdout(), n)
			}
			return nil
		},
	}

	root.AddCommand(titles, games)
	registerCommands = append(registerCommands, root)
}
