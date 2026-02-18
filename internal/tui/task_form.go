package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zarlcorp/core/pkg/zstyle"
	"github.com/zarlcorp/zvault/internal/task"
	"github.com/zarlcorp/zvault/internal/vault"
)

// formField identifies which form field is focused.
type formField int

const (
	fieldTitle formField = iota
	fieldPriority
	fieldDue
	fieldTags
	formFieldCount // sentinel
)

// taskFormModel handles create and edit forms for tasks.
type taskFormModel struct {
	vault    *vault.Vault
	editing  *task.Task // nil = create mode, non-nil = edit mode
	title    textinput.Model
	due      textinput.Model
	tags     textinput.Model
	priority task.Priority
	focused  formField
	err      string
	width    int
	height   int
}

var priorities = []task.Priority{
	task.PriorityNone,
	task.PriorityLow,
	task.PriorityMedium,
	task.PriorityHigh,
}

func newTaskFormModel(v *vault.Vault) taskFormModel {
	titleInput := textinput.New()
	titleInput.Placeholder = "task title"
	titleInput.Focus()
	titleInput.PromptStyle = lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent)
	titleInput.TextStyle = lipgloss.NewStyle().Foreground(zstyle.Text)

	dueInput := textinput.New()
	dueInput.Placeholder = "YYYY-MM-DD, tomorrow, +3d, next week"
	dueInput.PromptStyle = lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent)
	dueInput.TextStyle = lipgloss.NewStyle().Foreground(zstyle.Text)

	tagsInput := textinput.New()
	tagsInput.Placeholder = "comma-separated tags"
	tagsInput.PromptStyle = lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent)
	tagsInput.TextStyle = lipgloss.NewStyle().Foreground(zstyle.Text)

	return taskFormModel{
		vault:   v,
		title:   titleInput,
		due:     dueInput,
		tags:    tagsInput,
		focused: fieldTitle,
	}
}

func (m *taskFormModel) reset() {
	m.editing = nil
	m.title.SetValue("")
	m.due.SetValue("")
	m.tags.SetValue("")
	m.priority = task.PriorityNone
	m.focused = fieldTitle
	m.err = ""
	m.title.Focus()
	m.due.Blur()
	m.tags.Blur()
}

func (m *taskFormModel) loadTask(t task.Task) {
	m.editing = &t
	m.title.SetValue(t.Title)
	m.priority = t.Priority
	m.due.SetValue(formatDateForEdit(t.DueDate))
	if len(t.Tags) > 0 {
		m.tags.SetValue(strings.Join(t.Tags, ", "))
	} else {
		m.tags.SetValue("")
	}
	m.focused = fieldTitle
	m.err = ""
	m.title.Focus()
	m.due.Blur()
	m.tags.Blur()
}

func (m taskFormModel) Update(msg tea.Msg) (taskFormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case navigateMsg:
		if msg.view == viewTaskForm {
			if t, ok := msg.data.(task.Task); ok {
				m.loadTask(t)
			} else {
				m.reset()
			}
		}
		return m, nil

	case tea.KeyMsg:
		m.err = ""

		switch {
		case key.Matches(msg, zstyle.KeyBack):
			return m, func() tea.Msg { return navigateMsg{view: viewTaskList} }

		case msg.Type == tea.KeyCtrlS:
			return m.save()

		case key.Matches(msg, zstyle.KeyTab):
			return m.nextField(), nil

		case msg.Type == tea.KeyShiftTab:
			return m.prevField(), nil

		case key.Matches(msg, zstyle.KeyEnter):
			// enter on priority field cycles priority
			if m.focused == fieldPriority {
				m.cyclePriority()
				return m, nil
			}
			// enter on last field saves
			if m.focused == fieldTags {
				return m.save()
			}
			// enter on other fields moves to next
			return m.nextField(), nil

		case msg.String() == " " && m.focused == fieldPriority:
			m.cyclePriority()
			return m, nil
		}
	}

	return m.updateInputs(msg)
}

func (m *taskFormModel) cyclePriority() {
	for i, p := range priorities {
		if p == m.priority {
			m.priority = priorities[(i+1)%len(priorities)]
			return
		}
	}
	m.priority = priorities[0]
}

func (m taskFormModel) nextField() taskFormModel {
	m.blurAll()
	m.focused = (m.focused + 1) % formFieldCount
	m.focusCurrent()
	return m
}

func (m taskFormModel) prevField() taskFormModel {
	m.blurAll()
	if m.focused == 0 {
		m.focused = formFieldCount - 1
	} else {
		m.focused--
	}
	m.focusCurrent()
	return m
}

func (m *taskFormModel) blurAll() {
	m.title.Blur()
	m.due.Blur()
	m.tags.Blur()
}

