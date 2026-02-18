package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zarlcorp/core/pkg/zstyle"
	"github.com/zarlcorp/zvault/internal/task"
	"github.com/zarlcorp/zvault/internal/vault"
)

// filterMode cycles between task list filter modes.
type filterMode int

const (
	taskFilterAll filterMode = iota
	taskFilterPending
	taskFilterDone
	taskFilterByTag
	taskFilterCount // sentinel
)

// taskListModel is the task list view.
type taskListModel struct {
	vault    *vault.Vault
	tasks    []task.Task
	cursor   int
	filter   filterMode
	tags     []string    // unique tags across all tasks for tag filtering
	tagIndex int         // which tag is selected when cycling tags
	width    int
	height   int
	confirm  confirmKind // active confirmation prompt
	doneCount int        // cached count of done tasks for clear prompt
}

// confirmKind tracks which confirmation prompt is active.
type confirmKind int

const (
	confirmNone confirmKind = iota
	confirmDelete
	confirmClearDone
)

func newTaskListModel(v *vault.Vault) taskListModel {
	m := taskListModel{vault: v}
	m.loadTasks()
	return m
}

func (m *taskListModel) loadTasks() {
	if m.vault == nil {
		m.tasks = nil
		return
	}

	f := m.taskFilter()
	tasks, err := m.vault.Tasks().List(f)
	if err != nil {
		m.tasks = nil
		return
	}

	sortTasks(tasks)
	m.tasks = tasks

	// count done tasks for clear prompt
	all, err := m.vault.Tasks().List(task.Filter{})
	if err == nil {
		count := 0
		for _, t := range all {
			if t.Done {
				count++
			}
		}
		m.doneCount = count
	}

	// collect unique tags
	m.collectTags()

	// clamp cursor
	if m.cursor >= len(m.tasks) {
		m.cursor = max(0, len(m.tasks)-1)
	}
}

func (m *taskListModel) collectTags() {
	m.tags = nil
	if m.vault == nil {
		return
	}
	seen := make(map[string]bool)
	all, err := m.vault.Tasks().List(task.Filter{})
	if err != nil {
		return
	}
	for _, t := range all {
		for _, tag := range t.Tags {
			if !seen[tag] {
				seen[tag] = true
				m.tags = append(m.tags, tag)
			}
		}
	}
	sort.Strings(m.tags)
}

// advanceFilter moves to the next filter mode. When in tag mode, tab cycles
// through tags first; after the last tag it wraps to All. If no tags exist,
// tag mode is skipped entirely.
func (m *taskListModel) advanceFilter() {
	if m.filter == taskFilterByTag {
		// cycle through tags; wrap to All after last
		m.tagIndex++
		if m.tagIndex >= len(m.tags) {
			m.tagIndex = 0
			m.filter = taskFilterAll
		}
		return
	}

	next := (m.filter + 1) % taskFilterCount
	if next == taskFilterByTag && len(m.tags) == 0 {
		next = taskFilterAll
	}
	if next == taskFilterByTag {
		m.tagIndex = 0
	}
	m.filter = next
}

func (m taskListModel) taskFilter() task.Filter {
	switch m.filter {
	case taskFilterPending:
		return task.Filter{Status: task.FilterPending}
	case taskFilterDone:
		return task.Filter{Status: task.FilterDone}
	case taskFilterByTag:
		if len(m.tags) > 0 {
			return task.Filter{Tag: m.tags[m.tagIndex]}
		}
		return task.Filter{}
	default:
		return task.Filter{}
	}
}

// sortTasks sorts: pending first, then by priority (high > medium > low > none), then by due date (soonest first, nil last).
func sortTasks(tasks []task.Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		a, b := tasks[i], tasks[j]

		// pending before done
		if a.Done != b.Done {
			return !a.Done
		}

		// higher priority first
		pa, pb := priorityRank(a.Priority), priorityRank(b.Priority)
		if pa != pb {
			return pa > pb
		}

		// soonest due date first, nil last
		if a.DueDate != nil && b.DueDate != nil {
			if !a.DueDate.Equal(*b.DueDate) {
				return a.DueDate.Before(*b.DueDate)
			}
		}
		if a.DueDate != nil && b.DueDate == nil {
			return true
		}
		if a.DueDate == nil && b.DueDate != nil {
			return false
		}

		return false
	})
}

