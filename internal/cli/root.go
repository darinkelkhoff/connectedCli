package cli

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed versions.txt
var versionsFile string

// Versions are episode-styled: a zero-padded number plus a title in the show's
// spirit. Version (the number) may be overridden at build time via -ldflags.
var (
	Version  = "001"
	Codename = codenameFor(Version)
)

// codenameFor looks up the title for a version number in versions.txt.
// Falls back to "unknown" if the version isn't listed.
func codenameFor(v string) string {
	for _, line := range strings.Split(strings.TrimSpace(versionsFile), "\n") {
		if num, title, ok := strings.Cut(line, " "); ok && num == v {
			return title
		}
	}
	return "unknown"
}

// jsonOutput is the global --json flag, read by subcommands.
var jsonOutput bool

// registerCommands holds subcommands registered from their own files' init().
var registerCommands []*cobra.Command

// rundown is the show-style command listing, shared by the no-arg banner and --help.
const rundown = `Chapters:
  Intro            conctl --intro            the cold open
  Follow-up        conctl search <query>     search the transcripts
  Topics           conctl chapters [ep]      list an episode's chapters
                   conctl play [ep]          play an episode (or a chapter)
  The Rickies      conctl rickies            current chairmen & standings
  Show Notes       conctl notes [ep]         links & summary for an episode
  Closing          conctl --exit             the sign-off`

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "conctl",
		Short:         "A command-line companion for the Connected podcast",
		Version:       fmt.Sprintf("#%s - %s", Version, Codename),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("conctl {{.Version}}\n")
	// --help shows just usage and below (the Chapters rundown lives in the
	// no-argument banner, not in help).
	root.SetHelpTemplate("{{.UsageString}}")
	root.PersistentFlags().BoolVar(&jsonOutput, "json", false, "emit structured JSON for agents")

	for _, c := range registerCommands {
		root.AddCommand(c)
	}
	bindIntroExit(root)
	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "conctl:", err)
		os.Exit(1)
	}
}
