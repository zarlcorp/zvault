package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zarlcorp/zvault/internal/dates"
	"github.com/zarlcorp/zvault/internal/task"
)

func TestTaskFormCreateMode(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	view := m.View()
	if !strings.Contains(view, "new task") {
		t.Error("create mode should show 'new task'")
	}
	if m.editing != nil {
		t.Error("create mode should have nil editing")
	}
}

func TestTaskFormEditMode(t *testing.T) {
	v := openTestVault(t)
	tk := addTask(t, v, "Edit this", withPriority(task.PriorityHigh), withTags("work"))

	m := newTaskFormModel(v)
	m.loadTask(tk)

	view := m.View()
	if !strings.Contains(view, "edit task") {
		t.Error("edit mode should show 'edit task'")
	}
	if m.editing == nil {
		t.Error("edit mode should have non-nil editing")
	}
	if m.title.Value() != "Edit this" {
		t.Errorf("title input = %q, want 'Edit this'", m.title.Value())
	}
	if m.priority != task.PriorityHigh {
		t.Errorf("priority = %q, want high", m.priority)
	}
}

func TestTaskFormNavigateMsg(t *testing.T) {
	v := openTestVault(t)
	tk := addTask(t, v, "Nav edit")

	m := newTaskFormModel(v)

	// navigate with task data (edit mode)
	m, _ = m.Update(navigateMsg{view: viewTaskForm, data: tk})
	if m.editing == nil {
		t.Error("should be in edit mode")
	}

	// navigate with nil data (create mode)
	m, _ = m.Update(navigateMsg{view: viewTaskForm, data: nil})
	if m.editing != nil {
		t.Error("should be in create mode")
	}
}

func TestTaskFormEscCancels(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should produce navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewTaskList {
		t.Fatalf("nav = %d, want viewTaskList", nav.view)
	}
}

func TestTaskFormTabNavigation(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	if m.focused != fieldTitle {
		t.Fatalf("initial focus = %d, want fieldTitle", m.focused)
	}

	// tab to priority
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focused != fieldPriority {
		t.Fatalf("after tab: focus = %d, want fieldPriority", m.focused)
	}

	// tab to due
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focused != fieldDue {
		t.Fatalf("after 2 tabs: focus = %d, want fieldDue", m.focused)
	}

	// tab to tags
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focused != fieldTags {
		t.Fatalf("after 3 tabs: focus = %d, want fieldTags", m.focused)
	}

	// tab wraps to title
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focused != fieldTitle {
		t.Fatalf("after 4 tabs: focus = %d, want fieldTitle (wrap)", m.focused)
	}
}

func TestTaskFormShiftTabNavigation(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	// shift-tab from title should wrap to tags
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.focused != fieldTags {
		t.Fatalf("shift-tab from title: focus = %d, want fieldTags", m.focused)
	}
}

