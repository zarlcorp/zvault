package dates

import (
	"testing"
	"time"
)

func fixedTime(year int, month time.Month, day int) func() time.Time {
	return func() time.Time {
		return time.Date(year, month, day, 12, 0, 0, 0, time.Local)
	}
}

func ptr(t time.Time) *time.Time { return &t }

func TestFormatRelative(t *testing.T) {
	orig := NowFunc
	defer func() { NowFunc = orig }()
	NowFunc = fixedTime(2026, time.February, 18)

	tests := []struct {
		name string
		date time.Time
		want string
	}{
		{"today", time.Date(2026, 2, 18, 0, 0, 0, 0, time.Local), "today"},
		{"tomorrow", time.Date(2026, 2, 19, 0, 0, 0, 0, time.Local), "tomorrow"},
		{"yesterday", time.Date(2026, 2, 17, 0, 0, 0, 0, time.Local), "yesterday"},
		{"in 3 days", time.Date(2026, 2, 21, 0, 0, 0, 0, time.Local), "in 3 days"},
		{"in 7 days", time.Date(2026, 2, 25, 0, 0, 0, 0, time.Local), "in 7 days"},
		{"next week", time.Date(2026, 2, 28, 0, 0, 0, 0, time.Local), "next week"},
		{"far future", time.Date(2026, 4, 15, 0, 0, 0, 0, time.Local), "Apr 15"},
		{"overdue by 2", time.Date(2026, 2, 16, 0, 0, 0, 0, time.Local), "overdue by 2 days"},
		{"overdue by 5", time.Date(2026, 2, 13, 0, 0, 0, 0, time.Local), "overdue by 5 days"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRelative(tt.date)
			if got != tt.want {
				t.Errorf("FormatRelative(%v) = %q, want %q", tt.date, got, tt.want)
			}
		})
	}
}

func TestFormatDueNil(t *testing.T) {
	if got := FormatDue(nil); got != "" {
		t.Errorf("FormatDue(nil) = %q, want empty", got)
	}
}

func TestIsOverdue(t *testing.T) {
	orig := NowFunc
	defer func() { NowFunc = orig }()
	NowFunc = fixedTime(2026, time.February, 18)

	past := time.Date(2026, 2, 17, 0, 0, 0, 0, time.Local)
	today := time.Date(2026, 2, 18, 0, 0, 0, 0, time.Local)
	future := time.Date(2026, 2, 19, 0, 0, 0, 0, time.Local)

	if !IsOverdue(&past) {
		t.Error("past date should be overdue")
	}
	if IsOverdue(&today) {
		t.Error("today should not be overdue")
	}
	if IsOverdue(&future) {
		t.Error("future date should not be overdue")
	}
	if IsOverdue(nil) {
		t.Error("nil should not be overdue")
	}
}

func TestParse(t *testing.T) {
	orig := NowFunc
	defer func() { NowFunc = orig }()
	NowFunc = fixedTime(2026, time.February, 18)

	today := time.Date(2026, 2, 18, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name    string
		input   string
		want    *time.Time
		wantErr bool
	}{
		{"empty", "", nil, false},
		{"today", "today", ptr(today), false},
		{"tomorrow", "tomorrow", ptr(today.AddDate(0, 0, 1)), false},
		{"next week", "next week", ptr(today.AddDate(0, 0, 7)), false},
		{"plus 3d", "+3d", ptr(today.AddDate(0, 0, 3)), false},
		{"plus 1w", "+1w", ptr(today.AddDate(0, 0, 7)), false},
		{"plus 2w", "+2w", ptr(today.AddDate(0, 0, 14)), false},
		{"absolute", "2026-03-01", ptr(time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)), false},
		{"case insensitive", "Tomorrow", ptr(today.AddDate(0, 0, 1)), false},
		{"case insensitive next", "Next Week", ptr(today.AddDate(0, 0, 7)), false},
		{"invalid", "garbage", nil, true},
		{"bad plus", "+xd", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.want == nil && got != nil {
				t.Errorf("Parse(%q) = %v, want nil", tt.input, got)
				return
			}
			if tt.want != nil {
				if got == nil {
					t.Errorf("Parse(%q) = nil, want %v", tt.input, tt.want)
					return
				}
				if !got.Equal(*tt.want) {
					t.Errorf("Parse(%q) = %v, want %v", tt.input, *got, *tt.want)
				}
			}
		})
	}
}

func TestFormatForEdit(t *testing.T) {
	d := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	got := FormatForEdit(&d)
	if got != "2026-03-15" {
		t.Errorf("FormatForEdit = %q, want 2026-03-15", got)
	}
	if FormatForEdit(nil) != "" {
		t.Error("FormatForEdit(nil) should be empty")
	}
}
