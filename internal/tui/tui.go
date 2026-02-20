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
	password     passwordModel
	menu         menuModel
	secretList   secretListModel
	secretDetail secretDetailModel
	secretForm   secretFormModel
	taskList     taskListModel
	taskDetail   taskDetailModel
	taskForm     taskFormModel

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
		secretList:   newSecretList(),
		secretDetail: newSecretDetail(),
		secretForm:   newSecretForm(),
		taskList:     newTaskListModel(nil),
		taskDetail:   newTaskDetailModel(nil),
		taskForm:     newTaskFormModel(nil),
	}
}

// NewWithDir creates a root model with a custom vault directory (for testing).
func NewWithDir(version, vaultDir string) Model {
	return Model{
		version:      version,
		view:         viewPassword,
		password:     newPasswordModel(vaultDir),
		menu:         newMenuModel(),
		secretList:   newSecretList(),
		secretDetail: newSecretDetail(),
		secretForm:   newSecretForm(),
		taskList:     newTaskListModel(nil),
		taskDetail:   newTaskDetailModel(nil),
		taskForm:     newTaskFormModel(nil),
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
		// propagate navigateMsg to target sub-view so it can load data
		var cmd tea.Cmd
		switch msg.view {
		case viewSecretList:
			m.secretList.vault = m.vault
			m.secretList, _ = m.secretList.Update(msg)
		case viewSecretDetail:
			m.secretDetail.vault = m.vault
			m.secretDetail, cmd = m.secretDetail.Update(msg)
		case viewSecretForm:
			m.secretForm.vault = m.vault
			m.secretForm, _ = m.secretForm.Update(msg)
		case viewTaskList:
			m.taskList.vault = m.vault
			m.taskList, _ = m.taskList.Update(msg)
		case viewTaskDetail:
			m.taskDetail.vault = m.vault
			m.taskDetail, _ = m.taskDetail.Update(msg)
		case viewTaskForm:
			m.taskForm.vault = m.vault
			m.taskForm, _ = m.taskForm.Update(msg)
		}
		return m, cmd

	case vaultOpenedMsg:
		m.vault = msg.vault
		m.view = viewMenu
		m.menu = m.menu.refreshCounts(msg.vault)
		// propagate vault to secret views
		m.secretList.vault = msg.vault
		m.secretDetail.vault = msg.vault
		m.secretForm.vault = msg.vault
		// propagate vault to task views
		m.taskList.vault = msg.vault
		m.taskDetail.vault = msg.vault
		m.taskForm.vault = msg.vault
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
		if m.isTextInputActive() {
			// only ctrl+c quits when a text input is focused
			if msg.Type == tea.KeyCtrlC {
				return m, tea.Quit
			}
		} else {
			if key.Matches(msg, zstyle.KeyQuit) {
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
	w, h := m.width, m.height
	m.password.width = w
	m.password.height = h
	m.menu.width = w
	m.menu.height = h
	m.secretList.width = w
	m.secretList.height = h
	m.secretDetail.width = w
	m.secretDetail.height = h
	m.secretForm.width = w
	m.secretForm.height = h
	m.taskList.width = w
	m.taskList.height = h
	m.taskDetail.width = w
	m.taskDetail.height = h
	m.taskForm.width = w
	m.taskForm.height = h
	return m
}

// CurrentView returns the active view ID (exported for testing).
func (m Model) CurrentView() viewID { return m.view }

// isTextInputActive returns true when a text input is focused and q should not quit.
func (m Model) isTextInputActive() bool {
	switch m.view {
	case viewPassword:
		return true
	case viewSecretList:
		return m.secretList.searching
	case viewSecretForm:
		return true
	case viewTaskForm:
		return true
	}
	return false
}

// pendingFilter returns a task filter for pending tasks.
func pendingFilter() task.Filter {
	return task.Filter{Status: task.FilterPending}
}
