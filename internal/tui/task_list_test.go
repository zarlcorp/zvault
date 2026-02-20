package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zarlcorp/core/pkg/zfilesystem"
	"github.com/zarlcorp/zvault/internal/dates"
	"github.com/zarlcorp/zvault/internal/task"
	"github.com/zarlcorp/zvault/internal/vault"
)

func openTestVault(t *testing.T) *vault.Vault {
	t.Helper()
	v, err := vault.OpenFS(zfilesystem.NewMemFS(), "test-password")
	if err != nil {
		t.Fatalf("open vault: %v", err)
	}
	t.Cleanup(func() { v.Close() })
	return v
}

func addTask(t *testing.T, v *vault.Vault, title string, opts ...func(*task.Task)) task.Task {
	t.Helper()
	tk, err := task.New(title)
	if err != nil {
		t.Fatal(err)
	}
	for _, opt := range opts {
		opt(&tk)
	}
	if err := v.Tasks().Add(tk); err != nil {
		t.Fatalf("add task: %v", err)
	}
	return tk
}

func withPriority(p task.Priority) func(*task.Task) {
	return func(t *task.Task) { t.Priority = p }
}

func withDue(d time.Time) func(*task.Task) {
	return func(t *task.Task) { t.DueDate = &d }
}

func withDone() func(*task.Task) {
	return func(t *task.Task) {
		t.Done = true
		now := time.Now()
		t.CompletedAt = &now
	}
}

func withTags(tags ...string) func(*task.Task) {
	return func(t *task.Task) { t.Tags = tags }
}

func TestTaskListEmpty(t *testing.T) {
	v := openTestVault(t)
	m := newTaskListModel(v)
	view := m.View()
	if !strings.Contains(view, "no tasks") {
		t.Error("empty task list should show 'no tasks'")
	}
}

func TestTaskListShowsTasks(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "Buy groceries")
	addTask(t, v, "Fix bug")

	m := newTaskListModel(v)
	view := m.View()
	if !strings.Contains(view, "Buy groceries") {
		t.Error("view should show first task")
	}
	if !strings.Contains(view, "Fix bug") {
		t.Error("view should show second task")
	}
}

func TestTaskListPriorityIndicators(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "High task", withPriority(task.PriorityHigh))
	addTask(t, v, "Medium task", withPriority(task.PriorityMedium))
	addTask(t, v, "Low task", withPriority(task.PriorityLow))

	m := newTaskListModel(v)
	view := m.View()
	// priority indicators should be present (with ANSI codes)
	if !strings.Contains(view, "High task") {
		t.Error("should show high priority task")
	}
	if !strings.Contains(view, "Medium task") {
		t.Error("should show medium priority task")
	}
}

func TestTaskListCheckbox(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "Pending task")
	addTask(t, v, "Done task", withDone())

	m := newTaskListModel(v)
	view := m.View()
	if !strings.Contains(view, "[ ]") {
		t.Error("pending task should show unchecked checkbox")
	}
	if !strings.Contains(view, "[x]") {
		t.Error("done task should show checked checkbox")
	}
}

func TestTaskListDueDate(t *testing.T) {
	orig := dates.NowFunc
	defer func() { dates.NowFunc = orig }()
	dates.NowFunc = fixedTime(2026, time.February, 18)

	v := openTestVault(t)
	tomorrow := time.Date(2026, 2, 19, 0, 0, 0, 0, time.Local)
	addTask(t, v, "Tomorrow task", withDue(tomorrow))

	m := newTaskListModel(v)
	view := m.View()
	if !strings.Contains(view, "tomorrow") {
		t.Error("should show relative due date")
	}
}

func TestTaskListTags(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "Tagged task", withTags("urgent", "work"))

	m := newTaskListModel(v)
	view := m.View()
	if !strings.Contains(view, "#urgent") {
		t.Error("should show tag")
	}
	if !strings.Contains(view, "#work") {
		t.Error("should show second tag")
	}
}

