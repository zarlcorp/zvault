package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zarlcorp/core/pkg/zstyle"
	"github.com/zarlcorp/zvault/internal/secret"
	"github.com/zarlcorp/zvault/internal/vault"
)

// typeFilter enumerates the type filter options.
type typeFilter int

const (
	filterAll typeFilter = iota
	filterPassword
	filterAPIKey
	filterSSHKey
	filterNote
	filterCount // sentinel
)

func (f typeFilter) label() string {
	switch f {
	case filterAll:
		return "all"
	case filterPassword:
		return "password"
	case filterAPIKey:
		return "api key"
	case filterSSHKey:
		return "ssh key"
	case filterNote:
		return "note"
	}
	return ""
}

func (f typeFilter) secretType() secret.Type {
	switch f {
	case filterPassword:
		return secret.TypePassword
	case filterAPIKey:
		return secret.TypeAPIKey
	case filterSSHKey:
		return secret.TypeSSHKey
	case filterNote:
		return secret.TypeNote
	}
	return ""
}

// secretListModel displays a scrollable, filterable list of secrets.
type secretListModel struct {
	vault   *vault.Vault
	secrets []secret.Secret // current filtered set
	cursor  int
	filter  typeFilter

	// search
	searching bool
	search    textinput.Model

	// delete confirmation
	confirmDelete bool

	// status message
	status string

	width  int
	height int
}

func newSecretList() secretListModel {
	si := textinput.New()
	si.Placeholder = "search..."
	si.PromptStyle = lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent)
	si.TextStyle = lipgloss.NewStyle().Foreground(zstyle.Text)
	return secretListModel{search: si}
}

func (m secretListModel) loadSecrets() secretListModel {
	if m.vault == nil {
		m.secrets = nil
		return m
	}

	var all []secret.Secret
	var err error

	query := m.search.Value()
	if m.searching && query != "" {
		all, err = m.vault.Secrets().Search(query)
	} else {
		all, err = m.vault.Secrets().List()
	}
	if err != nil {
		m.secrets = nil
		return m
	}

	// apply type filter
	if m.filter != filterAll {
		ft := m.filter.secretType()
		var filtered []secret.Secret
		for _, s := range all {
			if s.Type == ft {
				filtered = append(filtered, s)
			}
		}
		all = filtered
	}

	m.secrets = all
	if m.cursor >= len(m.secrets) {
		m.cursor = max(0, len(m.secrets)-1)
	}
	return m
}

func (m secretListModel) Update(msg tea.Msg) (secretListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case navigateMsg:
		// refresh when navigated to
		if msg.view == viewSecretList {
			m.confirmDelete = false
			m.status = ""
			m = m.loadSecrets()
		}
		return m, nil

	case tea.KeyMsg:
		// handle delete confirmation
		if m.confirmDelete {
			return m.handleDeleteConfirm(msg)
		}

		// handle search mode input
		if m.searching {
			return m.handleSearchInput(msg)
		}

		return m.handleNormalKeys(msg)
	}
	return m, nil
}

func (m secretListModel) handleDeleteConfirm(msg tea.KeyMsg) (secretListModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.cursor < len(m.secrets) {
			s := m.secrets[m.cursor]
			if m.vault != nil {
				if err := m.vault.Secrets().Delete(s.ID); err != nil {
					m.confirmDelete = false
					return m, func() tea.Msg { return errMsg{err: err} }
				}
			}
			m.status = fmt.Sprintf("deleted '%s'", s.Name)
			m.confirmDelete = false
			m = m.loadSecrets()
		}
	case "n", "N", "esc":
		m.confirmDelete = false
	}
	return m, nil
}

