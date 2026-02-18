// Package cli implements all CLI subcommands for zvault.
package cli

import (
	"fmt"
	"os"
)

// ANSI escape codes for catppuccin mocha palette.
const (
	ansiReset = "\033[0m"

	// text colors
	ansiText     = "\033[38;2;205;214;244m" // #cdd6f4
	ansiMuted    = "\033[38;2;127;132;156m" // #7f849c (overlay1)
	ansiSubtext  = "\033[38;2;186;194;222m" // #bac2de (subtext1)
	ansiOverlay0 = "\033[38;2;108;112;134m" // #6c7086

	// accent colors
	ansiMauve    = "\033[38;2;203;166;247m" // #cba6f7 (zvault accent)
	ansiGreen    = "\033[38;2;166;227;161m" // #a6e3a1
	ansiRed      = "\033[38;2;243;139;168m" // #f38ba8
	ansiYellow   = "\033[38;2;249;226;175m" // #f9e2af
	ansiPeach    = "\033[38;2;250;179;135m" // #fab387
	ansiBlue     = "\033[38;2;137;180;250m" // #89b4fa
	ansiLavender = "\033[38;2;180;190;254m" // #b4befe

	ansiBold = "\033[1m"
)

// noColor returns true when colored output should be suppressed.
func noColor() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return ok
}

// color helpers — return empty strings when NO_COLOR is set.

func colorize(code, s string) string {
	if noColor() {
		return s
	}
	return code + s + ansiReset
}

func bold(s string) string       { return colorize(ansiBold, s) }
func green(s string) string      { return colorize(ansiGreen, s) }
func red(s string) string        { return colorize(ansiRed, s) }
func yellow(s string) string     { return colorize(ansiYellow, s) }
func peach(s string) string      { return colorize(ansiPeach, s) }
func blue(s string) string       { return colorize(ansiBlue, s) }
func muted(s string) string      { return colorize(ansiMuted, s) }
func boldMauve(s string) string  { return colorize(ansiBold+ansiMauve, s) }
func boldRed(s string) string    { return colorize(ansiBold+ansiRed, s) }
func boldYellow(s string) string { return colorize(ansiBold+ansiYellow, s) }

// errf prints a formatted error message to stderr.
func errf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, red("error: ")+format+"\n", args...)
}
