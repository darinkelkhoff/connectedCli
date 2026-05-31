package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/dkelkhoff/connectedCli/internal/feed"
	"github.com/dkelkhoff/connectedCli/internal/render"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "latest",
		Short: "Show the most recent Connected episode",
		RunE: func(c *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			eps, err := feed.Fetch(ctx)
			if err != nil {
				return err
			}
			e := feed.Latest(eps)
			if jsonOutput {
				return render.JSON(c.OutOrStdout(), e)
			}
			fmt.Fprintf(c.OutOrStdout(), "Connected #%d — %s\n%s\n", e.Number, e.Title, e.Link)
			return nil
		},
	}
	registerCommands = append(registerCommands, cmd)
}
