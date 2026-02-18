package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/zarlcorp/core/pkg/zstyle"
)

// placeholderModel is a stub view for screens built in future specs (048/049).
type placeholderModel struct {
	id     viewID
	width  int
	height int
}

func newPlaceholder(id viewID) placeholderModel {
	return placeholderModel{id: id}
}

func (m placeholderModel) Update(msg tea.Msg) (placeholderModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, zstyle.KeyBack) {
			parent := parentView(m.id)
			return m, func() tea.Msg { return navigateMsg{view: parent} }
		}
	}
	return m, nil
}

func (m placeholderModel) View() string {
	title := viewTitle(m.id)
	msg := zstyle.MutedText.Render(fmt.Sprintf("%s (coming soon)", title))
	return fmt.Sprintf("\n  %s\n", msg)
}
