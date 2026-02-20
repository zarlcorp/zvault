package tui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zarlcorp/core/pkg/zstyle"
	"github.com/zarlcorp/zvault/internal/secret"
	"github.com/zarlcorp/zvault/internal/totp"
	"github.com/zarlcorp/zvault/internal/vault"
)

// totpTickMsg triggers TOTP code refresh.
type totpTickMsg time.Time

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

	// TOTP live code
	totpCode      string
	totpRemaining int
	hasTOTP       bool

	width  int
	height int
}

// fieldAction determines what Enter does on a field.
type fieldAction int

const (
	actionNone fieldAction = iota
	actionCopy             // copy value to clipboard
	actionOpen             // open value as URL
)

type detailField struct {
	label      string
	value      string
	sensitive  bool
	labelColor lipgloss.Color // per-field label color
	live       bool           // live-updating field (e.g. TOTP code)
	action     fieldAction    // what Enter does
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
	m.hasTOTP = s.TOTPSecret() != ""
	if m.hasTOTP {
		m.refreshTOTP()
	}
	if m.cursor >= len(m.fields) {
		m.cursor = 0
	}
	return m
}

func (m *secretDetailModel) refreshTOTP() {
	code, remaining, err := totp.Generate(m.secret.TOTPSecret())
	if err != nil {
		m.totpCode = ""
		m.totpRemaining = 0
		return
	}
	m.totpCode = code
	m.totpRemaining = remaining
}

func totpTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return totpTickMsg(t)
	})
}

func buildDetailFields(s secret.Secret) []detailField {
	var fields []detailField

	// name rendered as header
	fields = append(fields, detailField{label: "name", value: s.Name, labelColor: zstyle.ZvaultAccent})
	// type uses the standard label color — the value gets special rendering in View()
	fields = append(fields, detailField{label: "type", value: string(s.Type), labelColor: zstyle.Lavender})

	sensitiveColor := zstyle.Peach
	normalColor := zstyle.Lavender

	switch s.Type {
	case secret.TypePassword:
		fields = append(fields, detailField{label: "url", value: s.URL(), labelColor: normalColor, action: actionOpen})
		fields = append(fields, detailField{label: "username", value: s.Username(), labelColor: normalColor, action: actionCopy})
		fields = append(fields, detailField{label: "password", value: s.Password(), sensitive: true, labelColor: sensitiveColor, action: actionCopy})
		if s.TOTPSecret() != "" {
			fields = append(fields, detailField{label: "totp secret", value: s.TOTPSecret(), sensitive: true, labelColor: sensitiveColor})
			fields = append(fields, detailField{label: "totp code", live: true, labelColor: zstyle.Green, action: actionCopy})
		}
		if s.Notes() != "" {
			fields = append(fields, detailField{label: "notes", value: s.Notes(), labelColor: normalColor})
		}
	case secret.TypeAPIKey:
		fields = append(fields, detailField{label: "service", value: s.Service(), labelColor: normalColor})
		fields = append(fields, detailField{label: "key", value: s.Key(), sensitive: true, labelColor: sensitiveColor, action: actionCopy})
		if s.Notes() != "" {
			fields = append(fields, detailField{label: "notes", value: s.Notes(), labelColor: normalColor})
		}
	case secret.TypeSSHKey:
		fields = append(fields, detailField{label: "label", value: s.Label(), labelColor: normalColor})
		fields = append(fields, detailField{label: "private key", value: s.PrivateKey(), sensitive: true, labelColor: sensitiveColor, action: actionCopy})
		fields = append(fields, detailField{label: "public key", value: s.PublicKey(), labelColor: normalColor, action: actionCopy})
		if s.Passphrase() != "" {
			fields = append(fields, detailField{label: "passphrase", value: s.Passphrase(), sensitive: true, labelColor: sensitiveColor, action: actionCopy})
		}
		if s.Notes() != "" {
			fields = append(fields, detailField{label: "notes", value: s.Notes(), labelColor: normalColor})
		}
	case secret.TypeNote:
		fields = append(fields, detailField{label: "content", value: s.Content(), labelColor: normalColor})
	}

	metaColor := zstyle.Subtext1

	if len(s.Tags) > 0 {
		fields = append(fields, detailField{label: "tags", value: strings.Join(s.Tags, ", "), labelColor: normalColor})
	}

	fields = append(fields, detailField{label: "created", value: s.CreatedAt.Format("2006-01-02 15:04"), labelColor: metaColor})
	fields = append(fields, detailField{label: "updated", value: s.UpdatedAt.Format("2006-01-02 15:04"), labelColor: metaColor})

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
				if m.hasTOTP {
					return m, totpTickCmd()
				}
			}
		}
		return m, nil

	case totpTickMsg:
		if !m.hasTOTP {
			return m, nil
		}
		m.refreshTOTP()
		return m, totpTickCmd()

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
	case key.Matches(msg, zstyle.KeyEnter):
		if m.cursor < len(m.fields) {
			return m.handleFieldAction(m.fields[m.cursor])
		}
	case msg.String() == "s":
		m.showSensitive = !m.showSensitive
	case msg.String() == "e":
		return m, func() tea.Msg {
			return navigateMsg{view: viewSecretForm, data: m.secretID}
		}
	case msg.String() == "d":
		m.confirmDelete = true
	}
	return m, nil
}

func (m secretDetailModel) handleFieldAction(f detailField) (secretDetailModel, tea.Cmd) {
	switch f.action {
	case actionCopy:
		return m.handleFieldCopy(f)
	case actionOpen:
		if f.value != "" {
			return m, openURL(f.value)
		}
	}
	return m, nil
}

func (m secretDetailModel) handleFieldCopy(f detailField) (secretDetailModel, tea.Cmd) {
	val := f.value
	if f.live && m.totpCode != "" {
		val = m.totpCode
	}
	return m, copyToClipboard(f.label, val)
}

// openURL opens a URL in the default browser.
func openURL(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "linux":
			cmd = exec.Command("xdg-open", url)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", url)
		default:
			return nil
		}
		_ = cmd.Start()
		return nil
	}
}

func (m secretDetailModel) View() string {
	var b strings.Builder

	if m.secretID == "" {
		b.WriteString("\n")
		b.WriteString(zstyle.MutedText.Render("  no secret selected"))
		b.WriteString("\n")
		return b.String()
	}

	valueStyle := lipgloss.NewStyle().Foreground(zstyle.Text)
	maskedStyle := lipgloss.NewStyle().Foreground(zstyle.Surface2)
	cursorStyle := lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent)
	codeStyle := lipgloss.NewStyle().Foreground(zstyle.Green)
	countdownStyle := lipgloss.NewStyle().Foreground(zstyle.Overlay1)

	b.WriteString("\n")
	for i, f := range m.fields {
		prefix := "  "
		if i == m.cursor {
			prefix = cursorStyle.Render("▸ ")
		}

		// name field: bold + accent
		labelStyle := lipgloss.NewStyle().Foreground(f.labelColor).Width(14)
		if f.label == "name" {
			labelStyle = labelStyle.Bold(true)
		}
		label := labelStyle.Render(f.label)

		var val string
		switch {
		case f.live:
			// TOTP code: show generated code + countdown
			if m.totpCode != "" {
				val = codeStyle.Render(m.totpCode) + " " + countdownStyle.Render(fmt.Sprintf("(%ds)", m.totpRemaining))
			} else {
				val = maskedStyle.Render("generating...")
			}
		case f.label == "type":
			// render type as badge
			val = typeBadge(secret.Type(f.value))
		case f.sensitive && !m.showSensitive:
			val = maskedStyle.Render("••••••••")
		default:
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