func (m *taskFormModel) focusCurrent() {
	switch m.focused {
	case fieldTitle:
		m.title.Focus()
	case fieldDue:
		m.due.Focus()
	case fieldTags:
		m.tags.Focus()
	}
}

func (m taskFormModel) updateInputs(msg tea.Msg) (taskFormModel, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focused {
	case fieldTitle:
		m.title, cmd = m.title.Update(msg)
	case fieldDue:
		m.due, cmd = m.due.Update(msg)
	case fieldTags:
		m.tags, cmd = m.tags.Update(msg)
	}
	return m, cmd
}

func (m taskFormModel) save() (taskFormModel, tea.Cmd) {
	title := strings.TrimSpace(m.title.Value())
	if title == "" {
		m.err = "title is required"
		return m, nil
	}

	// parse due date
	dueDate, err := parseRelativeDate(m.due.Value())
	if err != nil {
		m.err = err.Error()
		return m, nil
	}

	// parse tags
	var tags []string
	if raw := strings.TrimSpace(m.tags.Value()); raw != "" {
		for _, t := range strings.Split(raw, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	if m.vault == nil {
		m.err = "vault not available"
		return m, nil
	}

	if m.editing != nil {
		// update existing
		m.editing.Title = title
		m.editing.Priority = m.priority
		m.editing.DueDate = dueDate
		m.editing.Tags = tags
		if err := m.vault.Tasks().Update(*m.editing); err != nil {
			m.err = fmt.Sprintf("save task: %v", err)
			return m, nil
		}
	} else {
		// create new
		t := task.New(title)
		t.Priority = m.priority
		t.DueDate = dueDate
		t.Tags = tags
		if err := m.vault.Tasks().Add(t); err != nil {
			m.err = fmt.Sprintf("add task: %v", err)
			return m, nil
		}
	}

	return m, func() tea.Msg { return navigateMsg{view: viewTaskList} }
}

var (
	formLabelStyle    = lipgloss.NewStyle().Foreground(zstyle.Subtext1)
	formActiveLabel   = lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent).Bold(true)
	formPriorityStyle = lipgloss.NewStyle().Foreground(zstyle.Text)
)

func (m taskFormModel) View() string {
	var b strings.Builder
	b.WriteString("\n")

	// mode label
	mode := "new task"
	if m.editing != nil {
		mode = "edit task"
	}
	modeStyle := lipgloss.NewStyle().Foreground(zstyle.ZvaultAccent).Bold(true)
	b.WriteString(fmt.Sprintf("  %s\n\n", modeStyle.Render(mode)))

	// title field
	label := formLabelStyle
	if m.focused == fieldTitle {
		label = formActiveLabel
	}
	b.WriteString(fmt.Sprintf("  %s\n", label.Render("title")))
	b.WriteString(fmt.Sprintf("  %s\n\n", m.title.View()))

	// priority field
	label = formLabelStyle
	if m.focused == fieldPriority {
		label = formActiveLabel
	}
	b.WriteString(fmt.Sprintf("  %s\n", label.Render("priority")))
	b.WriteString(fmt.Sprintf("  %s\n\n", m.renderPrioritySelector()))

	// due date field
	label = formLabelStyle
	if m.focused == fieldDue {
		label = formActiveLabel
	}
	b.WriteString(fmt.Sprintf("  %s\n", label.Render("due date")))
	b.WriteString(fmt.Sprintf("  %s\n\n", m.due.View()))

	// tags field
	label = formLabelStyle
	if m.focused == fieldTags {
		label = formActiveLabel
	}
	b.WriteString(fmt.Sprintf("  %s\n", label.Render("tags")))
	b.WriteString(fmt.Sprintf("  %s\n", m.tags.View()))

	// error
	if m.err != "" {
		b.WriteString(fmt.Sprintf("\n  %s\n", zstyle.StatusErr.Render(m.err)))
	}

	return b.String()
}

func (m taskFormModel) renderPrioritySelector() string {
	display := formatPriority(m.priority)
	style := formPriorityStyle

	switch m.priority {
	case task.PriorityHigh:
		style = lipgloss.NewStyle().Foreground(zstyle.Red).Bold(true)
	case task.PriorityMedium:
		style = lipgloss.NewStyle().Foreground(zstyle.Yellow)
	case task.PriorityLow:
		style = lipgloss.NewStyle().Foreground(zstyle.Text)
	}

	if m.focused == fieldPriority {
		return fmt.Sprintf("▸ %s %s", style.Render(display),
			zstyle.MutedText.Render("(enter/space to cycle)"))
	}
	return fmt.Sprintf("  %s", style.Render(display))
}
