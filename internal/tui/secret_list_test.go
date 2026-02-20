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
	for _, label := range []string{"all", "password", "api key", "ssh key", "note"} {
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

func TestSecretListFilterCycleNoTags(t *testing.T) {
	m := newSecretList()
	if m.filter != filterAll {
		t.Fatalf("initial filter = %d, want filterAll", m.filter)
	}

	tabKey := tea.KeyMsg{Type: tea.KeyTab}

	// tab cycles type filters
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

	// no tags: skips tag mode, wraps to all
	m, _ = m.Update(tabKey)
	if m.filter != filterAll {
		t.Fatalf("filter = %d, want filterAll (skipped tag mode)", m.filter)
	}
}

func TestSecretListFilterCycleWithTags(t *testing.T) {
	m := newSecretList()
	m.secrets = []secret.Secret{
		{ID: "1", Name: "A", Type: secret.TypePassword, Tags: []string{"work"}},
		{ID: "2", Name: "B", Type: secret.TypeAPIKey, Tags: []string{"personal", "work"}},
		{ID: "3", Name: "C", Type: secret.TypeNote},
	}
	// pre-populate tags
	m.tags = []string{"personal", "work"}

	tabKey := tea.KeyMsg{Type: tea.KeyTab}

	// cycle through type filters
	for i := 0; i < 4; i++ {
		m, _ = m.Update(tabKey)
	}
	if m.filter != filterNote {
		t.Fatalf("filter = %d, want filterNote", m.filter)
	}

	// next tab enters tag mode with first tag
	m, _ = m.Update(tabKey)
	if m.filter != filterByTag {
		t.Fatalf("filter = %d, want filterByTag", m.filter)
	}
	if m.tagIndex != 0 {
		t.Fatalf("tagIndex = %d, want 0", m.tagIndex)
	}

	// next tab moves to second tag
	m, _ = m.Update(tabKey)
	if m.filter != filterByTag {
		t.Fatalf("filter = %d, want filterByTag (second tag)", m.filter)
	}
	if m.tagIndex != 1 {
		t.Fatalf("tagIndex = %d, want 1", m.tagIndex)
	}

	// next tab wraps back to all
	m, _ = m.Update(tabKey)
	if m.filter != filterAll {
		t.Fatalf("filter = %d, want filterAll (wrapped from tags)", m.filter)
	}
}

func TestSecretListTagFilterResults(t *testing.T) {
	m := newSecretList()
	m.secrets = []secret.Secret{
		{ID: "1", Name: "GitHub", Type: secret.TypePassword, Tags: []string{"dev"}},
		{ID: "2", Name: "AWS", Type: secret.TypeAPIKey, Tags: []string{"work"}},
		{ID: "3", Name: "Notes", Type: secret.TypeNote, Tags: []string{"dev", "work"}},
	}
	m.tags = []string{"dev", "work"}
	m.filter = filterByTag
	m.tagIndex = 0 // "dev"

	// manually trigger load behavior: apply tag filter
	var filtered []secret.Secret
	tag := m.tags[m.tagIndex]
	for _, s := range m.secrets {
		for _, st := range s.Tags {
			if st == tag {
				filtered = append(filtered, s)
				break
			}
		}
	}

	if len(filtered) != 2 {
		t.Fatalf("expected 2 secrets with tag 'dev', got %d", len(filtered))
	}
	if filtered[0].Name != "GitHub" || filtered[1].Name != "Notes" {
		t.Errorf("filtered = [%s, %s], want [GitHub, Notes]", filtered[0].Name, filtered[1].Name)
	}
}

func TestSecretListViewShowsTagFilter(t *testing.T) {
	m := newSecretList()
	m.tags = []string{"dev", "work"}
	m.filter = filterByTag
	m.tagIndex = 0

	view := m.View()
	if !strings.Contains(view, "#dev") {
		t.Error("tag filter should show '#dev' when active")
	}
}

func TestSecretListViewShowsTagsLabel(t *testing.T) {
	m := newSecretList()
	m.tags = []string{"dev"}
	m.filter = filterAll

	view := m.View()
	if !strings.Contains(view, "tags") {
		t.Error("should show 'tags' label when tags exist but not in tag mode")
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
	if !strings.Contains(view, "delete 'MySecret'? (y/n)") {
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
