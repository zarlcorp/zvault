// Package task defines the task model for zvault.
package task

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Priority levels for tasks.
type Priority string

const (
	PriorityNone   Priority = ""
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Task holds a single task item.
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Done        bool       `json:"done"`
	Priority    Priority   `json:"priority"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Tags        []string   `json:"tags"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// FilterStatus filters tasks by completion state.
type FilterStatus string

const (
	FilterAll     FilterStatus = "all"
	FilterPending FilterStatus = "pending"
	FilterDone    FilterStatus = "done"
)

// Filter controls which tasks are returned by List.
type Filter struct {
	Status   FilterStatus
	Priority Priority
	Tag      string
}

// New creates a task with the given title.
func New(title string) (Task, error) {
	id, err := generateID()
	if err != nil {
		return Task{}, fmt.Errorf("new task: %w", err)
	}
	now := time.Now()
	return Task{
		ID:        id,
		Title:     title,
		CreatedAt: now,
	}, nil
}

// generateID returns an 8-character hex string from 4 random bytes.
func generateID() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
