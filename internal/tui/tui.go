// Package tui implements the root Bubble Tea model for zvault.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zarlcorp/core/pkg/zstyle"
	"github.com/zarlcorp/zvault/internal/task"
	"github.com/zarlcorp/zvault/internal/vault"
)

// Model is the root TUI model.
type Model struct {
	vault   *vault.Vault
	view    viewID
	version string

	// sub-view models
	password       passwordModel
	menu           menuModel
	secretList     placeholderModel
	secretDetail   placeholderModel
	secretForm     placeholderModel
	taskList       placeholderModel
	taskDetail     placeholderModel
	taskForm       placeholderModel

	width  int
	height int
	err    string
}

// New creates the root TUI model.
func New(version string) Model {
	vaultDir := vault.DefaultDir()
	return Model{
		version:      version,
		view:         viewPassword,
		password:     newPasswordModel(vaultDir),
		menu:         newMenuModel(),
		secretList:   newPlaceholder(viewSecretList),
		secretDetail: newPlaceholder(viewSecretDetail),
		secretForm:   newPlaceholder(viewSecretForm),
		taskList:     newPlaceholder(viewTaskList),
		taskDetail:   newPlaceholder(viewTaskDetail),
		taskForm:     newPlaceholder(viewTaskForm),
	}
}

// NewWithDir creates a root model with a custom vault directory (for testing).
func NewWithDir(version, vaultDir string) Model {
	return Model{
		version:      version,
		view:         viewPassword,
		password:     newPasswordModel(vaultDir),
		menu:         newMenuModel(),
		secretList:   newPlaceholder(viewSecretList),
		secretDetail: newPlaceholder(viewSecretDetail),
		secretForm:   newPlaceholder(viewSecretForm),
		taskList:     newPlaceholder(viewTaskList),
		taskDetail:   newPlaceholder(viewTaskDetail),
		taskForm:     newPlaceholder(viewTaskForm),
	}
}

func (m Model) Init() tea.Cmd {
	return m.password.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.propagateSize()
		return m, nil

	case navigateMsg:
		m.view = msg.view
		m.err = ""
		// refresh menu counts when navigating to menu
		if msg.view == viewMenu && m.vault != nil {
			m.menu = m.menu.refreshCounts(m.vault)
		}
		return m, nil

	case vaultOpenedMsg:
		m.vault = msg.vault
		m.view = viewMenu
		m.menu = m.menu.refreshCounts(msg.vault)
		return m, nil

	case errMsg:
		if m.view == viewPassword {
			var cmd tea.Cmd
			m.password, cmd = m.password.Update(msg)
			return m, cmd
		}
		m.err = msg.err.Error()
		return m, nil

	case tea.KeyMsg:
		// global quit: q or ctrl+c, but not when typing in text inputs
		if m.view != viewPassword {
			if key.Matches(msg, zstyle.KeyQuit) {
				return m, tea.Quit
			}
		} else {
			// in password view, only ctrl+c quits
			if msg.Type == tea.KeyCtrlC {
				return m, tea.Quit
			}
		}
	}

	// dispatch to active sub-view
	return m.updateView(msg)
}

func (m Model) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.view {
	case viewPassword:
		m.password, cmd = m.password.Update(msg)
	case viewMenu:
		m.menu, cmd = m.menu.Update(msg)
	case viewSecretList:
		m.secretList, cmd = m.secretList.Update(msg)
	case viewSecretDetail:
		m.secretDetail, cmd = m.secretDetail.Update(msg)
	case viewSecretForm:
		m.secretForm, cmd = m.secretForm.Update(msg)
	case viewTaskList:
		m.taskList, cmd = m.taskList.Update(msg)
	case viewTaskDetail:
		m.taskDetail, cmd = m.taskDetail.Update(msg)
	case viewTaskForm:
		m.taskForm, cmd = m.taskForm.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	var b strings.Builder

	// header
	b.WriteString("\n")
	b.WriteString(renderHeader(m.view, m.width))
	b.WriteString("\n")

	// separator
	sep := lipgloss.NewStyle().Foreground(zstyle.Surface1)
	maxW := m.width
	if maxW <= 0 {
		maxW = 60
	}
	b.WriteString(sep.Render(strings.Repeat("─", maxW)))
	b.WriteString("\n")

	// content
	b.WriteString(m.viewContent())

	// error display (non-password views)
	if m.err != "" && m.view != viewPassword {
		errText := zstyle.StatusErr.Render("  " + m.err)
		b.WriteString("\n" + errText + "\n")
	}

	// footer spacer + footer
	b.WriteString("\n")
	b.WriteString(renderFooter(m.view, m.width))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewContent() string {
	switch m.view {
	case viewPassword:
		return m.password.View()
	case viewMenu:
		return m.menu.View()
	case viewSecretList:
		return m.secretList.View()
	case viewSecretDetail:
		return m.secretDetail.View()
	case viewSecretForm:
		return m.secretForm.View()
	case viewTaskList:
		return m.taskList.View()
	case viewTaskDetail:
		return m.taskDetail.View()
	case viewTaskForm:
		return m.taskForm.View()
	default:
		return fmt.Sprintf("  unknown view: %d", m.view)
	}
}

func (m Model) propagateSize() Model {
	m.password.width = m.width
	m.password.height = m.height
	m.menu.width = m.width
	m.menu.height = m.height
	m.secretList.width = m.width
	m.secretList.height = m.height
	m.secretDetail.width = m.width
	m.secretDetail.height = m.height
	m.secretForm.width = m.width
	m.secretForm.height = m.height
	m.taskList.width = m.width
	m.taskList.height = m.height
	m.taskDetail.width = m.width
	m.taskDetail.height = m.height
	m.taskForm.width = m.width
	m.taskForm.height = m.height
	return m
}

// CurrentView returns the active view ID (exported for testing).
func (m Model) CurrentView() viewID { return m.view }

// pendingFilter returns a task filter for pending tasks.
func pendingFilter() task.Filter {
	return task.Filter{Status: task.FilterPending}
}
