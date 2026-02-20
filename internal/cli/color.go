// Package cli implements all CLI subcommands for zvault.
package cli

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/zarlcorp/core/pkg/zstyle"
)

// lipgloss styles using catppuccin mocha palette via zstyle.
// lipgloss handles NO_COLOR automatically.
var (
	boldStyle       = lipgloss.NewStyle().Bold(true)
	greenStyle      = lipgloss.NewStyle().Foreground(zstyle.Green)
	redStyle        = lipgloss.NewStyle().Foreground(zstyle.Red)
	yellowStyle     = lipgloss.NewStyle().Foreground(zstyle.Yellow)
	peachStyle      = lipgloss.NewStyle().Foreground(zstyle.Peach)
	blueStyle       = lipgloss.NewStyle().Foreground(zstyle.Blue)
	mutedStyle      = lipgloss.NewStyle().Foreground(zstyle.Overlay1)
	boldMauveStyle  = lipgloss.NewStyle().Bold(true).Foreground(zstyle.Mauve)
	boldRedStyle    = lipgloss.NewStyle().Bold(true).Foreground(zstyle.Red)
	boldYellowStyle = lipgloss.NewStyle().Bold(true).Foreground(zstyle.Yellow)
)

func bold(s string) string       { return boldStyle.Render(s) }
func green(s string) string      { return greenStyle.Render(s) }
func red(s string) string        { return redStyle.Render(s) }
func yellow(s string) string     { return yellowStyle.Render(s) }
func peach(s string) string      { return peachStyle.Render(s) }
func blue(s string) string       { return blueStyle.Render(s) }
func muted(s string) string      { return mutedStyle.Render(s) }
func boldMauve(s string) string  { return boldMauveStyle.Render(s) }
func boldRed(s string) string    { return boldRedStyle.Render(s) }
func boldYellow(s string) string { return boldYellowStyle.Render(s) }

// errf prints a formatted error message to stderr.
func errf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, red("error: ")+format+"\n", args...)
}
