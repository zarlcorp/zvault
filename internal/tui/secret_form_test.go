package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zarlcorp/zvault/internal/secret"
)

func TestSecretFormInitForCreate(t *testing.T) {
	m := newSecretForm()
	m = m.initForCreate()

	if m.mode != formCreate {
		t.Fatalf("mode = %d, want formCreate", m.mode)
	}
	if m.secType != secret.TypePassword {
		t.Fatalf("secType = %q, want TypePassword", m.secType)
	}
	if len(m.inputs) == 0 {
		t.Fatal("should have form inputs")
	}
	// first input should be name
	if m.inputs[0].fieldKey != "name" {
		t.Fatalf("first input = %q, want 'name'", m.inputs[0].fieldKey)
	}
	// last input should be tags
	last := m.inputs[len(m.inputs)-1]
	if last.fieldKey != "tags" {
		t.Fatalf("last input = %q, want 'tags'", last.fieldKey)
	}
}

func TestSecretFormPasswordFields(t *testing.T) {
	m := newSecretForm()
	m = m.initForCreate()
	m.secType = secret.TypePassword
	m.inputs = buildFormInputs(secret.TypePassword, secret.Secret{})

	keys := make(map[string]bool)
	for _, inp := range m.inputs {
		keys[inp.fieldKey] = true
	}
	for _, want := range []string{"name", "url", "username", "password", "totp_secret", "notes", "tags"} {
		if !keys[want] {
			t.Errorf("missing field %q for password type", want)
		}
	}
}

func TestSecretFormAPIKeyFields(t *testing.T) {
	inputs := buildFormInputs(secret.TypeAPIKey, secret.Secret{})
	keys := make(map[string]bool)
	for _, inp := range inputs {
		keys[inp.fieldKey] = true
	}
	for _, want := range []string{"name", "service", "key", "notes", "tags"} {
		if !keys[want] {
			t.Errorf("missing field %q for apikey type", want)
		}
	}
}

func TestSecretFormSSHKeyFields(t *testing.T) {
	inputs := buildFormInputs(secret.TypeSSHKey, secret.Secret{})
	keys := make(map[string]bool)
	for _, inp := range inputs {
		keys[inp.fieldKey] = true
	}
	for _, want := range []string{"name", "label", "private_key", "public_key", "passphrase", "notes", "tags"} {
		if !keys[want] {
			t.Errorf("missing field %q for sshkey type", want)
		}
	}
}

func TestSecretFormNoteFields(t *testing.T) {
	inputs := buildFormInputs(secret.TypeNote, secret.Secret{})
	keys := make(map[string]bool)
	for _, inp := range inputs {
		keys[inp.fieldKey] = true
	}
	for _, want := range []string{"name", "content", "tags"} {
		if !keys[want] {
			t.Errorf("missing field %q for note type", want)
		}
	}
}

func TestSecretFormMaskedFields(t *testing.T) {
	inputs := buildFormInputs(secret.TypePassword, secret.Secret{})
	masked := make(map[string]bool)
	for _, inp := range inputs {
		if inp.masked {
			masked[inp.fieldKey] = true
		}
	}
	if !masked["password"] {
		t.Error("password field should be masked")
	}
	if !masked["totp_secret"] {
		t.Error("totp_secret field should be masked")
	}
	if masked["name"] {
		t.Error("name field should not be masked")
	}
}

func TestSecretFormTabNavigation(t *testing.T) {
	m := newSecretForm()
	m = m.initForCreate()

	// initial focus is on type selector (-1)
	if m.focused != -1 {
		t.Fatalf("focused = %d, want -1", m.focused)
	}

	// tab from type selector (-1) moves to first input (0)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focused != 0 {
		t.Fatalf("focused = %d, want 0 after tab from type selector", m.focused)
	}

	// tab from 0 moves to 1
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focused != 1 {
		t.Fatalf("focused = %d, want 1 after tab", m.focused)
	}

	// shift+tab from 1 moves back to 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.focused != 0 {
		t.Fatalf("focused = %d, want 0 after shift+tab", m.focused)
	}

	// shift+tab from 0 goes back to type selector (-1)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.focused != -1 {
		t.Fatalf("focused = %d, want -1 after shift+tab from 0", m.focused)
	}

	// shift+tab at -1 stays at -1 (bounded)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.focused != -1 {
		t.Fatalf("focused = %d, want -1 (bounded)", m.focused)
	}
}

