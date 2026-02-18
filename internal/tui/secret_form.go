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

// formMode indicates create vs edit.
type formMode int

const (
	formCreate formMode = iota
	formEdit
)

// secretFormModel handles create/edit of secrets.
type secretFormModel struct {
	vault    *vault.Vault
	mode     formMode
	editID   string // set in edit mode
	secType  secret.Type
	typeSel  int // index in typeOptions (create only)
	inputs   []formInput
	focused  int
	changed  bool
	err      string

	// confirm discard on esc if changed
	confirmDiscard bool

	width  int
	height int
}

type formInput struct {
	label    string
	fieldKey string
	input    textinput.Model
	masked   bool
}

var typeOptions = []secret.Type{
	secret.TypePassword,
	secret.TypeAPIKey,
	secret.TypeSSHKey,
	secret.TypeNote,
}

func typeLabel(t secret.Type) string {
	switch t {
	case secret.TypePassword:
		return "password"
	case secret.TypeAPIKey:
		return "api key"
	case secret.TypeSSHKey:
		return "ssh key"
	case secret.TypeNote:
		return "note"
	}
	return string(t)
}

func newSecretForm() secretFormModel {
	return secretFormModel{
		secType: secret.TypePassword,
	}
}

func (m secretFormModel) initForCreate() secretFormModel {
	m.mode = formCreate
	m.editID = ""
	m.typeSel = 0
	m.secType = typeOptions[0]
	m.changed = false
	m.err = ""
	m.confirmDiscard = false
	m.inputs = buildFormInputs(m.secType, secret.Secret{})
	m.focused = -1 // type selector focused first
	m.focusCurrent()
	return m
}

func (m secretFormModel) initForEdit(id string) secretFormModel {
	m.mode = formEdit
	m.editID = id
	m.changed = false
	m.err = ""
	m.confirmDiscard = false

	if m.vault == nil {
		return m
	}

	s, err := m.vault.Secrets().Get(id)
	if err != nil {
		m.err = err.Error()
		return m
	}

	m.secType = s.Type
	for i, t := range typeOptions {
		if t == s.Type {
			m.typeSel = i
			break
		}
	}
	m.inputs = buildFormInputs(s.Type, s)
	m.focused = 0
	m.focusCurrent()
	return m
}

func buildFormInputs(t secret.Type, s secret.Secret) []formInput {
	var inputs []formInput

	addInput := func(label, fieldKey, value string, masked bool) {
		ti := textinput.New()
		ti.Placeholder = label
		ti.PromptStyle = lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent)
		ti.TextStyle = lipgloss.NewStyle().Foreground(zstyle.Text)
		if masked {
			ti.EchoMode = textinput.EchoPassword
			ti.EchoCharacter = '•'
		}
		ti.SetValue(value)
		inputs = append(inputs, formInput{
			label:    label,
			fieldKey: fieldKey,
			input:    ti,
			masked:   masked,
		})
	}

	// name is always first
	addInput("name", "name", s.Name, false)

	switch t {
	case secret.TypePassword:
		addInput("url", "url", s.URL(), false)
		addInput("username", "username", s.Username(), false)
		addInput("password", "password", s.Password(), true)
		addInput("totp secret", "totp_secret", s.TOTPSecret(), true)
		addInput("notes", "notes", s.Notes(), false)
	case secret.TypeAPIKey:
		addInput("service", "service", s.Service(), false)
		addInput("key", "key", s.Key(), true)
		addInput("notes", "notes", s.Notes(), false)
	case secret.TypeSSHKey:
		addInput("label", "label", s.Label(), false)
		addInput("private key", "private_key", s.PrivateKey(), true)
		addInput("public key", "public_key", s.PublicKey(), false)
		addInput("passphrase", "passphrase", s.Passphrase(), true)
		addInput("notes", "notes", s.Notes(), false)
	case secret.TypeNote:
		addInput("content", "content", s.Content(), false)
	}

	// tags always last
	tags := ""
	if len(s.Tags) > 0 {
		tags = strings.Join(s.Tags, ", ")
	}
	addInput("tags", "tags", tags, false)

	return inputs
}

func (m *secretFormModel) focusCurrent() {
	for i := range m.inputs {
		if i == m.focused {
			m.inputs[i].input.Focus()
		} else {
			m.inputs[i].input.Blur()
		}
	}
}