func TestTaskListSortOrder(t *testing.T) {
	v := openTestVault(t)

	addTask(t, v, "Done task", withDone())
	addTask(t, v, "Low priority", withPriority(task.PriorityLow))
	addTask(t, v, "High priority", withPriority(task.PriorityHigh))
	addTask(t, v, "Medium priority", withPriority(task.PriorityMedium))

	m := newTaskListModel(v)

	// pending tasks should come first, sorted by priority
	if len(m.tasks) != 4 {
		t.Fatalf("expected 4 tasks, got %d", len(m.tasks))
	}

	// first three should be pending, ordered by priority
	if m.tasks[0].Title != "High priority" {
		t.Errorf("first task = %q, want 'High priority'", m.tasks[0].Title)
	}
	if m.tasks[1].Title != "Medium priority" {
		t.Errorf("second task = %q, want 'Medium priority'", m.tasks[1].Title)
	}
	if m.tasks[2].Title != "Low priority" {
		t.Errorf("third task = %q, want 'Low priority'", m.tasks[2].Title)
	}
	// done task should be last
	if m.tasks[3].Title != "Done task" {
		t.Errorf("last task = %q, want 'Done task'", m.tasks[3].Title)
	}
}

func TestTaskListSortByDueDate(t *testing.T) {
	v := openTestVault(t)

	later := time.Date(2026, 3, 15, 0, 0, 0, 0, time.Local)
	sooner := time.Date(2026, 2, 20, 0, 0, 0, 0, time.Local)
	addTask(t, v, "Later", withDue(later))
	addTask(t, v, "Sooner", withDue(sooner))
	addTask(t, v, "No date")

	m := newTaskListModel(v)

	if m.tasks[0].Title != "Sooner" {
		t.Errorf("first = %q, want 'Sooner'", m.tasks[0].Title)
	}
	if m.tasks[1].Title != "Later" {
		t.Errorf("second = %q, want 'Later'", m.tasks[1].Title)
	}
	if m.tasks[2].Title != "No date" {
		t.Errorf("third = %q, want 'No date'", m.tasks[2].Title)
	}
}

func TestTaskListNavigateUpDown(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "Task 1")
	addTask(t, v, "Task 2")
	addTask(t, v, "Task 3")

	m := newTaskListModel(v)
	if m.cursor != 0 {
		t.Fatalf("initial cursor = %d, want 0", m.cursor)
	}

	// move down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 1 {
		t.Fatalf("cursor after down = %d, want 1", m.cursor)
	}

	// move down again
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Fatalf("cursor after 2 down = %d, want 2", m.cursor)
	}

	// can't go past end
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Fatalf("cursor should stay at 2, got %d", m.cursor)
	}

	// move up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 1 {
		t.Fatalf("cursor after up = %d, want 1", m.cursor)
	}
}

func TestTaskListEscGoesBack(t *testing.T) {
	v := openTestVault(t)
	m := newTaskListModel(v)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc should produce navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewMenu {
		t.Fatalf("nav = %d, want viewMenu", nav.view)
	}
}

func TestTaskListEnterOpensDetail(t *testing.T) {
	v := openTestVault(t)
	tk := addTask(t, v, "Detail task")

	m := newTaskListModel(v)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should produce navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewTaskDetail {
		t.Fatalf("nav = %d, want viewTaskDetail", nav.view)
	}
	if nav.data != tk.ID {
		t.Fatalf("nav data = %v, want %q", nav.data, tk.ID)
	}
}

func TestTaskListNewTask(t *testing.T) {
	v := openTestVault(t)
	m := newTaskListModel(v)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd == nil {
		t.Fatal("n should produce navigate command")
	}
	msg := cmd()
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.view != viewTaskForm {
		t.Fatalf("nav = %d, want viewTaskForm", nav.view)
	}
	if nav.data != nil {
		t.Fatalf("nav data = %v, want nil (create mode)", nav.data)
	}
}

func TestTaskListToggleDone(t *testing.T) {
	v := openTestVault(t)
	tk := addTask(t, v, "Toggle me")

	m := newTaskListModel(v)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	// verify task is now done
	got, err := v.Tasks().Get(tk.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !got.Done {
		t.Error("task should be done after toggle")
	}
	if got.CompletedAt == nil {
		t.Error("task should have CompletedAt set")
	}

	// toggle back
	// after reload, the done task may have moved - find it
	for i, tt := range m.tasks {
		if tt.ID == tk.ID {
			m.cursor = i
			break
		}
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})

	got, err = v.Tasks().Get(tk.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Done {
		t.Error("task should be pending after second toggle")
	}
}

func TestTaskListDeleteConfirmation(t *testing.T) {
	v := openTestVault(t)
	tk := addTask(t, v, "Delete me")

	m := newTaskListModel(v)

	// press d - should show confirmation
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if m.confirm != confirmDelete {
		t.Fatal("should be in delete confirmation")
	}
	view := m.View()
	if !strings.Contains(view, "delete") {
		t.Error("should show delete prompt")
	}

	// press n to cancel
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m.confirm != confirmNone {
		t.Fatal("should cancel confirmation")
	}

	// task should still exist
	_, err := v.Tasks().Get(tk.ID)
	if err != nil {
		t.Fatal("task should still exist after cancel")
	}

	// press d again, then y to confirm
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	_, err = v.Tasks().Get(tk.ID)
	if err == nil {
		t.Fatal("task should be deleted after confirm")
	}
}

func TestTaskListClearDone(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "Pending task")
	addTask(t, v, "Done 1", withDone())
	addTask(t, v, "Done 2", withDone())

	m := newTaskListModel(v)

	// press x - should show confirmation
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if m.confirm != confirmClearDone {
		t.Fatal("should be in clear-done confirmation")
	}
	view := m.View()
	if !strings.Contains(view, "clear 2 done") {
		t.Errorf("should show count in clear prompt, got: %s", view)
	}

	// confirm
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	// only pending task should remain
	all, err := v.Tasks().List(task.Filter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("remaining = %d, want 1", len(all))
	}
	if all[0].Title != "Pending task" {
		t.Errorf("remaining title = %q, want 'Pending task'", all[0].Title)
	}
}

