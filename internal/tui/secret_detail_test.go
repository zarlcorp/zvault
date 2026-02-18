package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zarlcorp/zvault/internal/secret"
)

func TestSecretDetailViewNoSecret(t *testing.T) {
	m := newSecretDetail()
	view := m.View()
	if !strings.Contains(view, "no secret selected") {
		t.Error("empty detail should show 'no secret selected'")
	}
}

func TestSecretDetailEscNavigatesBack(t *testing.T) {
	m := newSecretDetail()
	m.secretID = "test"
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should produce a navigate command")
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

func TestSecretDetailToggleSensitive(t *testing.T) {
	m := newSecretDetail()
	m.secretID = "test"
	m.secret = secret.NewPassword("test", "http://example.com", "user", "pass123")
	m.fields = buildDetailFields(m.secret)

	if m.showSensitive {
		t.Fatal("should start with sensitive fields hidden")
	}

	// toggle with s
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if !m.showSensitive {
		t.Fatal("s should toggle sensitive fields to visible")
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if m.showSensitive {
		t.Fatal("s should toggle sensitive fields back to hidden")
	}
}

func TestSecretDetailSensitiveMasked(t *testing.T) {
	m := newSecretDetail()
	m.secretID = "test"
	m.secret = secret.NewPassword("test", "http://example.com", "user", "secretpass")
	m.fields = buildDetailFields(m.secret)
	m.showSensitive = false

	view := m.View()
	if strings.Contains(view, "secretpass") {
		t.Error("password should be masked when showSensitive is false")
	}
	if !strings.Contains(view, "••••••••") {
		t.Error("masked value should show dots")
	}
}

func TestSecretDetailSensitiveRevealed(t *testing.T) {
	m := newSecretDetail()
	m.secretID = "test"
	m.secret = secret.NewPassword("test", "http://example.com", "user", "secretpass")
	m.fields = buildDetailFields(m.secret)
	m.showSensitive = true

	view := m.View()
	if !strings.Contains(view, "secretpass") {
		t.Error("password should be visible when showSensitive is true")
	}
}

func TestSecretDetailCursorNavigation(t *testing.T) {
	m := newSecretDetail()
	m.secretID = "test"
	m.secret = secret.NewPassword("test", "http://example.com", "user", "pass")
	m.fields = buildDetailFields(m.secret)
	m.cursor = 0

	// move down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}

	// move up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.cursor)
	}

	// can't go below 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0 (bounded)", m.cursor)
	}
}

func TestSecretDetailDeleteConfirmation(t *testing.T) {
	m := newSecretDetail()
	m.secretID = "test"
	m.secret = secret.NewPassword("MyPass", "", "", "")
	m.fields = buildDetailFields(m.secret)

	// d triggers confirm
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if !m.confirmDelete {
		t.Fatal("d should trigger delete confirmation")
	}

	// n cancels
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m.confirmDelete {
		t.Fatal("n should cancel delete confirmation")
	}
}

func TestSecretDetailDeleteConfirmView(t *testing.T) {
	m := newSecretDetail()
	m.secretID = "test"
	m.secret = secret.NewPassword("MyPass", "", "", "")
	m.fields = buildDetailFields(m.secret)
	m.confirmDelete = true

	view := m.View()
	if !strings.Contains(view, "delete 'MyPass'? (y/n)") {
		t.Error("delete confirmation should show secret name")
	}
}

func TestSecretDetailEditNavigatesToForm(t *testing.T) {
	m := newSecretDetail()
	m.secretID = "abc123"
	m.fields = buildDetailFields(secret.NewPassword("test", "", "", ""))

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if cmd == nil {
		t.Fatal("e should produce a navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewSecretForm {
		t.Fatalf("nav view = %d, want viewSecretForm", nav.view)
	}
	if nav.data != "abc123" {
		t.Fatalf("nav data = %v, want 'abc123'", nav.data)
	}
}

func TestSecretDetailClipboardMsg(t *testing.T) {
	m := newSecretDetail()
	m.secretID = "test"

	m, _ = m.Update(clipboardCopiedMsg{field: "password"})
	if !strings.Contains(m.clipboardMsg, "copied password") {
		t.Fatalf("clipboardMsg = %q, want 'copied password...'", m.clipboardMsg)
	}

	m, _ = m.Update(clipboardClearedMsg{})
	if m.clipboardMsg != "" {
		t.Fatalf("clipboardMsg should be cleared, got %q", m.clipboardMsg)
	}
}

func TestBuildDetailFieldsPassword(t *testing.T) {
	s := secret.NewPassword("Test", "http://example.com", "user", "pass123")
	s.Fields["totp_secret"] = "JBSWY3DPEHPK3PXP"
	s.Fields["notes"] = "my notes"
	s.Tags = []string{"web", "dev"}
	fields := buildDetailFields(s)

	labels := make(map[string]bool)
	for _, f := range fields {
		labels[f.label] = true
	}
	for _, want := range []string{"name", "type", "url", "username", "password", "totp secret", "notes", "tags", "created", "updated"} {
		if !labels[want] {
			t.Errorf("missing field %q", want)
		}
	}

	// check sensitive flags
	for _, f := range fields {
		switch f.label {
		case "password", "totp secret":
			if !f.sensitive {
				t.Errorf("field %q should be sensitive", f.label)
			}
		case "name", "url", "username":
			if f.sensitive {
				t.Errorf("field %q should not be sensitive", f.label)
			}
		}
	}
}

func TestBuildDetailFieldsAPIKey(t *testing.T) {
	s := secret.NewAPIKey("AWS", "aws", "AKIA1234")
	fields := buildDetailFields(s)

	var foundKey bool
	for _, f := range fields {
		if f.label == "key" {
			foundKey = true
			if !f.sensitive {
				t.Error("Key field should be sensitive")
			}
		}
	}
	if !foundKey {
		t.Error("api key should have key field")
	}
}

func TestBuildDetailFieldsSSHKey(t *testing.T) {
	s := secret.NewSSHKey("Server", "prod", "-----BEGIN-----", "ssh-rsa AAAA")
	s.Fields["passphrase"] = "secret"
	fields := buildDetailFields(s)

	sensitiveLabels := make(map[string]bool)
	for _, f := range fields {
		if f.sensitive {
			sensitiveLabels[f.label] = true
		}
	}
	if !sensitiveLabels["private key"] {
		t.Error("private key should be sensitive")
	}
	if !sensitiveLabels["passphrase"] {
		t.Error("passphrase should be sensitive")
	}
}

func TestBuildDetailFieldsNote(t *testing.T) {
	s := secret.NewNote("Ideas", "some content")
	fields := buildDetailFields(s)

	var foundContent bool
	for _, f := range fields {
		if f.label == "content" {
			foundContent = true
			if f.value != "some content" {
				t.Errorf("Content value = %q, want 'some content'", f.value)
			}
		}
	}
	if !foundContent {
		t.Error("note should have content field")
	}
}

func TestBuildDetailFieldsMetadata(t *testing.T) {
	s := secret.NewPassword("test", "", "", "")
	s.CreatedAt = time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	s.UpdatedAt = time.Date(2026, 2, 18, 14, 0, 0, 0, time.UTC)
	fields := buildDetailFields(s)

	for _, f := range fields {
		if f.label == "created" && !strings.Contains(f.value, "2026-01-15") {
			t.Errorf("Created = %q, want date containing '2026-01-15'", f.value)
		}
		if f.label == "updated" && !strings.Contains(f.value, "2026-02-18") {
			t.Errorf("Updated = %q, want date containing '2026-02-18'", f.value)
		}
	}
}