func (m secretFormModel) Update(msg tea.Msg) (secretFormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case navigateMsg:
		if msg.view == viewSecretForm {
			if msg.data == nil {
				m = m.initForCreate()
				return m, textinput.Blink
			}
			if id, ok := msg.data.(string); ok {
				m = m.initForEdit(id)
				return m, textinput.Blink
			}
		}
		return m, nil

	case tea.KeyMsg:
		if m.confirmDiscard {
			return m.handleDiscardConfirm(msg)
		}
		return m.handleKeys(msg)
	}

	// pass to focused input
	if m.focused >= 0 && m.focused < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focused].input, cmd = m.inputs[m.focused].input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m secretFormModel) handleDiscardConfirm(msg tea.KeyMsg) (secretFormModel, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.confirmDiscard = false
		return m, func() tea.Msg { return navigateMsg{view: viewSecretList} }
	case "n", "N", "esc":
		m.confirmDiscard = false
	}
	return m, nil
}

func (m secretFormModel) handleKeys(msg tea.KeyMsg) (secretFormModel, tea.Cmd) {
	minFocus := 0
	if m.mode == formCreate {
		minFocus = -1 // type selector position
	}

	switch {
	case msg.Type == tea.KeyCtrlS:
		return m.save()

	case key.Matches(msg, zstyle.KeyBack):
		if m.changed {
			m.confirmDiscard = true
			return m, nil
		}
		return m, func() tea.Msg { return navigateMsg{view: viewSecretList} }

	case msg.Type == tea.KeyShiftTab:
		m.changed = true
		if m.focused > minFocus {
			m.focused--
			m.focusCurrent()
		}
		return m, nil

	case key.Matches(msg, zstyle.KeyTab):
		m.changed = true
		if m.focused < len(m.inputs)-1 {
			m.focused++
			m.focusCurrent()
		}
		return m, nil

	case key.Matches(msg, zstyle.KeyEnter):
		// enter on type selector moves to first input
		if m.focused == -1 {
			m.focused = 0
			m.focusCurrent()
			return m, nil
		}
		// enter on last field saves
		if m.focused == len(m.inputs)-1 {
			return m.save()
		}
		// otherwise, next field
		m.changed = true
		if m.focused < len(m.inputs)-1 {
			m.focused++
			m.focusCurrent()
		}
		return m, nil

	case msg.String() == "left" && m.mode == formCreate && m.focused == -1:
		// cycle type backward (only when type selector is focused)
		if m.typeSel > 0 {
			m.typeSel--
		} else {
			m.typeSel = len(typeOptions) - 1
		}
		m.secType = typeOptions[m.typeSel]
		m.rebuildInputsPreserveName()
		return m, nil

	case msg.String() == "right" && m.mode == formCreate && m.focused == -1:
		// cycle type forward (only when type selector is focused)
		m.typeSel = (m.typeSel + 1) % len(typeOptions)
		m.secType = typeOptions[m.typeSel]
		m.rebuildInputsPreserveName()
		return m, nil
	}

	// pass to focused input (only when focused >= 0)
	if m.focused >= 0 && m.focused < len(m.inputs) {
		var cmd tea.Cmd
		old := m.inputs[m.focused].input.Value()
		m.inputs[m.focused].input, cmd = m.inputs[m.focused].input.Update(msg)
		if m.inputs[m.focused].input.Value() != old {
			m.changed = true
		}
		return m, cmd
	}
	return m, nil
}

func (m *secretFormModel) rebuildInputsPreserveName() {
	// preserve name value
	name := ""
	if len(m.inputs) > 0 {
		name = m.inputs[0].input.Value()
	}
	m.inputs = buildFormInputs(m.secType, secret.Secret{})
	if len(m.inputs) > 0 {
		m.inputs[0].input.SetValue(name)
	}
	if m.focused >= len(m.inputs) {
		m.focused = 0
	}
	m.focusCurrent()
}