func TestTaskListClearDoneNoOp(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "Pending task")

	m := newTaskListModel(v)

	// press x with no done tasks - should not trigger confirmation
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if m.confirm != confirmNone {
		t.Fatal("should not enter confirmation when no done tasks")
	}
}

func TestTaskListFilterCycling(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "Pending")
	addTask(t, v, "Done", withDone())

	m := newTaskListModel(v)

	// initial: all
	if m.filter != taskFilterAll {
		t.Fatalf("initial filter = %d, want taskFilterAll", m.filter)
	}
	if len(m.tasks) != 2 {
		t.Fatalf("all: len = %d, want 2", len(m.tasks))
	}

	// tab -> pending
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.filter != taskFilterPending {
		t.Fatalf("after tab: filter = %d, want taskFilterPending", m.filter)
	}
	if len(m.tasks) != 1 {
		t.Fatalf("pending: len = %d, want 1", len(m.tasks))
	}

	// tab -> done
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.filter != taskFilterDone {
		t.Fatalf("after 2 tabs: filter = %d, want taskFilterDone", m.filter)
	}
	if len(m.tasks) != 1 {
		t.Fatalf("done: len = %d, want 1", len(m.tasks))
	}

	// tab -> wraps back to all
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.filter != taskFilterAll {
		t.Fatalf("after 3 tabs: filter = %d, want taskFilterAll", m.filter)
	}
}

func TestTaskListFilterLabel(t *testing.T) {
	v := openTestVault(t)
	m := newTaskListModel(v)

	tests := []struct {
		filter filterMode
		want   string
	}{
		{taskFilterAll, "filter: all"},
		{taskFilterPending, "filter: pending"},
		{taskFilterDone, "filter: done"},
	}

	for _, tt := range tests {
		m.filter = tt.filter
		view := m.View()
		if !strings.Contains(view, tt.want) {
			t.Errorf("filter %d: view should contain %q", tt.filter, tt.want)
		}
	}
}

func TestTaskListEnterOnEmptyList(t *testing.T) {
	v := openTestVault(t)
	m := newTaskListModel(v)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter on empty list should not produce command")
	}
}

func TestTaskListCursorClamps(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "Task 1")
	addTask(t, v, "Task 2")
	addTask(t, v, "Task 3")

	m := newTaskListModel(v)
	m.cursor = 2

	// delete last task
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	// cursor should clamp to new last index
	if m.cursor >= len(m.tasks) {
		t.Fatalf("cursor = %d, but only %d tasks remain", m.cursor, len(m.tasks))
	}
}

func TestTaskListFilterCyclingWithTags(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "Work task", withTags("work"))
	addTask(t, v, "Home task", withTags("home"))

	m := newTaskListModel(v)

	// initial: all
	if m.filter != taskFilterAll {
		t.Fatalf("initial filter = %d, want taskFilterAll", m.filter)
	}

	// tab -> pending
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.filter != taskFilterPending {
		t.Fatalf("after 1 tab: filter = %d, want taskFilterPending", m.filter)
	}

	// tab -> done
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.filter != taskFilterDone {
		t.Fatalf("after 2 tabs: filter = %d, want taskFilterDone", m.filter)
	}

	// tab -> by tag (first tag)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.filter != taskFilterByTag {
		t.Fatalf("after 3 tabs: filter = %d, want taskFilterByTag", m.filter)
	}
	if m.tagIndex != 0 {
		t.Fatalf("tagIndex = %d, want 0", m.tagIndex)
	}

	// tab -> by tag (second tag)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.filter != taskFilterByTag {
		t.Fatalf("after 4 tabs: filter = %d, want taskFilterByTag", m.filter)
	}
	if m.tagIndex != 1 {
		t.Fatalf("tagIndex = %d, want 1", m.tagIndex)
	}

	// tab -> wraps back to all (past last tag)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.filter != taskFilterAll {
		t.Fatalf("after 5 tabs: filter = %d, want taskFilterAll", m.filter)
	}
}

