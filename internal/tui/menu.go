package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zarlcorp/core/pkg/zstyle"
	"github.com/zarlcorp/zvault/internal/vault"
)

type menuItem int

const (
	menuSecrets menuItem = iota
	menuTasks
	menuItemCount // sentinel
)

// menuModel is the main menu with Secrets and Tasks options.
type menuModel struct {
	cursor       menuItem
	secretCount  int
	pendingCount int
	width        int
	height       int
}

func newMenuModel() menuModel {
	return menuModel{}
}

// refreshCounts loads counts from the vault.
func (m menuModel) refreshCounts(v *vault.Vault) menuModel {
	if v == nil {
		return m
	}
	secrets, err := v.Secrets().List()
	if err == nil {
		m.secretCount = len(secrets)
	}
	tasks, err := v.Tasks().List(pendingFilter())
	if err == nil {
		m.pendingCount = len(tasks)
	}
	return m
}

func (m menuModel) Update(msg tea.Msg) (menuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, zstyle.KeyUp):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, zstyle.KeyDown):
			if m.cursor < menuItemCount-1 {
				m.cursor++
			}
		case key.Matches(msg, zstyle.KeyEnter):
			return m, m.selectItem()
		}
	}
	return m, nil
}

func (m menuModel) selectItem() tea.Cmd {
	switch m.cursor {
	case menuSecrets:
		return func() tea.Msg { return navigateMsg{view: viewSecretList} }
	case menuTasks:
		return func() tea.Msg { return navigateMsg{view: viewTaskList} }
	}
	return nil
}

func (m menuModel) View() string {
	var b strings.Builder

	items := []struct {
		label string
		count string
	}{
		{"Secrets", fmt.Sprintf("(%d)", m.secretCount)},
		{"Tasks", fmt.Sprintf("(%d pending)", m.pendingCount)},
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(zstyle.ZvaultAccent).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(zstyle.Text)

	countStyle := lipgloss.NewStyle().
		Foreground(zstyle.Overlay1)

	cursorActive := lipgloss.NewStyle().
		Foreground(zstyle.ZvaultAccent).
		Render("▸ ")

	cursorInactive := "  "

	b.WriteString("\n")
	for i, item := range items {
		cursor := cursorInactive
		style := normalStyle
		if menuItem(i) == m.cursor {
			cursor = cursorActive
			style = selectedStyle
		}
		label := style.Render(item.label)
		count := countStyle.Render(" " + item.count)
		b.WriteString(fmt.Sprintf("  %s%s%s\n", cursor, label, count))
	}

	return b.String()
}
