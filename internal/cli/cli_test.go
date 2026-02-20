package cli

import (
	"strings"
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	tests := []struct {
		input   string
		want    time.Time
		wantErr bool
	}{
		{"today", today, false},
		{"tomorrow", today.AddDate(0, 0, 1), false},
		{"next week", today.AddDate(0, 0, 7), false},
		{"+3d", today.AddDate(0, 0, 3), false},
		{"+0d", today, false},
		{"+14d", today.AddDate(0, 0, 14), false},
		{"2026-03-15", time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC), false},
		{"2026-12-31", time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), false},
		{"not-a-date", time.Time{}, true},
		{"", time.Time{}, true},
		{"+d", time.Time{}, true}, // missing number
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseDate(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got %v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("parseDate(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDateCaseInsensitive(t *testing.T) {
	_, err := parseDate("Tomorrow")
	if err != nil {
		t.Fatalf("expected Tomorrow to parse: %v", err)
	}

	_, err = parseDate("NEXT WEEK")
	if err != nil {
		t.Fatalf("expected NEXT WEEK to parse: %v", err)
	}
}

func TestParseTags(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"work", []string{"work"}},
		{"work,personal", []string{"work", "personal"}},
		{"work, personal, dev", []string{"work", "personal", "dev"}},
		{",,,", nil},
		{"  work  ,  dev  ", []string{"work", "dev"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseTags(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseTags(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("parseTags(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseIDs(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"abc123", []string{"abc123"}},
		{"a1,b2,c3", []string{"a1", "b2", "c3"}},
		{"a1, b2, c3", []string{"a1", "b2", "c3"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseIDs(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseIDs(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("parseIDs(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{"h", "high", true},
		{"H", "high", true},
		{"high", "high", true},
		{"HIGH", "high", true},
		{"m", "medium", true},
		{"medium", "medium", true},
		{"l", "low", true},
		{"low", "low", true},
		{"x", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := parsePriority(tt.input)
			if ok != tt.ok {
				t.Fatalf("parsePriority(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && string(got) != tt.want {
				t.Fatalf("parsePriority(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHasFlag(t *testing.T) {
	args := []string{"--show", "-t", "password", "--verbose"}

	if !hasFlag(args, "--show") {
		t.Fatal("expected --show to be found")
	}
	if !hasFlag(args, "--verbose") {
		t.Fatal("expected --verbose to be found")
	}
	if hasFlag(args, "--missing") {
		t.Fatal("expected --missing to not be found")
	}
}

func TestFlagValue(t *testing.T) {
	args := []string{"-t", "password", "-n", "my secret", "--tags", "work,dev"}

	tests := []struct {
		flag string
		want string
	}{
		{"-t", "password"},
		{"-n", "my secret"},
		{"--tags", "work,dev"},
		{"--missing", ""},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			got := flagValue(args, tt.flag)
			if got != tt.want {
				t.Fatalf("flagValue(%q) = %q, want %q", tt.flag, got, tt.want)
			}
		})
	}
}

func TestFlagValueAtEnd(t *testing.T) {
	// flag at end without value
	args := []string{"-t"}
	got := flagValue(args, "-t")
	if got != "" {
		t.Fatalf("expected empty for flag at end, got %q", got)
	}
}

func TestStripFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		valueFlags []string
		bareFlags  []string
		want       []string
	}{
		{
			"value flags",
			[]string{"-p", "h", "-d", "tomorrow", "buy milk"},
			[]string{"-p", "-d"},
			nil,
			[]string{"buy milk"},
		},
		{
			"bare flags",
			[]string{"--show", "abc123"},
			nil,
			[]string{"--show"},
			[]string{"abc123"},
		},
		{
			"mixed",
			[]string{"-p", "h", "--tags", "work,dev", "--pending", "my", "task"},
			[]string{"-p", "--tags"},
			[]string{"--pending"},
			[]string{"my", "task"},
		},
		{
			"no flags",
			[]string{"hello", "world"},
			[]string{"-p"},
			[]string{"--show"},
			[]string{"hello", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripFlags(tt.args, tt.valueFlags, tt.bareFlags)
			if len(got) != len(tt.want) {
				t.Fatalf("stripFlags() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("stripFlags()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFormatDueDate(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	tests := []struct {
		name     string
		date     *time.Time
		contains string
	}{
		{"nil", nil, ""},
		{"today", &today, "today"},
		{"tomorrow", timePtr(today.AddDate(0, 0, 1)), "tomorrow"},
		{"in 3 days", timePtr(today.AddDate(0, 0, 3)), "in 3d"},
		{"overdue", timePtr(today.AddDate(0, 0, -2)), "overdue 2d"},
		{"far future", timePtr(today.AddDate(0, 0, 30)), today.AddDate(0, 0, 30).Format("2006-01-02")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDueDate(tt.date)
			if tt.contains == "" && got != "" {
				t.Fatalf("expected empty, got %q", got)
			}
			if tt.contains != "" && !containsPlain(got, tt.contains) {
				t.Fatalf("formatDueDate() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

// containsPlain strips ANSI codes before checking Contains.
func containsPlain(s, substr string) bool {
	plain := stripANSI(s)
	return strings.Contains(plain, substr)
}

// stripANSI removes ANSI escape sequences.
func stripANSI(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			// skip until 'm'
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++ // skip 'm'
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func TestColorNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	got := red("hello")
	// lipgloss respects NO_COLOR — output should be plain text
	plain := stripANSI(got)
	if plain != "hello" {
		t.Fatalf("expected plain 'hello' with NO_COLOR, got %q (plain: %q)", got, plain)
	}

	got = bold("world")
	plain = stripANSI(got)
	if plain != "world" {
		t.Fatalf("expected plain 'world' with NO_COLOR, got %q (plain: %q)", got, plain)
	}
}

func TestContainsTag(t *testing.T) {
	tags := []string{"work", "dev", "personal"}

	if !containsTag(tags, "work") {
		t.Fatal("expected work to be found")
	}
	if containsTag(tags, "missing") {
		t.Fatal("expected missing to not be found")
	}
	if containsTag(nil, "work") {
		t.Fatal("expected nil tags to not contain anything")
	}
}

func TestCompletionOutput(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		contains []string
	}{
		{
			"bash",
			bashCompletion,
			[]string{"_zvault", "complete -F", "secret", "task", "export", "completion"},
		},
		{
			"zsh",
			zshCompletion,
			[]string{"#compdef zvault", "_zvault", "secret", "task", "export"},
		},
		{
			"fish",
			fishCompletion,
			[]string{"complete -c zvault", "secret", "task", "export", "completion"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, s := range tt.contains {
				if !strings.Contains(tt.script, s) {
					t.Errorf("%s completion missing %q", tt.name, s)
				}
			}
		})
	}
}

func TestNoColorEnvPresence(t *testing.T) {
	// NO_COLOR spec: presence of the variable (even empty) disables color.
	// lipgloss handles this automatically.
	t.Setenv("NO_COLOR", "1")
	got := red("test")
	plain := stripANSI(got)
	if plain != "test" {
		t.Fatalf("expected plain text with NO_COLOR, got %q (plain: %q)", got, plain)
	}
}
