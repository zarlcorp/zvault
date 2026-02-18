package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/zarlcorp/core/pkg/zstyle"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(zstyle.ZvaultAccent).
			Bold(true)

	headerTitleStyle = lipgloss.NewStyle().
				Foreground(zstyle.Subtext1)
)

// renderHeader returns the app name and current view title.
func renderHeader(id viewID, width int) string {
	app := headerStyle.Render("zvault")
	title := viewTitle(id)
	if title == "" {
		return fmt.Sprintf("  %s", app)
	}
	sep := zstyle.MutedText.Render(" / ")
	vt := headerTitleStyle.Render(title)
	_ = width
	return fmt.Sprintf("  %s%s%s", app, sep, vt)
}
