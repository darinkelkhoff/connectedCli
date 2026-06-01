package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// wordmark is the "conctl" block-letter logo (ANSI Shadow style), one entry per
// row. Seven rows tall: six solid body rows plus the shadow row, so the gradient
// reaches a full solid row of blue before the blue shadow.
var wordmark = []string{
	` ██████╗   ██████╗   ███╗   ██╗   ██████╗  ████████╗  ██╗     `,
	`██╔════╝  ██╔═══██╗  ████╗  ██║  ██╔════╝  ╚══██╔══╝  ██║     `,
	`██║       ██║   ██║  ██╔██╗ ██║  ██║          ██║     ██║     `,
	`██║       ██║   ██║  ██║╚██╗██║  ██║          ██║     ██║     `,
	`██║       ██║   ██║  ██║╚██╗██║  ██║          ██║     ██║     `,
	`╚██████╗  ╚██████╔╝  ██║ ╚████║  ╚██████╗     ██║     ███████╗`,
	` ╚═════╝   ╚═════╝   ╚═╝  ╚═══╝   ╚═════╝     ╚═╝     ╚══════╝`,
}

// ANSI color codes; a vertical gradient evokes the colorful Connected artwork.
const (
	ansiReset = "\033[0m"
	ansiBold  = "\033[1m"
	ansiDim   = "\033[2m"
)

// wordmarkGradient colors the seven wordmark rows top-to-bottom:
// green, yellow, orange, red, purple, blue (solid), blue (shadow).
var wordmarkGradient = []string{
	"\033[38;5;46m",  // green
	"\033[38;5;226m", // yellow
	"\033[38;5;208m", // orange
	"\033[38;5;196m", // red
	"\033[38;5;129m", // purple
	"\033[38;5;33m",  // blue (solid row)
	"\033[38;5;33m",  // blue (shadow)
}

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

	fmt.Fprintln(w)
	for i, line := range wordmark {
		fmt.Fprintln(w, colorize(line, wordmarkGradient[i%len(wordmarkGradient)]))
	}

	width := utf8.RuneCountInString(wordmark[0])
	center := func(s string) string {
		pad := (width - utf8.RuneCountInString(s)) / 2
		if pad < 0 {
			pad = 0
		}
		return strings.Repeat(" ", pad) + s
	}

	tagline := "The Connected CLI"
	ver := fmt.Sprintf("#%s - %s", Version, Codename)
	fmt.Fprintf(w, "\n%s\n%s\n\n", colorize(center(tagline), ansiBold), center(ver))
	fmt.Fprintln(w, rundown)
	fmt.Fprintf(w, "\n%s\n\n", colorize("Run `conctl --help` for all commands and flags.", ansiDim))
}