func (m secretListModel) handleSearchInput(msg tea.KeyMsg) (secretListModel, tea.Cmd) {
	switch {
	case key.Matches(msg, zstyle.KeyBack):
		m.searching = false
		m.search.Blur()
		m.search.SetValue("")
		m = m.loadSecrets()
		return m, nil
	case key.Matches(msg, zstyle.KeyEnter):
		m.searching = false
		m.search.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.search, cmd = m.search.Update(msg)
	m = m.loadSecrets()
	return m, cmd
}

func (m secretListModel) handleNormalKeys(msg tea.KeyMsg) (secretListModel, tea.Cmd) {
	switch {
	case key.Matches(msg, zstyle.KeyUp):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, zstyle.KeyDown):
		if m.cursor < len(m.secrets)-1 {
			m.cursor++
		}
	case key.Matches(msg, zstyle.KeyEnter):
		if m.cursor < len(m.secrets) {
			s := m.secrets[m.cursor]
			return m, func() tea.Msg {
				return navigateMsg{view: viewSecretDetail, data: s.ID}
			}
		}
	case key.Matches(msg, zstyle.KeyBack):
		parent := parentView(viewSecretList)
		return m, func() tea.Msg { return navigateMsg{view: parent} }
	case key.Matches(msg, zstyle.KeyFilter):
		m.searching = true
		m.search.Focus()
		m.status = ""
		return m, textinput.Blink
	case msg.String() == "tab":
		m.filter = (m.filter + 1) % filterCount
		m.status = ""
		m = m.loadSecrets()
	case msg.String() == "n":
		return m, func() tea.Msg {
			return navigateMsg{view: viewSecretForm, data: nil}
		}
	case msg.String() == "d":
		if len(m.secrets) > 0 && m.cursor < len(m.secrets) {
			m.confirmDelete = true
			m.status = ""
		}
	}
	return m, nil
}

func (m secretListModel) View() string {
	var b strings.Builder

	// filter tabs
	b.WriteString("\n  ")
	for i := typeFilter(0); i < filterCount; i++ {
		label := i.label()
		if i == m.filter {
			style := lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent).Bold(true)
			b.WriteString(style.Render(label))
		} else {
			style := lipgloss.NewStyle().Foreground(zstyle.Overlay1)
			b.WriteString(style.Render(label))
		}
		if i < filterCount-1 {
			sep := lipgloss.NewStyle().Foreground(zstyle.Surface2).Render(" | ")
			b.WriteString(sep)
		}
	}
	b.WriteString("\n")

	// search bar
	if m.searching {
		b.WriteString("  " + m.search.View() + "\n")
	} else if m.search.Value() != "" {
		muted := zstyle.MutedText.Render(fmt.Sprintf("  search: %s", m.search.Value()))
		b.WriteString(muted + "\n")
	}

	b.WriteString("\n")

	if len(m.secrets) == 0 {
		msg := zstyle.MutedText.Render("  no secrets found")
		b.WriteString(msg + "\n")
	} else {
		// calculate visible area (rough: height minus header/footer/filter/padding)
		visibleHeight := m.height - 10
		if visibleHeight < 3 {
			visibleHeight = 3
		}

		// scroll window
		start := 0
		if m.cursor >= visibleHeight {
			start = m.cursor - visibleHeight + 1
		}
		end := start + visibleHeight
		if end > len(m.secrets) {
			end = len(m.secrets)
		}

		for i := start; i < end; i++ {
			s := m.secrets[i]
			cursor := "  "
			nameStyle := lipgloss.NewStyle().Foreground(zstyle.Text)
			if i == m.cursor {
				cursor = lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent).Render("▸ ")
				nameStyle = lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent).Bold(true)
			}

			name := nameStyle.Render(s.Name)
			badge := typeBadge(s.Type)
			tags := ""
			if len(s.Tags) > 0 {
				tags = zstyle.MutedText.Render(" [" + strings.Join(s.Tags, ", ") + "]")
			}

			b.WriteString(fmt.Sprintf("  %s%s %s%s\n", cursor, name, badge, tags))
		}
	}

	// delete confirmation
	if m.confirmDelete && m.cursor < len(m.secrets) {
		s := m.secrets[m.cursor]
		warn := lipgloss.NewStyle().Foreground(zstyle.Warning)
		b.WriteString("\n")
		b.WriteString(warn.Render(fmt.Sprintf("  Delete '%s'? (y/n)", s.Name)))
		b.WriteString("\n")
	}

	// status message
	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(zstyle.StatusOK.Render("  " + m.status))
		b.WriteString("\n")
	}

	return b.String()
}

// typeBadge returns a styled badge for the secret type.
func typeBadge(t secret.Type) string {
	var label string
	var color lipgloss.Color
	switch t {
	case secret.TypePassword:
		label = "pw"
		color = zstyle.Blue
	case secret.TypeAPIKey:
		label = "api"
		color = zstyle.Peach
	case secret.TypeSSHKey:
		label = "ssh"
		color = zstyle.Green
	case secret.TypeNote:
		label = "note"
		color = zstyle.Yellow
	default:
		label = string(t)
		color = zstyle.Overlay1
	}
	return lipgloss.NewStyle().Foreground(color).Render("[" + label + "]")
}