func priorityRank(p task.Priority) int {
	switch p {
	case task.PriorityHigh:
		return 3
	case task.PriorityMedium:
		return 2
	case task.PriorityLow:
		return 1
	default:
		return 0
	}
}

// tasksRefreshedMsg signals the list should reload.
type tasksRefreshedMsg struct{}

func (m taskListModel) Update(msg tea.Msg) (taskListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tasksRefreshedMsg:
		m.loadTasks()
		return m, nil

	case navigateMsg:
		if msg.view == viewTaskList {
			m.loadTasks()
			m.confirm = confirmNone
		}
		return m, nil

	case tea.KeyMsg:
		// handle confirmation prompts first
		if m.confirm != confirmNone {
			return m.handleConfirm(msg)
		}

		switch {
		case key.Matches(msg, zstyle.KeyBack):
			return m, func() tea.Msg { return navigateMsg{view: viewMenu} }

		case key.Matches(msg, zstyle.KeyUp):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, zstyle.KeyDown):
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}

		case key.Matches(msg, zstyle.KeyEnter):
			if len(m.tasks) > 0 {
				t := m.tasks[m.cursor]
				return m, func() tea.Msg { return navigateMsg{view: viewTaskDetail, data: t.ID} }
			}

		case msg.String() == " ":
			// toggle done
			if len(m.tasks) > 0 {
				return m.toggleDone()
			}

		case msg.String() == "n":
			return m, func() tea.Msg { return navigateMsg{view: viewTaskForm, data: nil} }

		case msg.String() == "d":
			if len(m.tasks) > 0 {
				m.confirm = confirmDelete
			}

		case msg.String() == "x":
			if m.doneCount > 0 {
				m.confirm = confirmClearDone
			}

		case key.Matches(msg, zstyle.KeyTab):
			m.advanceFilter()
			m.loadTasks()
		}
	}
	return m, nil
}

func (m taskListModel) handleConfirm(msg tea.KeyMsg) (taskListModel, tea.Cmd) {
	switch msg.String() {
	case "y":
		switch m.confirm {
		case confirmDelete:
			m.confirm = confirmNone
			return m.deleteSelected()
		case confirmClearDone:
			m.confirm = confirmNone
			return m.clearDone()
		}
	case "n", "esc":
		m.confirm = confirmNone
	}
	return m, nil
}

func (m taskListModel) toggleDone() (taskListModel, tea.Cmd) {
	if m.vault == nil || len(m.tasks) == 0 {
		return m, nil
	}

	t := m.tasks[m.cursor]
	t.Done = !t.Done
	if t.Done {
		now := nowFunc()
		t.CompletedAt = &now
	} else {
		t.CompletedAt = nil
	}

	if err := m.vault.Tasks().Update(t); err != nil {
		return m, func() tea.Msg { return errMsg{err: err} }
	}

	m.loadTasks()
	return m, nil
}

func (m taskListModel) deleteSelected() (taskListModel, tea.Cmd) {
	if m.vault == nil || len(m.tasks) == 0 {
		return m, nil
	}

	t := m.tasks[m.cursor]
	if err := m.vault.Tasks().Delete(t.ID); err != nil {
		return m, func() tea.Msg { return errMsg{err: err} }
	}

	m.loadTasks()
	return m, nil
}

func (m taskListModel) clearDone() (taskListModel, tea.Cmd) {
	if m.vault == nil {
		return m, nil
	}

	_, err := m.vault.Tasks().ClearDone()
	if err != nil {
		return m, func() tea.Msg { return errMsg{err: err} }
	}

	m.loadTasks()
	return m, nil
}

