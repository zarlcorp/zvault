// Package dates provides shared date parsing and formatting for zvault.
package dates

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NowFunc allows tests to override the current time.
var NowFunc = time.Now

// FormatRelative returns a human-readable relative date string.
func FormatRelative(t time.Time) string {
	now := NowFunc()
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

// FormatDue returns a display string for a due date. Returns empty for nil.
func FormatDue(t *time.Time) string {
	if t == nil {
		return ""
	}
	return FormatRelative(*t)
}

// IsOverdue returns true if the date is before today.
func IsOverdue(t *time.Time) bool {
	if t == nil {
		return false
	}
	now := NowFunc()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	target := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return target.Before(today)
}

// Parse parses a date string that can be:
//   - empty string: returns (nil, nil)
//   - "today": current day
//   - "tomorrow": next day
//   - "next week": 7 days from now
//   - "+Nd": N days from now
//   - "+Nw": N weeks from now
//   - "YYYY-MM-DD": absolute date
func Parse(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	now := NowFunc()
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

// FormatForEdit returns a YYYY-MM-DD string for editing, or empty for nil.
func FormatForEdit(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}
