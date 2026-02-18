package tui

import (
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

const clipboardClearDelay = 10 * time.Second

// clipboardCopiedMsg signals that a value was copied.
type clipboardCopiedMsg struct {
	field string
}

// clipboardClearedMsg signals the auto-clear timer expired.
type clipboardClearedMsg struct{}

// copyToClipboard copies val to the system clipboard and schedules auto-clear.
func copyToClipboard(field, val string) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			if err := clipboard.WriteAll(val); err != nil {
				return errMsg{err: err}
			}
			return clipboardCopiedMsg{field: field}
		},
		scheduleClipboardClear(),
	)
}

// scheduleClipboardClear returns a tick command that fires after the clear delay.
func scheduleClipboardClear() tea.Cmd {
	return tea.Tick(clipboardClearDelay, func(time.Time) tea.Msg {
		// clear clipboard contents
		_ = clipboard.WriteAll("")
		return clipboardClearedMsg{}
	})
}