var (
	taskHighStyle   = lipgloss.NewStyle().Foreground(zstyle.Red).Bold(true)
	taskMediumStyle = lipgloss.NewStyle().Foreground(zstyle.Yellow)
	taskDoneStyle   = lipgloss.NewStyle().Foreground(zstyle.Overlay0).Strikethrough(true)
	taskDueDimStyle = lipgloss.NewStyle().Foreground(zstyle.Overlay1)
	taskDueRedStyle = lipgloss.NewStyle().Foreground(zstyle.Red)
	taskTagStyle    = lipgloss.NewStyle().Foreground(zstyle.Sapphire)
	taskCursorStyle = lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent)
)

func (m taskListModel) View() string {
	var b strings.Builder
	b.WriteString("\n")

	// filter indicator
	filterLabel := m.filterLabel()
	b.WriteString(fmt.Sprintf("  %s\n\n", zstyle.MutedText.Render(filterLabel)))

	if len(m.tasks) == 0 {
		b.WriteString(fmt.Sprintf("  %s\n", zstyle.MutedText.Render("no tasks")))
		return b.String()
	}

	// visible tasks (constrain to available height)
	maxVisible := m.height - 10 // header, footer, filter, padding
	if maxVisible < 3 {
		maxVisible = 3
	}

	// calculate scroll window
	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.tasks) {
		end = len(m.tasks)
		start = max(0, end-maxVisible)
	}

	for i := start; i < end; i++ {
		t := m.tasks[i]
		line := m.renderTaskLine(t, i == m.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// show scroll indicator
	if len(m.tasks) > maxVisible {
		b.WriteString(fmt.Sprintf("\n  %s\n", zstyle.MutedText.Render(
			fmt.Sprintf("(%d of %d)", m.cursor+1, len(m.tasks)),
		)))
	}

	// confirmation prompt
	if m.confirm != confirmNone {
		b.WriteString("\n")
		b.WriteString(m.confirmPrompt())
	}

	return b.String()
}

func (m taskListModel) filterLabel() string {
	switch m.filter {
	case taskFilterPending:
		return "filter: pending"
	case taskFilterDone:
		return "filter: done"
	case taskFilterByTag:
		if len(m.tags) > 0 {
			return "filter: #" + m.tags[m.tagIndex]
		}
		return "filter: all"
	default:
		return "filter: all"
	}
}

func (m taskListModel) renderTaskLine(t task.Task, selected bool) string {
	var parts []string

	// cursor
	cursor := "  "
	if selected {
		cursor = taskCursorStyle.Render("▸ ")
	}

	// checkbox
	check := "[ ]"
	if t.Done {
		check = "[x]"
	}

	// priority indicator
	pri := "  "
	switch t.Priority {
	case task.PriorityHigh:
		pri = taskHighStyle.Render("!!")
	case task.PriorityMedium:
		pri = taskMediumStyle.Render(" !")
	}

	// title
	title := t.Title
	if t.Done {
		title = taskDoneStyle.Render(title)
	}

	parts = append(parts, cursor+check, pri, title)

	// due date
	if t.DueDate != nil {
		dueStr := formatDueDate(t.DueDate)
		if t.Done {
			dueStr = taskDueDimStyle.Render("due: " + dueStr)
		} else if isOverdue(t.DueDate) {
			dueStr = taskDueRedStyle.Render("due: " + dueStr)
		} else {
			dueStr = taskDueDimStyle.Render("due: " + dueStr)
		}
		parts = append(parts, " "+dueStr)
	}

	// tags
	if len(t.Tags) > 0 {
		var tagStrs []string
		for _, tag := range t.Tags {
			tagStrs = append(tagStrs, "#"+tag)
		}
		parts = append(parts, " "+taskTagStyle.Render(strings.Join(tagStrs, " ")))
	}

	return "  " + strings.Join(parts, " ")
}

func (m taskListModel) confirmPrompt() string {
	switch m.confirm {
	case confirmDelete:
		if len(m.tasks) > 0 {
			t := m.tasks[m.cursor]
			return fmt.Sprintf("  %s",
				zstyle.StatusWarn.Render(fmt.Sprintf("Delete \"%s\"? (y/n)", t.Title)))
		}
	case confirmClearDone:
		return fmt.Sprintf("  %s",
			zstyle.StatusWarn.Render(fmt.Sprintf("Clear %d done tasks? (y/n)", m.doneCount)))
	}
	return ""
}