func (m secretFormModel) save() (secretFormModel, tea.Cmd) {
	// collect values
	vals := make(map[string]string)
	for _, inp := range m.inputs {
		vals[inp.fieldKey] = inp.input.Value()
	}

	name := strings.TrimSpace(vals["name"])
	if name == "" {
		m.err = "name is required"
		return m, nil
	}

	// parse tags
	var tags []string
	if t := strings.TrimSpace(vals["tags"]); t != "" {
		for _, tag := range strings.Split(t, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	if m.mode == formCreate {
		return m.createSecret(name, vals, tags)
	}
	return m.updateSecret(name, vals, tags)
}

func (m secretFormModel) createSecret(name string, vals map[string]string, tags []string) (secretFormModel, tea.Cmd) {
	var s secret.Secret
	switch m.secType {
	case secret.TypePassword:
		s = secret.NewPassword(name, vals["url"], vals["username"], vals["password"])
		if v := vals["totp_secret"]; v != "" {
			s.Fields["totp_secret"] = v
		}
		if v := vals["notes"]; v != "" {
			s.Fields["notes"] = v
		}
	case secret.TypeAPIKey:
		s = secret.NewAPIKey(name, vals["service"], vals["key"])
		if v := vals["notes"]; v != "" {
			s.Fields["notes"] = v
		}
	case secret.TypeSSHKey:
		s = secret.NewSSHKey(name, vals["label"], vals["private_key"], vals["public_key"])
		if v := vals["passphrase"]; v != "" {
			s.Fields["passphrase"] = v
		}
		if v := vals["notes"]; v != "" {
			s.Fields["notes"] = v
		}
	case secret.TypeNote:
		s = secret.NewNote(name, vals["content"])
	}
	s.Tags = tags

	if m.vault != nil {
		if err := m.vault.Secrets().Add(s); err != nil {
			m.err = err.Error()
			return m, nil
		}
	}

	return m, func() tea.Msg {
		return navigateMsg{view: viewSecretList}
	}
}

func (m secretFormModel) updateSecret(name string, vals map[string]string, tags []string) (secretFormModel, tea.Cmd) {
	if m.vault == nil {
		return m, func() tea.Msg { return navigateMsg{view: viewSecretList} }
	}

	s, err := m.vault.Secrets().Get(m.editID)
	if err != nil {
		m.err = err.Error()
		return m, nil
	}

	s.Name = name
	s.Tags = tags

	// update fields based on type
	switch s.Type {
	case secret.TypePassword:
		s.Fields["url"] = vals["url"]
		s.Fields["username"] = vals["username"]
		s.Fields["password"] = vals["password"]
		s.Fields["totp_secret"] = vals["totp_secret"]
		s.Fields["notes"] = vals["notes"]
	case secret.TypeAPIKey:
		s.Fields["service"] = vals["service"]
		s.Fields["key"] = vals["key"]
		s.Fields["notes"] = vals["notes"]
	case secret.TypeSSHKey:
		s.Fields["label"] = vals["label"]
		s.Fields["private_key"] = vals["private_key"]
		s.Fields["public_key"] = vals["public_key"]
		s.Fields["passphrase"] = vals["passphrase"]
		s.Fields["notes"] = vals["notes"]
	case secret.TypeNote:
		s.Fields["content"] = vals["content"]
	}

	if err := m.vault.Secrets().Update(s); err != nil {
		m.err = err.Error()
		return m, nil
	}

	return m, func() tea.Msg {
		return navigateMsg{view: viewSecretList}
	}
}

func (m secretFormModel) View() string {
	var b strings.Builder

	b.WriteString("\n")

	// type selector (create only)
	if m.mode == formCreate {
		cursor := "  "
		if m.focused == -1 {
			cursor = lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent).Render("▸ ")
		}
		typeLbl := lipgloss.NewStyle().Foreground(zstyle.Subtext1).Render("type")
		b.WriteString(fmt.Sprintf("  %s%s  ", cursor, typeLbl))
		for i, t := range typeOptions {
			label := typeLabel(t)
			if i == m.typeSel {
				style := lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent).Bold(true)
				b.WriteString(style.Render(label))
			} else {
				style := lipgloss.NewStyle().Foreground(zstyle.Overlay1)
				b.WriteString(style.Render(label))
			}
			if i < len(typeOptions)-1 {
				b.WriteString(lipgloss.NewStyle().Foreground(zstyle.Surface2).Render(" | "))
			}
		}
		if m.focused == -1 {
			hint := zstyle.MutedText.Render("  (←/→ to change)")
			b.WriteString(hint)
		}
		b.WriteString("\n\n")
	}

	// form fields
	labelStyle := lipgloss.NewStyle().Foreground(zstyle.Subtext1)
	cursorActive := lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent).Render("▸ ")
	cursorInactive := "  "

	for i, inp := range m.inputs {
		cursor := cursorInactive
		if i == m.focused {
			cursor = cursorActive
		}
		label := labelStyle.Render(inp.label)
		b.WriteString(fmt.Sprintf("  %s%s\n", cursor, label))
		b.WriteString(fmt.Sprintf("    %s\n", inp.input.View()))
		if i < len(m.inputs)-1 {
			b.WriteString("\n")
		}
	}

	// confirm discard
	if m.confirmDiscard {
		warn := lipgloss.NewStyle().Foreground(zstyle.Warning)
		b.WriteString("\n")
		b.WriteString(warn.Render("  Discard changes? (y/n)"))
		b.WriteString("\n")
	}

	// error
	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(zstyle.StatusErr.Render("  " + m.err))
		b.WriteString("\n")
	}

	return b.String()
}
