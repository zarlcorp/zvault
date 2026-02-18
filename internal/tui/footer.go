package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zarlcorp/core/pkg/zstyle"
)

var (
	footerKeyStyle  = lipgloss.NewStyle().Foreground(zstyle.Lavender).Bold(true)
	footerDescStyle = lipgloss.NewStyle().Foreground(zstyle.Overlay1)
	footerSepStyle  = lipgloss.NewStyle().Foreground(zstyle.Surface2)
)

// helpEntry is a key/description pair for the footer.
type helpEntry struct {
	key  string
	desc string
}

// renderFooter returns context-sensitive keybinding help for the current view.
func renderFooter(id viewID, width int) string {
	entries := helpFor(id)
	if len(entries) == 0 {
		return ""
	}

	var parts []string
	sep := footerSepStyle.Render(" | ")
	for _, e := range entries {
		k := footerKeyStyle.Render(e.key)
		d := footerDescStyle.Render(e.desc)
		parts = append(parts, k+" "+d)
	}

	line := strings.Join(parts, sep)
	_ = width
	return "  " + line
}

// helpFor returns the keybinding entries for a given view.
func helpFor(id viewID) []helpEntry {
	switch id {
	case viewPassword:
		return []helpEntry{
			{"enter", "submit"},
			{"tab", "next field"},
			{"ctrl+c", "quit"},
		}
	case viewMenu:
		return []helpEntry{
			{"↑/k", "up"},
			{"↓/j", "down"},
			{"enter", "select"},
			{"q", "quit"},
		}
	case viewSecretList:
		return []helpEntry{
			{"↑/k", "up"},
			{"↓/j", "down"},
			{"enter", "open"},
			{"esc", "back"},
			{"q", "quit"},
		}
	case viewTaskList:
		return []helpEntry{
			{"↑/k", "up"},
			{"↓/j", "down"},
			{"enter", "detail"},
			{"n", "new"},
			{"space", "toggle done"},
			{"d", "delete"},
			{"x", "clear done"},
			{"tab", "filter"},
			{"esc", "back"},
		}
	case viewSecretDetail:
		return []helpEntry{
			{"esc", "back"},
			{"q", "quit"},
		}
	case viewTaskDetail:
		return []helpEntry{
			{"e", "edit"},
			{"space", "toggle done"},
			{"d", "delete"},
			{"esc", "back"},
		}
	case viewSecretForm:
		return []helpEntry{
			{"tab", "next field"},
			{"enter", "save"},
			{"esc", "cancel"},
		}
	case viewTaskForm:
		return []helpEntry{
			{"tab", "next field"},
			{"ctrl+s", "save"},
			{"esc", "cancel"},
		}
	default:
		return []helpEntry{
			{"q", "quit"},
		}
	}
}