func TestTaskFormPriorityCycling(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	if m.priority != task.PriorityNone {
		t.Fatalf("initial priority = %q, want none", m.priority)
	}

	// move to priority field
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	// cycle through priorities with enter
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.priority != task.PriorityLow {
		t.Fatalf("after 1 cycle: priority = %q, want low", m.priority)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.priority != task.PriorityMedium {
		t.Fatalf("after 2 cycles: priority = %q, want medium", m.priority)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.priority != task.PriorityHigh {
		t.Fatalf("after 3 cycles: priority = %q, want high", m.priority)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.priority != task.PriorityNone {
		t.Fatalf("after 4 cycles: priority = %q, want none (wrap)", m.priority)
	}
}

func TestTaskFormPriorityCycleWithSpace(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	// move to priority field
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	// cycle with space
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if m.priority != task.PriorityLow {
		t.Fatalf("after space: priority = %q, want low", m.priority)
	}
}

func TestTaskFormSaveNewTask(t *testing.T) {
	orig := dates.NowFunc
	defer func() { dates.NowFunc = orig }()
	dates.NowFunc = fixedTime(2026, time.February, 18)

	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	// type title
	m.title.SetValue("New task")

	// set priority
	m.priority = task.PriorityMedium

	// set due date
	m.due.SetValue("tomorrow")

	// set tags
	m.tags.SetValue("work, urgent")

	// save with ctrl+s
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatal("ctrl+s should produce navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewTaskList {
		t.Fatalf("nav = %d, want viewTaskList", nav.view)
	}

	// verify task was created
	tasks, err := v.Tasks().List(task.Filter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "New task" {
		t.Errorf("title = %q, want 'New task'", tasks[0].Title)
	}
	if tasks[0].Priority != task.PriorityMedium {
		t.Errorf("priority = %q, want medium", tasks[0].Priority)
	}
	if tasks[0].DueDate == nil {
		t.Error("due date should be set")
	} else {
		expected := time.Date(2026, 2, 19, 0, 0, 0, 0, time.Local)
		if !tasks[0].DueDate.Equal(expected) {
			t.Errorf("due = %v, want %v", tasks[0].DueDate, expected)
		}
	}
	if len(tasks[0].Tags) != 2 || tasks[0].Tags[0] != "work" || tasks[0].Tags[1] != "urgent" {
		t.Errorf("tags = %v, want [work urgent]", tasks[0].Tags)
	}
}

func TestTaskFormSaveEnterOnLast(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	m.title.SetValue("Enter save")

	// navigate to tags (last field)
	m.focused = fieldTags
	m.title.Blur()
	m.tags.Focus()

	// enter on last field should save
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on last field should save")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewTaskList {
		t.Fatalf("nav = %d, want viewTaskList", nav.view)
	}
}

func TestTaskFormSaveRequiresTitle(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	// leave title empty, try to save
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.err != "title is required" {
		t.Fatalf("err = %q, want 'title is required'", m.err)
	}
}

func TestTaskFormSaveInvalidDate(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	m.title.SetValue("Test")
	m.due.SetValue("invalid-date")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.err == "" {
		t.Fatal("should show error for invalid date")
	}
}

func TestTaskFormUpdateExisting(t *testing.T) {
	v := openTestVault(t)
	tk := addTask(t, v, "Original", withPriority(task.PriorityLow))

	m := newTaskFormModel(v)
	m.loadTask(tk)

	m.title.SetValue("Updated")
	m.priority = task.PriorityHigh

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatal("should navigate after save")
	}

	// verify update
	got, err := v.Tasks().Get(tk.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "Updated" {
		t.Errorf("title = %q, want 'Updated'", got.Title)
	}
	if got.Priority != task.PriorityHigh {
		t.Errorf("priority = %q, want high", got.Priority)
	}
}

func TestTaskFormEnterOnTitle(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	// enter on title should move to next field, not save
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.focused != fieldPriority {
		t.Fatalf("focus = %d, want fieldPriority (enter on title moves forward)", m.focused)
	}
}

func TestTaskFormNoVault(t *testing.T) {
	m := newTaskFormModel(nil)
	m.reset()
	m.title.SetValue("Test")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.err != "vault not available" {
		t.Fatalf("err = %q, want 'vault not available'", m.err)
	}
}

func TestTaskFormLoadWithDueDate(t *testing.T) {
	v := openTestVault(t)
	due := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	tk := addTask(t, v, "With due", withDue(due))

	m := newTaskFormModel(v)
	m.loadTask(tk)

	if m.due.Value() != "2026-03-15" {
		t.Errorf("due input = %q, want '2026-03-15'", m.due.Value())
	}
}

func TestTaskFormEmptyDueDate(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	m.title.SetValue("No due date")
	// leave due empty

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatal("should save successfully")
	}

	tasks, _ := v.Tasks().List(task.Filter{})
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].DueDate != nil {
		t.Error("due date should be nil")
	}
}

func TestTaskFormEmptyTags(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	m.title.SetValue("No tags")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	tasks, _ := v.Tasks().List(task.Filter{})
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if len(tasks[0].Tags) != 0 {
		t.Errorf("tags = %v, want empty", tasks[0].Tags)
	}
}

func TestTaskFormViewShowsPrioritySelector(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	view := m.View()
	if !strings.Contains(view, "priority") {
		t.Error("should show priority label")
	}
	if !strings.Contains(view, "none") {
		t.Error("should show initial priority value")
	}
}

func TestTaskFormClearError(t *testing.T) {
	v := openTestVault(t)
	m := newTaskFormModel(v)
	m.reset()

	// trigger error
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.err == "" {
		t.Fatal("should have error")
	}

	// any key should clear error
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if m.err != "" {
		t.Fatalf("error should be cleared, got %q", m.err)
	}
}
