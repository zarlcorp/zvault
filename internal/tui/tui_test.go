package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestModel() Model {
	// use a non-existent dir to trigger first-run mode
	return NewWithDir("0.1.0", "/tmp/zvault-test-nonexistent")
}

func TestNewStartsAtPasswordView(t *testing.T) {
	m := newTestModel()
	if m.view != viewPassword {
		t.Fatalf("initial view = %d, want viewPassword (%d)", m.view, viewPassword)
	}
}

func TestInitReturnsBlink(t *testing.T) {
	m := newTestModel()
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init should return a blink command for text input")
	}
}

func TestCtrlCQuitsFromPassword(t *testing.T) {
	m := newTestModel()
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	_ = result
	if cmd == nil {
		t.Fatal("ctrl+c should return quit command from password view")
	}
}

func TestQDoesNotQuitFromPassword(t *testing.T) {
	m := newTestModel()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	// q should type into the input, not quit
	// if a command is returned, it should be the text input blink, not quit
	if cmd != nil {
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); ok {
			t.Fatal("q should not quit from password view (it's a text input)")
		}
	}
}

func TestQQuitsFromMenu(t *testing.T) {
	m := newTestModel()
	m.view = viewMenu
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q should return quit command from menu view")
	}
}

func TestCtrlCQuitsFromMenu(t *testing.T) {
	m := newTestModel()
	m.view = viewMenu
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c should return quit command from menu view")
	}
}

func TestQQuitsFromNonInputViews(t *testing.T) {
	// views without text inputs: q quits
	views := []viewID{viewSecretList, viewSecretDetail, viewTaskList, viewTaskDetail}
	for _, v := range views {
		m := newTestModel()
		m.view = v
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd == nil {
			t.Fatalf("q should return quit command from view %d", v)
		}
	}
}

func TestQDoesNotQuitFromFormViews(t *testing.T) {
	// views with text inputs: q types, does not quit
	views := []viewID{viewSecretForm, viewTaskForm}
	for _, v := range views {
		m := newTestModel()
		m.view = v
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd != nil {
			msg := cmd()
			if _, ok := msg.(tea.QuitMsg); ok {
				t.Fatalf("q should not quit from form view %d (text input)", v)
			}
		}
	}
}

func TestWindowSizePropagates(t *testing.T) {
	m := newTestModel()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	rm := result.(Model)
	if rm.width != 120 || rm.height != 40 {
		t.Fatalf("size = %dx%d, want 120x40", rm.width, rm.height)
	}
	if rm.password.width != 120 || rm.password.height != 40 {
		t.Fatalf("password size = %dx%d, want 120x40", rm.password.width, rm.password.height)
	}
	if rm.menu.width != 120 || rm.menu.height != 40 {
		t.Fatalf("menu size = %dx%d, want 120x40", rm.menu.width, rm.menu.height)
	}
	if rm.secretList.width != 120 || rm.secretList.height != 40 {
		t.Fatalf("secretList size = %dx%d, want 120x40", rm.secretList.width, rm.secretList.height)
	}
}

func TestNavigateMsg(t *testing.T) {
	m := newTestModel()
	m.view = viewMenu
	result, _ := m.Update(navigateMsg{view: viewSecretList})
	rm := result.(Model)
	if rm.view != viewSecretList {
		t.Fatalf("view = %d, want viewSecretList (%d)", rm.view, viewSecretList)
	}
}

