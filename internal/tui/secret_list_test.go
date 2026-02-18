package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zarlcorp/zvault/internal/secret"
)

func TestSecretListViewEmpty(t *testing.T) {
	m := newSecretList()
	view := m.View()
	if !strings.Contains(view, "no secrets found") {
		t.Error("empty list should show 'no secrets found'")
	}
}

func TestSecretListViewShowsFilterTabs(t *testing.T) {
	m := newSecretList()
	view := m.View()
	for _, label := range []string{"All", "Password", "API Key", "SSH Key", "Note"} {
		if !strings.Contains(view, label) {
			t.Errorf("list should show filter tab %q", label)
		}
	}
}

func TestSecretListEscNavigatesBack(t *testing.T) {
	m := newSecretList()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should produce a navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewMenu {
		t.Fatalf("nav view = %d, want viewMenu", nav.view)
	}
}

func TestSecretListCursorNavigation(t *testing.T) {
	m := newSecretList()
	m.secrets = []secret.Secret{
		{ID: "1", Name: "First", Type: secret.TypePassword},
		{ID: "2", Name: "Second", Type: secret.TypeAPIKey},
		{ID: "3", Name: "Third", Type: secret.TypeNote},
	}
	m.cursor = 0

	// move down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}

	// move down again
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Fatalf("cursor = %d, want 2", m.cursor)
	}

	// can't go past end
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Fatalf("cursor = %d, want 2 (bounded)", m.cursor)
	}

	// move up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}
}

func TestSecretListEnterNavigatesToDetail(t *testing.T) {
	m := newSecretList()
	m.secrets = []secret.Secret{
		{ID: "abc123", Name: "Test", Type: secret.TypePassword},
	}
	m.cursor = 0

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should produce a navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewSecretDetail {
		t.Fatalf("nav view = %d, want viewSecretDetail", nav.view)
	}
	if nav.data != "abc123" {
		t.Fatalf("nav data = %v, want 'abc123'", nav.data)
	}
}

func TestSecretListSearchMode(t *testing.T) {
	m := newSecretList()
	if m.searching {
		t.Fatal("should not start in search mode")
	}

	// enter search mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !m.searching {
		t.Fatal("should be in search mode after /")
	}

	// esc exits search mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if m.searching {
		t.Fatal("should exit search mode on esc")
	}
}

func TestSecretListFilterCycle(t *testing.T) {
	m := newSecretList()
	if m.filter != filterAll {
		t.Fatalf("initial filter = %d, want filterAll", m.filter)
	}

	tabKey := tea.KeyMsg{Type: tea.KeyTab}

	// tab cycles filter
	m, _ = m.Update(tabKey)
	if m.filter != filterPassword {
		t.Fatalf("filter = %d, want filterPassword", m.filter)
	}

	m, _ = m.Update(tabKey)
	if m.filter != filterAPIKey {
		t.Fatalf("filter = %d, want filterAPIKey", m.filter)
	}

	m, _ = m.Update(tabKey)
	if m.filter != filterSSHKey {
		t.Fatalf("filter = %d, want filterSSHKey", m.filter)
	}

	m, _ = m.Update(tabKey)
	if m.filter != filterNote {
		t.Fatalf("filter = %d, want filterNote", m.filter)
	}

	// wraps back to all
	m, _ = m.Update(tabKey)
	if m.filter != filterAll {
		t.Fatalf("filter = %d, want filterAll (wrapped)", m.filter)
	}
}

func TestSecretListNewNavigatesToForm(t *testing.T) {
	m := newSecretList()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd == nil {
		t.Fatal("n should produce a navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewSecretForm {
		t.Fatalf("nav view = %d, want viewSecretForm", nav.view)
	}
	if nav.data != nil {
		t.Fatalf("nav data should be nil for new secret, got %v", nav.data)
	}
}

func TestSecretListDeleteConfirmation(t *testing.T) {
	m := newSecretList()
	m.secrets = []secret.Secret{
		{ID: "1", Name: "MySecret", Type: secret.TypePassword},
	}
	m.cursor = 0

	// press d to trigger delete confirm
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if !m.confirmDelete {
		t.Fatal("d should trigger delete confirmation")
	}

	// pressing n cancels
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m.confirmDelete {
		t.Fatal("n should cancel delete confirmation")
	}
}

func TestSecretListDeleteConfirmView(t *testing.T) {
	m := newSecretList()
	m.secrets = []secret.Secret{
		{ID: "1", Name: "MySecret", Type: secret.TypePassword},
	}
	m.cursor = 0
	m.confirmDelete = true

	view := m.View()
	if !strings.Contains(view, "Delete 'MySecret'? (y/n)") {
		t.Error("delete confirmation should show secret name")
	}
}

func TestSecretListViewShowsSecrets(t *testing.T) {
	m := newSecretList()
	m.secrets = []secret.Secret{
		{ID: "1", Name: "GitHub", Type: secret.TypePassword, Tags: []string{"dev"}},
		{ID: "2", Name: "AWS Key", Type: secret.TypeAPIKey},
	}
	m.height = 30 // enough height for all items

	view := m.View()
	if !strings.Contains(view, "GitHub") {
		t.Error("list should show secret name 'GitHub'")
	}
	if !strings.Contains(view, "AWS Key") {
		t.Error("list should show secret name 'AWS Key'")
	}
	if !strings.Contains(view, "[pw]") {
		t.Error("list should show password badge")
	}
	if !strings.Contains(view, "[api]") {
		t.Error("list should show api badge")
	}
	if !strings.Contains(view, "dev") {
		t.Error("list should show tags")
	}
}

func TestTypeBadge(t *testing.T) {
	tests := []struct {
		typ  secret.Type
		want string
	}{
		{secret.TypePassword, "pw"},
		{secret.TypeAPIKey, "api"},
		{secret.TypeSSHKey, "ssh"},
		{secret.TypeNote, "note"},
	}
	for _, tt := range tests {
		badge := typeBadge(tt.typ)
		if !strings.Contains(badge, tt.want) {
			t.Errorf("typeBadge(%s) = %q, should contain %q", tt.typ, badge, tt.want)
		}
	}
}
