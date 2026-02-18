package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zarlcorp/core/pkg/zstyle"
	"github.com/zarlcorp/zvault/internal/secret"
	"github.com/zarlcorp/zvault/internal/vault"
)

// secretDetailModel displays a single secret's fields.
type secretDetailModel struct {
	vault    *vault.Vault
	secret   secret.Secret
	secretID string
	showSensitive bool

	// clipboard feedback
	clipboardMsg string

	// delete confirmation
	confirmDelete bool

	// field cursor for copy
	cursor int
	fields []detailField

	width  int
	height int
}

type detailField struct {
	label     string
	value     string
	sensitive bool
}

func newSecretDetail() secretDetailModel {
	return secretDetailModel{}
}

func (m secretDetailModel) load() secretDetailModel {
	if m.vault == nil || m.secretID == "" {
		return m
	}
	s, err := m.vault.Secrets().Get(m.secretID)
	if err != nil {
		return m
	}
	m.secret = s
	m.fields = buildDetailFields(s)
	if m.cursor >= len(m.fields) {
		m.cursor = 0
	}
	return m
}

func buildDetailFields(s secret.Secret) []detailField {
	var fields []detailField

	fields = append(fields, detailField{label: "name", value: s.Name})
	fields = append(fields, detailField{label: "type", value: string(s.Type)})

	switch s.Type {
	case secret.TypePassword:
		fields = append(fields, detailField{label: "url", value: s.URL()})
		fields = append(fields, detailField{label: "username", value: s.Username()})
		fields = append(fields, detailField{label: "password", value: s.Password(), sensitive: true})
		if s.TOTPSecret() != "" {
			fields = append(fields, detailField{label: "totp secret", value: s.TOTPSecret(), sensitive: true})
		}
		if s.Notes() != "" {
			fields = append(fields, detailField{label: "notes", value: s.Notes()})
		}
	case secret.TypeAPIKey:
		fields = append(fields, detailField{label: "service", value: s.Service()})
		fields = append(fields, detailField{label: "key", value: s.Key(), sensitive: true})
		if s.Notes() != "" {
			fields = append(fields, detailField{label: "notes", value: s.Notes()})
		}
	case secret.TypeSSHKey:
		fields = append(fields, detailField{label: "label", value: s.Label()})
		fields = append(fields, detailField{label: "private key", value: s.PrivateKey(), sensitive: true})
		fields = append(fields, detailField{label: "public key", value: s.PublicKey()})
		if s.Passphrase() != "" {
			fields = append(fields, detailField{label: "passphrase", value: s.Passphrase(), sensitive: true})
		}
		if s.Notes() != "" {
			fields = append(fields, detailField{label: "notes", value: s.Notes()})
		}
	case secret.TypeNote:
		fields = append(fields, detailField{label: "content", value: s.Content()})
	}

	if len(s.Tags) > 0 {
		fields = append(fields, detailField{label: "tags", value: strings.Join(s.Tags, ", ")})
	}

	fields = append(fields, detailField{label: "created", value: s.CreatedAt.Format("2006-01-02 15:04")})
	fields = append(fields, detailField{label: "updated", value: s.UpdatedAt.Format("2006-01-02 15:04")})

	return fields
}

func (m secretDetailModel) Update(msg tea.Msg) (secretDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case navigateMsg:
		if msg.view == viewSecretDetail {
			if id, ok := msg.data.(string); ok {
				m.secretID = id
				m.showSensitive = false
				m.confirmDelete = false
				m.clipboardMsg = ""
				m.cursor = 0
				m = m.load()
			}
		}
		return m, nil

	case clipboardCopiedMsg:
		m.clipboardMsg = fmt.Sprintf("copied %s (clears in 10s)", msg.field)
		return m, nil

	case clipboardClearedMsg:
		m.clipboardMsg = ""
		return m, nil

	case tea.KeyMsg:
		if m.confirmDelete {
			return m.handleDeleteConfirm(msg)
		}
		return m.handleKeys(msg)
	}
	return m, nil
}

func (m secretDetailModel) handleDeleteConfirm(msg tea.KeyMsg) (secretDetailModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.vault != nil {
			if err := m.vault.Secrets().Delete(m.secretID); err != nil {
				m.confirmDelete = false
				return m, func() tea.Msg { return errMsg{err: err} }
			}
		}
		m.confirmDelete = false
		return m, func() tea.Msg { return navigateMsg{view: viewSecretList} }
	case "n", "N", "esc":
		m.confirmDelete = false
	}
	return m, nil
}

func (m secretDetailModel) handleKeys(msg tea.KeyMsg) (secretDetailModel, tea.Cmd) {
	switch {
	case key.Matches(msg, zstyle.KeyBack):
		return m, func() tea.Msg { return navigateMsg{view: viewSecretList} }
	case key.Matches(msg, zstyle.KeyUp):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, zstyle.KeyDown):
		if m.cursor < len(m.fields)-1 {
			m.cursor++
		}
	case msg.String() == "s":
		m.showSensitive = !m.showSensitive
	case msg.String() == "c":
		if m.cursor < len(m.fields) {
			f := m.fields[m.cursor]
			return m, copyToClipboard(f.label, f.value)
		}
	case msg.String() == "e":
		return m, func() tea.Msg {
			return navigateMsg{view: viewSecretForm, data: m.secretID}
		}
	case msg.String() == "d":
		m.confirmDelete = true
	}
	return m, nil
}

func (m secretDetailModel) View() string {
	var b strings.Builder

	if m.secretID == "" {
		b.WriteString("\n")
		b.WriteString(zstyle.MutedText.Render("  no secret selected"))
		b.WriteString("\n")
		return b.String()
	}

	labelStyle := lipgloss.NewStyle().Foreground(zstyle.Subtext1).Width(14)
	valueStyle := lipgloss.NewStyle().Foreground(zstyle.Text)
	maskedStyle := lipgloss.NewStyle().Foreground(zstyle.Surface2)
	cursorStyle := lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent)

	b.WriteString("\n")
	for i, f := range m.fields {
		prefix := "  "
		if i == m.cursor {
			prefix = cursorStyle.Render("▸ ")
		}

		label := labelStyle.Render(f.label)
		var val string
		if f.sensitive && !m.showSensitive {
			val = maskedStyle.Render("••••••••")
		} else {
			val = valueStyle.Render(f.value)
		}

		b.WriteString(fmt.Sprintf("  %s%s  %s\n", prefix, label, val))
	}

	// delete confirmation
	if m.confirmDelete {
		warn := lipgloss.NewStyle().Foreground(zstyle.Warning)
		b.WriteString("\n")
		b.WriteString(warn.Render(fmt.Sprintf("  delete '%s'? (y/n)", m.secret.Name)))
		b.WriteString("\n")
	}

	// clipboard status
	if m.clipboardMsg != "" {
		b.WriteString("\n")
		b.WriteString(zstyle.StatusOK.Render("  " + m.clipboardMsg))
		b.WriteString("\n")
	}

	return b.String()
}
