package tui

import "github.com/zarlcorp/core/pkg/zstyle"

// renderHeader returns the app name and current view title.
func renderHeader(id viewID, width int) string {
	_ = width
	return zstyle.RenderHeader("zvault", viewTitle(id), zstyle.ZvaultAccent)
}
