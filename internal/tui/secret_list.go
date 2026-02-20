package tui

import (
	"fmt"
	"sort"
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
	filterByTag
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

	// tag filtering
	tags     []string // unique tags across all secrets
	tagIndex int      // selected tag when cycling

	// search
	searching bool
	search    textinput.Model

	// delete confirmation
	confirmDelete bool

	// status/error messages
	status string
	err    string

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
		m.err = fmt.Sprintf("load secrets: %v", err)
		return m
	}

	m.err = ""

	// collect tags from unfiltered set
	m.collectTags(all)

	// apply type filter
	if m.filter != filterAll && m.filter != filterByTag {
		ft := m.filter.secretType()
		var filtered []secret.Secret
		for _, s := range all {
			if s.Type == ft {
				filtered = append(filtered, s)
			}
		}
		all = filtered
	}

	// apply tag filter
	if m.filter == filterByTag && len(m.tags) > 0 {
		tag := m.tags[m.tagIndex]
		var filtered []secret.Secret
		for _, s := range all {
			for _, st := range s.Tags {
				if st == tag {
					filtered = append(filtered, s)
					break
				}
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

func (m *secretListModel) collectTags(secrets []secret.Secret) {
	m.tags = nil
	seen := make(map[string]bool)
	for _, s := range secrets {
		for _, tag := range s.Tags {
			if !seen[tag] {
				seen[tag] = true
				m.tags = append(m.tags, tag)
			}
		}
	}
	sort.Strings(m.tags)
}

// advanceFilter moves to the next filter mode. In tag mode, tab cycles
// through tags; after the last tag it wraps to all. If no tags exist,
// tag mode is skipped.
func (m *secretListModel) advanceFilter() {
	if m.filter == filterByTag {
		m.tagIndex++
		if m.tagIndex >= len(m.tags) {
			m.tagIndex = 0
			m.filter = filterAll
		}
		return
	}

	next := (m.filter + 1) % filterCount
	if next == filterByTag && len(m.tags) == 0 {
		next = filterAll
	}
	if next == filterByTag {
		m.tagIndex = 0
	}
	m.filter = next
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
		m.advanceFilter()
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

	// filter tabs (type filters, then active tag if in tag mode)
	b.WriteString("\n  ")
	activeStyle := lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(zstyle.Overlay1)
	sepStyle := lipgloss.NewStyle().Foreground(zstyle.Surface2)
	for i := typeFilter(0); i <= filterNote; i++ {
		label := i.label()
		if i == m.filter {
			b.WriteString(activeStyle.Render(label))
		} else {
			b.WriteString(inactiveStyle.Render(label))
		}
		b.WriteString(sepStyle.Render(" | "))
	}
	// tag indicator
	if m.filter == filterByTag && len(m.tags) > 0 {
		b.WriteString(activeStyle.Render("#" + m.tags[m.tagIndex]))
	} else if len(m.tags) > 0 {
		b.WriteString(inactiveStyle.Render("tags"))
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
		b.WriteString(warn.Render(fmt.Sprintf("  delete '%s'? (y/n)", s.Name)))
		b.WriteString("\n")
	}

	// error message
	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(zstyle.StatusErr.Render("  " + m.err))
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
