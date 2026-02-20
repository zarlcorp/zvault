package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zarlcorp/core/pkg/zstyle"
	"github.com/zarlcorp/zvault/internal/dates"
	"github.com/zarlcorp/zvault/internal/task"
	"github.com/zarlcorp/zvault/internal/vault"
)

// taskDetailModel shows full details for a single task.
type taskDetailModel struct {
	vault   *vault.Vault
	task    task.Task
	loaded  bool
	confirm confirmKind
	width   int
	height  int
}

func newTaskDetailModel(v *vault.Vault) taskDetailModel {
	return taskDetailModel{vault: v}
}

func (m *taskDetailModel) loadTask(id string) {
	if m.vault == nil {
		return
	}
	t, err := m.vault.Tasks().Get(id)
	if err != nil {
		m.loaded = false
		return
	}
	m.task = t
	m.loaded = true
	m.confirm = confirmNone
}

func (m taskDetailModel) Update(msg tea.Msg) (taskDetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case navigateMsg:
		if msg.view == viewTaskDetail {
			if id, ok := msg.data.(string); ok {
				m.loadTask(id)
			}
		}
		return m, nil

	case tea.KeyMsg:
		// handle confirmation first
		if m.confirm != confirmNone {
			return m.handleConfirm(msg)
		}

		switch {
		case key.Matches(msg, zstyle.KeyBack):
			return m, func() tea.Msg { return navigateMsg{view: viewTaskList} }

		case msg.String() == "e":
			if m.loaded {
				t := m.task
				return m, func() tea.Msg { return navigateMsg{view: viewTaskForm, data: t} }
			}

		case msg.String() == " ":
			if m.loaded {
				return m.toggleDone()
			}

		case msg.String() == "d":
			if m.loaded {
				m.confirm = confirmDelete
			}
		}
	}
	return m, nil
}

func (m taskDetailModel) handleConfirm(msg tea.KeyMsg) (taskDetailModel, tea.Cmd) {
	switch msg.String() {
	case "y":
		if m.confirm == confirmDelete {
			m.confirm = confirmNone
			return m.deleteTask()
		}
	case "n", "esc":
		m.confirm = confirmNone
	}
	return m, nil
}

func (m taskDetailModel) toggleDone() (taskDetailModel, tea.Cmd) {
	if m.vault == nil {
		return m, nil
	}

	m.task.Done = !m.task.Done
	if m.task.Done {
		now := dates.NowFunc()
		m.task.CompletedAt = &now
	} else {
		m.task.CompletedAt = nil
	}

	if err := m.vault.Tasks().Update(m.task); err != nil {
		return m, func() tea.Msg { return errMsg{err: err} }
	}
	return m, nil
}

func (m taskDetailModel) deleteTask() (taskDetailModel, tea.Cmd) {
	if m.vault == nil {
		return m, nil
	}

	if err := m.vault.Tasks().Delete(m.task.ID); err != nil {
		return m, func() tea.Msg { return errMsg{err: err} }
	}
	return m, func() tea.Msg { return navigateMsg{view: viewTaskList} }
}

var (
	detailLabelStyle = lipgloss.NewStyle().Foreground(zstyle.Overlay1).Width(12)
	detailValueStyle = lipgloss.NewStyle().Foreground(zstyle.Text)
)

func (m taskDetailModel) View() string {
	if !m.loaded {
		return fmt.Sprintf("\n  %s\n", zstyle.MutedText.Render("no task selected"))
	}

	var b strings.Builder
	b.WriteString("\n")

	// title
	titleStyle := lipgloss.NewStyle().Foreground(zstyle.Text).Bold(true)
	b.WriteString(fmt.Sprintf("  %s\n\n", titleStyle.Render(m.task.Title)))

	// status
	status := "pending"
	statusStyle := detailValueStyle.Foreground(zstyle.Yellow)
	if m.task.Done {
		status = "done"
		statusStyle = detailValueStyle.Foreground(zstyle.Green)
	}
	b.WriteString(fmt.Sprintf("  %s%s\n", detailLabelStyle.Render("status"), statusStyle.Render(status)))

	// priority
	pri := formatPriority(m.task.Priority)
	priStyle := priorityStyle(m.task.Priority)
	b.WriteString(fmt.Sprintf("  %s%s\n", detailLabelStyle.Render("priority"), priStyle.Render(pri)))

	// due date
	if m.task.DueDate != nil {
		dueStr := dates.FormatRelative(*m.task.DueDate)
		dateStr := m.task.DueDate.Format("2006-01-02")
		display := fmt.Sprintf("%s (%s)", dueStr, dateStr)
		dueStyle := detailValueStyle
		if !m.task.Done && dates.IsOverdue(m.task.DueDate) {
			dueStyle = detailValueStyle.Foreground(zstyle.Red)
		}
		b.WriteString(fmt.Sprintf("  %s%s\n", detailLabelStyle.Render("due"), dueStyle.Render(display)))
	}

	// tags
	if len(m.task.Tags) > 0 {
		var tagStrs []string
		for _, tag := range m.task.Tags {
			tagStrs = append(tagStrs, "#"+tag)
		}
		b.WriteString(fmt.Sprintf("  %s%s\n",
			detailLabelStyle.Render("tags"),
			taskTagStyle.Render(strings.Join(tagStrs, " "))))
	}

	// created
	b.WriteString(fmt.Sprintf("  %s%s\n",
		detailLabelStyle.Render("created"),
		zstyle.MutedText.Render(m.task.CreatedAt.Format("2006-01-02 15:04"))))

	// completed
	if m.task.CompletedAt != nil {
		b.WriteString(fmt.Sprintf("  %s%s\n",
			detailLabelStyle.Render("completed"),
			zstyle.MutedText.Render(m.task.CompletedAt.Format("2006-01-02 15:04"))))
	}

	// confirmation prompt
	if m.confirm == confirmDelete {
		b.WriteString(fmt.Sprintf("\n  %s\n",
			zstyle.StatusWarn.Render(fmt.Sprintf("delete \"%s\"? (y/n)", m.task.Title))))
	}

	return b.String()
}

func formatPriority(p task.Priority) string {
	switch p {
	case task.PriorityHigh:
		return "high"
	case task.PriorityMedium:
		return "medium"
	case task.PriorityLow:
		return "low"
	default:
		return "none"
	}
}

func priorityStyle(p task.Priority) lipgloss.Style {
	switch p {
	case task.PriorityHigh:
		return detailValueStyle.Foreground(zstyle.Red)
	case task.PriorityMedium:
		return detailValueStyle.Foreground(zstyle.Yellow)
	default:
		return detailValueStyle
	}
}
