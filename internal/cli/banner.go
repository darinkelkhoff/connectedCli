package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// wordmark is the "conctl" block-letter logo (ANSI Shadow style), one entry per row.
var wordmark = []string{
	` ██████╗   ██████╗   ███╗   ██╗   ██████╗  ████████╗  ██╗     `,
	`██╔════╝  ██╔═══██╗  ████╗  ██║  ██╔════╝  ╚══██╔══╝  ██║     `,
	`██║       ██║   ██║  ██╔██╗ ██║  ██║          ██║     ██║     `,
	`██║       ██║   ██║  ██║╚██╗██║  ██║          ██║     ██║     `,
	`╚██████╗  ╚██████╔╝  ██║ ╚████║  ╚██████╗     ██║     ███████╗`,
	` ╚═════╝   ╚═════╝   ╚═╝  ╚═══╝   ╚═════╝     ╚═╝     ╚══════╝`,
}

// ANSI color codes; a vertical gradient evokes the colorful Connected artwork.
const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiDim     = "\033[2m"
	ansiRed     = "\033[91m"
	ansiGreen   = "\033[92m"
	ansiYellow  = "\033[93m"
	ansiBlue    = "\033[94m"
	ansiMagenta = "\033[95m"
	ansiCyan    = "\033[96m"
)

var wordmarkGradient = []string{ansiCyan, ansiBlue, ansiMagenta, ansiRed, ansiYellow, ansiGreen}

// colorEnabled reports whether w is a real terminal and color isn't suppressed.
func colorEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// printBanner writes the no-argument splash: three-host dots, the conctl
// wordmark, the tagline, and the show-style rundown. It deliberately does NOT
// print cobra's full usage/flags help — that's reserved for -h/--help.
func printBanner(w io.Writer) {
	color := colorEnabled(w)
	colorize := func(s, code string) string {
		if !color {
			return s
		}
		return code + s + ansiReset
	}

	// Three overlapping-circle motif (the three hosts), red / green / blue.
	dots := colorize("●", ansiRed) + " " + colorize("●", ansiGreen) + " " + colorize("●", ansiBlue)
	fmt.Fprintf(w, "\n            %s\n\n", dots)

	for i, line := range wordmark {
		fmt.Fprintln(w, colorize(line, wordmarkGradient[i%len(wordmarkGradient)]))
	}

	tagline := "the Connected CLI"
	fmt.Fprintf(w, "\n%s%s\n\n", strings.Repeat(" ", 21), colorize(tagline, ansiBold))
	fmt.Fprintln(w, rundown)
	fmt.Fprintln(w)
}