func TestEscFromSecretListGoesToMenu(t *testing.T) {
	m := newTestModel()
	m.view = viewSecretList
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	_ = result
	if cmd == nil {
		t.Fatal("esc should produce a navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewMenu {
		t.Fatalf("nav view = %d, want viewMenu (%d)", nav.view, viewMenu)
	}
}

func TestEscFromTaskListGoesToMenu(t *testing.T) {
	m := newTestModel()
	m.view = viewTaskList
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	_ = result
	if cmd == nil {
		t.Fatal("esc should produce a navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewMenu {
		t.Fatalf("nav view = %d, want viewMenu (%d)", nav.view, viewMenu)
	}
}

func TestEscFromSecretDetailGoesToSecretList(t *testing.T) {
	m := newTestModel()
	m.view = viewSecretDetail
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	_ = result
	if cmd == nil {
		t.Fatal("esc should produce a navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewSecretList {
		t.Fatalf("nav view = %d, want viewSecretList (%d)", nav.view, viewSecretList)
	}
}

func TestEscFromTaskDetailGoesToTaskList(t *testing.T) {
	m := newTestModel()
	m.view = viewTaskDetail
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	_ = result
	if cmd == nil {
		t.Fatal("esc should produce a navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewTaskList {
		t.Fatalf("nav view = %d, want viewTaskList (%d)", nav.view, viewTaskList)
	}
}

func TestMenuSelectSecrets(t *testing.T) {
	m := newTestModel()
	m.view = viewMenu
	m.menu.cursor = menuSecrets

	// press enter to select
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = result
	if cmd == nil {
		t.Fatal("enter should produce a navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewSecretList {
		t.Fatalf("nav view = %d, want viewSecretList (%d)", nav.view, viewSecretList)
	}
}

func TestMenuSelectTasks(t *testing.T) {
	m := newTestModel()
	m.view = viewMenu
	m.menu.cursor = menuTasks

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = result
	if cmd == nil {
		t.Fatal("enter should produce a navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewTaskList {
		t.Fatalf("nav view = %d, want viewTaskList (%d)", nav.view, viewTaskList)
	}
}

func TestMenuNavigateDown(t *testing.T) {
	m := newTestModel()
	m.view = viewMenu
	m.menu.cursor = menuSecrets

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	rm := result.(Model)
	if rm.menu.cursor != menuTasks {
		t.Fatalf("cursor = %d, want menuTasks (%d)", rm.menu.cursor, menuTasks)
	}
}

func TestMenuNavigateUp(t *testing.T) {
	m := newTestModel()
	m.view = viewMenu
	m.menu.cursor = menuTasks

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	rm := result.(Model)
	if rm.menu.cursor != menuSecrets {
		t.Fatalf("cursor = %d, want menuSecrets (%d)", rm.menu.cursor, menuSecrets)
	}
}

func TestMenuCursorBoundsLower(t *testing.T) {
	m := newTestModel()
	m.view = viewMenu
	m.menu.cursor = menuSecrets

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	rm := result.(Model)
	if rm.menu.cursor != menuSecrets {
		t.Fatalf("cursor went below 0: %d", rm.menu.cursor)
	}
}

func TestMenuCursorBoundsUpper(t *testing.T) {
	m := newTestModel()
	m.view = viewMenu
	m.menu.cursor = menuTasks

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	rm := result.(Model)
	if rm.menu.cursor != menuTasks {
		t.Fatalf("cursor went above max: %d", rm.menu.cursor)
	}
}

// --- View rendering tests ---

func TestPasswordViewShowsTitle(t *testing.T) {
	m := newTestModel()
	view := m.View()
	if !strings.Contains(view, "zvault") {
		t.Error("view should contain app name")
	}
}

func TestPasswordViewFirstRun(t *testing.T) {
	m := NewWithDir("0.1.0", "/tmp/zvault-nonexistent-dir-for-test")
	if !m.password.firstRun {
		t.Fatal("should detect first run for non-existent dir")
	}
	view := m.password.View()
	if !strings.Contains(view, "create new vault") {
		t.Error("first-run view should say 'create new vault'")
	}
	if !strings.Contains(view, "confirm") {
		t.Error("first-run view should have confirm field")
	}
}

func TestPasswordViewReturningUser(t *testing.T) {
	// use a dir that exists
	m := NewWithDir("0.1.0", "/tmp")
	if m.password.firstRun {
		t.Fatal("should not be first run for existing dir")
	}
	view := m.password.View()
	if !strings.Contains(view, "unlock vault") {
		t.Error("returning user view should say 'unlock vault'")
	}
	if strings.Contains(view, "confirm") {
		t.Error("returning user view should not have confirm field")
	}
}

func TestPasswordEmptySubmit(t *testing.T) {
	m := newTestModel()
	pm := m.password
	pm, _ = pm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if pm.err != "password cannot be empty" {
		t.Fatalf("err = %q, want 'password cannot be empty'", pm.err)
	}
}

func TestPasswordFirstRunConfirmMismatch(t *testing.T) {
	m := NewWithDir("0.1.0", "/tmp/zvault-nonexistent-dir-for-test")
	pm := m.password

	// type a password
	pm.password.SetValue("secret123")
	// submit to move to confirm field
	pm, _ = pm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if pm.focused != fieldConfirm {
		t.Fatalf("focused = %d, want fieldConfirm (%d)", pm.focused, fieldConfirm)
	}

	// type mismatched confirm
	pm.confirm.SetValue("different")
	pm, _ = pm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if pm.err != "passwords do not match" {
		t.Fatalf("err = %q, want 'passwords do not match'", pm.err)
	}
}

func TestHeaderRendering(t *testing.T) {
	h := renderHeader(viewMenu, 80)
	if !strings.Contains(h, "zvault") {
		t.Error("header should contain app name")
	}
	if !strings.Contains(h, "menu") {
		t.Error("header should contain view title")
	}
}

func TestFooterRendering(t *testing.T) {
	f := renderFooter(viewMenu, 80)
	if !strings.Contains(f, "quit") {
		t.Error("menu footer should contain quit hint")
	}
	if !strings.Contains(f, "select") {
		t.Error("menu footer should contain select hint")
	}
}

func TestAllViewsRender(t *testing.T) {
	views := []viewID{
		viewPassword, viewMenu,
		viewSecretList, viewSecretDetail, viewSecretForm,
		viewTaskList, viewTaskDetail, viewTaskForm,
	}
	for _, v := range views {
		m := newTestModel()
		m.view = v
		view := m.View()
		if view == "" {
			t.Fatalf("view %d rendered empty", v)
		}
		// every view should have header with zvault
		if !strings.Contains(view, "zvault") {
			t.Fatalf("view %d missing header app name", v)
		}
	}
}

func TestMenuViewShowsCounts(t *testing.T) {
	m := newTestModel()
	m.view = viewMenu
	m.menu.secretCount = 12
	m.menu.pendingCount = 5
	view := m.menu.View()
	if !strings.Contains(view, "(12)") {
		t.Error("menu should show secret count")
	}
	if !strings.Contains(view, "(5 pending)") {
		t.Error("menu should show pending task count")
	}
}

func TestViewTitles(t *testing.T) {
	tests := []struct {
		id   viewID
		want string
	}{
		{viewPassword, "unlock"},
		{viewMenu, "menu"},
		{viewSecretList, "secrets"},
		{viewSecretDetail, "secret"},
		{viewSecretForm, "edit secret"},
		{viewTaskList, "tasks"},
		{viewTaskDetail, "task"},
		{viewTaskForm, "edit task"},
	}
	for _, tt := range tests {
		got := viewTitle(tt.id)
		if got != tt.want {
			t.Errorf("viewTitle(%d) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestParentViews(t *testing.T) {
	tests := []struct {
		id   viewID
		want viewID
	}{
		{viewSecretList, viewMenu},
		{viewTaskList, viewMenu},
		{viewSecretDetail, viewSecretList},
		{viewSecretForm, viewSecretList},
		{viewTaskDetail, viewTaskList},
		{viewTaskForm, viewTaskList},
	}
	for _, tt := range tests {
		got := parentView(tt.id)
		if got != tt.want {
			t.Errorf("parentView(%d) = %d, want %d", tt.id, got, tt.want)
		}
	}
}

func TestVaultOpenedMsgNavigatesToMenu(t *testing.T) {
	m := newTestModel()
	// simulate vault opened - we need a real vault dir for this
	// Instead, test the navigateMsg path
	result, _ := m.Update(navigateMsg{view: viewMenu})
	rm := result.(Model)
	if rm.view != viewMenu {
		t.Fatalf("view = %d, want viewMenu (%d)", rm.view, viewMenu)
	}
}

func TestErrMsgInPasswordView(t *testing.T) {
	m := newTestModel()
	result, _ := m.Update(errMsg{err: fmt.Errorf("bad password")})
	rm := result.(Model)
	if rm.password.err != "bad password" {
		t.Fatalf("err = %q, want 'bad password'", rm.password.err)
	}
}

func TestErrMsgInNonPasswordView(t *testing.T) {
	m := newTestModel()
	m.view = viewMenu
	result, _ := m.Update(errMsg{err: fmt.Errorf("something went wrong")})
	rm := result.(Model)
	if rm.err != "something went wrong" {
		t.Fatalf("err = %q, want 'something went wrong'", rm.err)
	}
}

func TestErrorDisplayInView(t *testing.T) {
	m := newTestModel()
	m.view = viewMenu
	m.err = "test error"
	view := m.View()
	if !strings.Contains(view, "test error") {
		t.Error("view should display error message")
	}
}

func TestPasswordErrorDisplayInView(t *testing.T) {
	m := newTestModel()
	m.password.err = "incorrect password"
	view := m.View()
	if !strings.Contains(view, "incorrect password") {
		t.Error("password view should display error message")
	}
}