func TestSecretFormEscNoChanges(t *testing.T) {
	m := newSecretForm()
	m = m.initForCreate()
	m.changed = false

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc without changes should navigate back immediately")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewSecretList {
		t.Fatalf("nav view = %d, want viewSecretList", nav.view)
	}
}

func TestSecretFormEscWithChangesConfirm(t *testing.T) {
	m := newSecretForm()
	m = m.initForCreate()
	m.changed = true

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd != nil {
		t.Fatal("esc with changes should show confirmation, not navigate")
	}
	if !m.confirmDiscard {
		t.Fatal("should show discard confirmation")
	}

	// confirm with y
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("y should navigate back")
	}
}

func TestSecretFormEscWithChangesDeny(t *testing.T) {
	m := newSecretForm()
	m = m.initForCreate()
	m.changed = true

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if !m.confirmDiscard {
		t.Fatal("should show discard confirmation")
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m.confirmDiscard {
		t.Fatal("n should dismiss discard confirmation")
	}
}

func TestSecretFormSaveEmptyName(t *testing.T) {
	m := newSecretForm()
	m = m.initForCreate()
	// don't set name, just try ctrl+s
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.err != "name is required" {
		t.Fatalf("err = %q, want 'name is required'", m.err)
	}
}

func TestSecretFormViewShowsTypeSelector(t *testing.T) {
	m := newSecretForm()
	m = m.initForCreate()
	view := m.View()
	if !strings.Contains(view, "password") {
		t.Error("create form should show type selector with password")
	}
	if !strings.Contains(view, "api key") {
		t.Error("create form should show type selector with api key")
	}
}

func TestSecretFormViewShowsFields(t *testing.T) {
	m := newSecretForm()
	m = m.initForCreate()
	view := m.View()
	if !strings.Contains(view, "name") {
		t.Error("form should show name field")
	}
	if !strings.Contains(view, "url") {
		t.Error("password form should show url field")
	}
	if !strings.Contains(view, "username") {
		t.Error("password form should show username field")
	}
}

func TestSecretFormConfirmDiscardView(t *testing.T) {
	m := newSecretForm()
	m = m.initForCreate()
	m.confirmDiscard = true

	view := m.View()
	if !strings.Contains(view, "Discard changes? (y/n)") {
		t.Error("should show discard confirmation")
	}
}

func TestSecretFormErrorView(t *testing.T) {
	m := newSecretForm()
	m = m.initForCreate()
	m.err = "test error"

	view := m.View()
	if !strings.Contains(view, "test error") {
		t.Error("should show error message")
	}
}

func TestSecretFormNavigateInitCreate(t *testing.T) {
	m := newSecretForm()
	m, _ = m.Update(navigateMsg{view: viewSecretForm, data: nil})
	if m.mode != formCreate {
		t.Fatalf("mode = %d, want formCreate", m.mode)
	}
}

func TestSecretFormNavigateInitEdit(t *testing.T) {
	m := newSecretForm()
	m, _ = m.Update(navigateMsg{view: viewSecretForm, data: "abc123"})
	if m.mode != formEdit {
		t.Fatalf("mode = %d, want formEdit", m.mode)
	}
	if m.editID != "abc123" {
		t.Fatalf("editID = %q, want 'abc123'", m.editID)
	}
}

func TestSecretFormEnterOnLastFieldSaves(t *testing.T) {
	m := newSecretForm()
	m = m.initForCreate()
	// move to last field
	m.focused = len(m.inputs) - 1
	m.focusCurrent()
	// set a name so save doesn't fail on validation
	m.inputs[0].input.SetValue("Test Secret")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on last field should save (produce navigate command)")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewSecretList {
		t.Fatalf("nav view = %d, want viewSecretList after save", nav.view)
	}
}

func TestSecretFormEnterOnNonLastFieldAdvances(t *testing.T) {
	m := newSecretForm()
	m = m.initForCreate()
	m.focused = 0

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.focused != 1 {
		t.Fatalf("focused = %d, want 1 after enter on non-last field", m.focused)
	}
}

func TestTypeLabel(t *testing.T) {
	tests := []struct {
		typ  secret.Type
		want string
	}{
		{secret.TypePassword, "password"},
		{secret.TypeAPIKey, "api key"},
		{secret.TypeSSHKey, "ssh key"},
		{secret.TypeNote, "note"},
	}
	for _, tt := range tests {
		got := typeLabel(tt.typ)
		if got != tt.want {
			t.Errorf("typeLabel(%s) = %q, want %q", tt.typ, got, tt.want)
		}
	}
}
