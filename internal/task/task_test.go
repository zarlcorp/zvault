package task

import (
	"encoding/hex"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tk := New("buy milk")

	if tk.Title != "buy milk" {
		t.Fatalf("title = %q, want %q", tk.Title, "buy milk")
	}
	if tk.Done {
		t.Fatal("new task should not be done")
	}
	if tk.Priority != PriorityNone {
		t.Fatalf("priority = %q, want empty", tk.Priority)
	}
	if tk.DueDate != nil {
		t.Fatal("due_date should be nil")
	}
	if tk.CompletedAt != nil {
		t.Fatal("completed_at should be nil")
	}
	assertValidID(t, tk.ID)
	if tk.CreatedAt.IsZero() {
		t.Fatal("created_at is zero")
	}
}

func TestTaskFields(t *testing.T) {
	tk := New("deploy")
	tk.Priority = PriorityHigh
	tk.Tags = []string{"ops", "urgent"}

	due := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	tk.DueDate = &due

	if tk.Priority != PriorityHigh {
		t.Fatalf("priority = %q, want %q", tk.Priority, PriorityHigh)
	}
	if len(tk.Tags) != 2 {
		t.Fatalf("tags len = %d, want 2", len(tk.Tags))
	}
	if tk.DueDate == nil || !tk.DueDate.Equal(due) {
		t.Fatalf("due_date = %v, want %v", tk.DueDate, due)
	}
}

func TestTaskCompletion(t *testing.T) {
	tk := New("finish spec")
	if tk.Done {
		t.Fatal("should not be done initially")
	}

	now := time.Now()
	tk.Done = true
	tk.CompletedAt = &now

	if !tk.Done {
		t.Fatal("should be done after marking")
	}
	if tk.CompletedAt == nil {
		t.Fatal("completed_at should be set")
	}
}

func TestPriorityConstants(t *testing.T) {
	tests := []struct {
		priority Priority
		want     string
	}{
		{PriorityNone, ""},
		{PriorityLow, "low"},
		{PriorityMedium, "medium"},
		{PriorityHigh, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.priority) != tt.want {
				t.Fatalf("got %q, want %q", tt.priority, tt.want)
			}
		})
	}
}

func TestFilterStatusConstants(t *testing.T) {
	tests := []struct {
		status FilterStatus
		want   string
	}{
		{FilterAll, "all"},
		{FilterPending, "pending"},
		{FilterDone, "done"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Fatalf("got %q, want %q", tt.status, tt.want)
			}
		})
	}
}

func TestIDUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for range 100 {
		tk := New("test")
		if seen[tk.ID] {
			t.Fatalf("duplicate ID: %s", tk.ID)
		}
		seen[tk.ID] = true
	}
}

func TestIDFormat(t *testing.T) {
	for range 50 {
		tk := New("test")
		assertValidID(t, tk.ID)
	}
}

func assertValidID(t *testing.T, id string) {
	t.Helper()
	if len(id) != 8 {
		t.Fatalf("id length = %d, want 8", len(id))
	}
	if _, err := hex.DecodeString(id); err != nil {
		t.Fatalf("id %q is not valid hex: %v", id, err)
	}
}
