package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zarlcorp/zvault/internal/task"
)

func TestTaskDetailNotLoaded(t *testing.T) {
	v := openTestVault(t)
	m := newTaskDetailModel(v)
	view := m.View()
	if !strings.Contains(view, "no task selected") {
		t.Error("unloaded detail should show 'no task selected'")
	}
}

func TestTaskDetailShowsTask(t *testing.T) {
	orig := nowFunc
	defer func() { nowFunc = orig }()
	nowFunc = fixedTime(2026, time.February, 18)

	v := openTestVault(t)
	due := time.Date(2026, 2, 20, 0, 0, 0, 0, time.Local)
	tk := addTask(t, v, "Review PR", withPriority(task.PriorityHigh), withDue(due), withTags("work"))

	m := newTaskDetailModel(v)
	m.loadTask(tk.ID)

	view := m.View()
	if !strings.Contains(view, "Review PR") {
		t.Error("should show title")
	}
	if !strings.Contains(view, "high") {
		t.Error("should show priority")
	}
	if !strings.Contains(view, "in 2 days") {
		t.Error("should show relative due date")
	}
	if !strings.Contains(view, "#work") {
		t.Error("should show tags")
	}
	if !strings.Contains(view, "pending") {
		t.Error("should show status")
	}
}

func TestTaskDetailShowsDoneStatus(t *testing.T) {
	v := openTestVault(t)
	tk := addTask(t, v, "Finished task", withDone())

	m := newTaskDetailModel(v)
	m.loadTask(tk.ID)

	view := m.View()
	if !strings.Contains(view, "done") {
		t.Error("should show done status")
	}
	if !strings.Contains(view, "completed") {
		t.Error("should show completed timestamp")
	}
}

func TestTaskDetailEscGoesBack(t *testing.T) {
	v := openTestVault(t)
	m := newTaskDetailModel(v)
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

func TestTaskDetailEditNavigates(t *testing.T) {
	v := openTestVault(t)
	tk := addTask(t, v, "Edit me")

	m := newTaskDetailModel(v)
	m.loadTask(tk.ID)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if cmd == nil {
		t.Fatal("e should produce navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewTaskForm {
		t.Fatalf("nav = %d, want viewTaskForm", nav.view)
	}
	taskData, ok := nav.data.(task.Task)
	if !ok {
		t.Fatalf("expected task.Task data, got %T", nav.data)
	}
	if taskData.Title != "Edit me" {
		t.Fatalf("task title = %q, want 'Edit me'", taskData.Title)
	}
}

func TestTaskDetailToggleDone(t *testing.T) {
	v := openTestVault(t)
	tk := addTask(t, v, "Toggle me")

	m := newTaskDetailModel(v)
	m.loadTask(tk.ID)

	// toggle to done
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !m.task.Done {
		t.Error("task should be done after toggle")
	}

	// toggle back
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if m.task.Done {
		t.Error("task should be pending after second toggle")
	}
}

func TestTaskDetailDeleteConfirm(t *testing.T) {
	v := openTestVault(t)
	tk := addTask(t, v, "Delete me")

	m := newTaskDetailModel(v)
	m.loadTask(tk.ID)

	// press d
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if m.confirm != confirmDelete {
		t.Fatal("should be in delete confirmation")
	}

	view := m.View()
	if !strings.Contains(view, "Delete") {
		t.Error("should show delete prompt")
	}

	// press y to confirm
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	_ = m

	// should navigate back to list
	if cmd == nil {
		t.Fatal("should navigate after delete")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewTaskList {
		t.Fatalf("nav = %d, want viewTaskList", nav.view)
	}

	// task should be gone
	_, err := v.Tasks().Get(tk.ID)
	if err == nil {
		t.Fatal("task should be deleted")
	}
}

func TestTaskDetailDeleteCancel(t *testing.T) {
	v := openTestVault(t)
	tk := addTask(t, v, "Keep me")

	m := newTaskDetailModel(v)
	m.loadTask(tk.ID)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m.confirm != confirmNone {
		t.Fatal("should cancel confirmation")
	}

	_, err := v.Tasks().Get(tk.ID)
	if err != nil {
		t.Fatal("task should still exist")
	}
}

func TestTaskDetailOverdueDateRed(t *testing.T) {
	orig := nowFunc
	defer func() { nowFunc = orig }()
	nowFunc = fixedTime(2026, time.February, 18)

	v := openTestVault(t)
	past := time.Date(2026, 2, 15, 0, 0, 0, 0, time.Local)
	tk := addTask(t, v, "Overdue task", withDue(past))

	m := newTaskDetailModel(v)
	m.loadTask(tk.ID)

	view := m.View()
	if !strings.Contains(view, "overdue") {
		t.Error("should show overdue text")
	}
}

func TestTaskDetailNavigateMsg(t *testing.T) {
	v := openTestVault(t)
	tk := addTask(t, v, "Nav task")

	m := newTaskDetailModel(v)
	m, _ = m.Update(navigateMsg{view: viewTaskDetail, data: tk.ID})
	if !m.loaded {
		t.Fatal("should be loaded after navigateMsg")
	}
	if m.task.Title != "Nav task" {
		t.Fatalf("title = %q, want 'Nav task'", m.task.Title)
	}
}