func TestTaskListFilterSkipsTagModeNoTags(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "No tags task")

	m := newTaskListModel(v)

	// all -> pending -> done -> all (should skip tag mode)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // pending
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // done
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // should skip to all
	if m.filter != taskFilterAll {
		t.Fatalf("after 3 tabs with no tags: filter = %d, want taskFilterAll", m.filter)
	}
}

func TestTaskListTagCycling(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "A", withTags("alpha"))
	addTask(t, v, "B", withTags("beta"))
	addTask(t, v, "C", withTags("gamma"))

	m := newTaskListModel(v)

	// advance to tag mode: all -> pending -> done -> byTag
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	if m.filter != taskFilterByTag {
		t.Fatalf("filter = %d, want taskFilterByTag", m.filter)
	}

	// tags are sorted: alpha, beta, gamma
	wantTags := []string{"alpha", "beta", "gamma"}
	for i, want := range wantTags {
		if m.tagIndex != i {
			t.Fatalf("tagIndex = %d, want %d", m.tagIndex, i)
		}
		if len(m.tags) <= i || m.tags[m.tagIndex] != want {
			t.Fatalf("tag = %q, want %q", m.tags[m.tagIndex], want)
		}
		if i < len(wantTags)-1 {
			m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
		}
	}

	// one more tab wraps to all
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.filter != taskFilterAll {
		t.Fatalf("after cycling all tags: filter = %d, want taskFilterAll", m.filter)
	}
}

func TestTaskListFilterLabelByTag(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "Tagged", withTags("work", "urgent"))

	m := newTaskListModel(v)

	// advance to tag mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // pending
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // done
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // byTag (first tag)

	// tags sorted: urgent, work
	label := m.filterLabel()
	if label != "filter: #urgent" {
		t.Fatalf("label = %q, want %q", label, "filter: #urgent")
	}

	// cycle to next tag
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	label = m.filterLabel()
	if label != "filter: #work" {
		t.Fatalf("label = %q, want %q", label, "filter: #work")
	}
}

func TestTaskListFilterByTagShowsMatchingTasks(t *testing.T) {
	v := openTestVault(t)
	addTask(t, v, "Work task 1", withTags("work"))
	addTask(t, v, "Home task", withTags("home"))
	addTask(t, v, "Work task 2", withTags("work"))
	addTask(t, v, "Untagged task")

	m := newTaskListModel(v)

	// all tasks visible initially
	if len(m.tasks) != 4 {
		t.Fatalf("all: len = %d, want 4", len(m.tasks))
	}

	// advance to tag mode: all -> pending -> done -> byTag
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	// tags sorted: home, work — first tag is "home"
	if m.filter != taskFilterByTag {
		t.Fatalf("filter = %d, want taskFilterByTag", m.filter)
	}
	if len(m.tags) == 0 {
		t.Fatal("no tags collected")
	}
	if m.tags[m.tagIndex] != "home" {
		t.Fatalf("first tag = %q, want 'home'", m.tags[m.tagIndex])
	}
	if len(m.tasks) != 1 {
		t.Fatalf("home filter: len = %d, want 1", len(m.tasks))
	}
	if m.tasks[0].Title != "Home task" {
		t.Fatalf("home filter: task = %q, want 'Home task'", m.tasks[0].Title)
	}

	// tab to next tag: "work"
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.tags[m.tagIndex] != "work" {
		t.Fatalf("second tag = %q, want 'work'", m.tags[m.tagIndex])
	}
	if len(m.tasks) != 2 {
		t.Fatalf("work filter: len = %d, want 2", len(m.tasks))
	}

	workTitles := make(map[string]bool)
	for _, tk := range m.tasks {
		workTitles[tk.Title] = true
	}
	if !workTitles["Work task 1"] || !workTitles["Work task 2"] {
		t.Fatalf("work filter: tasks = %v, want Work task 1 and Work task 2", m.tasks)
	}
}
