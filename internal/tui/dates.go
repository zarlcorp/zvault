package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// nowFunc allows tests to override the current time.
var nowFunc = time.Now

// formatRelativeDate returns a human-readable relative date string.
// Future: "tomorrow", "in 3 days", "next week", "Mar 1"
// Past: "yesterday", "overdue by 2 days"
// Today: "today"
func formatRelativeDate(t time.Time) string {
	now := nowFunc()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	target := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	days := int(target.Sub(today).Hours() / 24)

	switch {
	case days == 0:
		return "today"
	case days == 1:
		return "tomorrow"
	case days == -1:
		return "yesterday"
	case days > 1 && days <= 7:
		return fmt.Sprintf("in %d days", days)
	case days > 7 && days <= 14:
		return "next week"
	case days > 14:
		return t.Format("Jan 2")
	case days < -1:
		return fmt.Sprintf("overdue by %d days", -days)
	}
	return t.Format("Jan 2")
}

// formatDueDate returns a display string for a due date in the list view.
// Returns empty string for nil.
func formatDueDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatRelativeDate(*t)
}

// isOverdue returns true if the date is before today.
func isOverdue(t *time.Time) bool {
	if t == nil {
		return false
	}
	now := nowFunc()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	target := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return target.Before(today)
}

// parseRelativeDate parses a date string that can be:
// - YYYY-MM-DD: absolute date
// - tomorrow: next day
// - +3d: 3 days from now
// - +1w: 1 week from now
// - next week: 7 days from now
// - today: current day
// Returns nil time pointer and error for empty/invalid input.
func parseRelativeDate(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	now := nowFunc()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	lower := strings.ToLower(s)

	switch lower {
	case "today":
		return &today, nil
	case "tomorrow":
		t := today.AddDate(0, 0, 1)
		return &t, nil
	case "next week":
		t := today.AddDate(0, 0, 7)
		return &t, nil
	}

	// +Nd format (days)
	if strings.HasPrefix(lower, "+") && strings.HasSuffix(lower, "d") {
		numStr := lower[1 : len(lower)-1]
		n, err := strconv.Atoi(numStr)
		if err != nil {
			return nil, fmt.Errorf("invalid day offset: %s", s)
		}
		t := today.AddDate(0, 0, n)
		return &t, nil
	}

	// +Nw format (weeks)
	if strings.HasPrefix(lower, "+") && strings.HasSuffix(lower, "w") {
		numStr := lower[1 : len(lower)-1]
		n, err := strconv.Atoi(numStr)
		if err != nil {
			return nil, fmt.Errorf("invalid week offset: %s", s)
		}
		t := today.AddDate(0, 0, n*7)
		return &t, nil
	}

	// YYYY-MM-DD
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %s (use YYYY-MM-DD, tomorrow, +3d, next week)", s)
	}
	return &t, nil
}

// formatDateForEdit returns a YYYY-MM-DD string for editing, or empty for nil.
func formatDateForEdit(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
