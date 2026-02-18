// Package vault coordinates encrypted storage for secrets and tasks.
package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/zarlcorp/core/pkg/zfilesystem"
	"github.com/zarlcorp/core/pkg/zstore"
	"github.com/zarlcorp/zvault/internal/secret"
	"github.com/zarlcorp/zvault/internal/task"
)

// Vault holds encrypted collections for secrets and tasks.
type Vault struct {
	store   *zstore.Store
	secrets *SecretStore
	tasks   *TaskStore
}

// Open opens or creates a vault at the given directory with the provided password.
func Open(dir string, password string) (*Vault, error) {
	fs := zfilesystem.NewOSFileSystem(dir)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create vault directory: %w", err)
	}

	store, err := zstore.Open(fs, []byte(password))
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	secretCol, err := zstore.NewCollection[secret.Secret](store, "secrets")
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("open secrets collection: %w", err)
	}

	taskCol, err := zstore.NewCollection[task.Task](store, "tasks")
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("open tasks collection: %w", err)
	}

	return &Vault{
		store:   store,
		secrets: &SecretStore{col: secretCol},
		tasks:   &TaskStore{col: taskCol},
	}, nil
}

// OpenFS opens or creates a vault using the provided filesystem (for testing).
func OpenFS(fs zfilesystem.ReadWriteFileFS, password string) (*Vault, error) {
	store, err := zstore.Open(fs, []byte(password))
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	secretCol, err := zstore.NewCollection[secret.Secret](store, "secrets")
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("open secrets collection: %w", err)
	}

	taskCol, err := zstore.NewCollection[task.Task](store, "tasks")
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("open tasks collection: %w", err)
	}

	return &Vault{
		store:   store,
		secrets: &SecretStore{col: secretCol},
		tasks:   &TaskStore{col: taskCol},
	}, nil
}

// Secrets returns the secret store.
func (v *Vault) Secrets() *SecretStore { return v.secrets }

// Tasks returns the task store.
func (v *Vault) Tasks() *TaskStore { return v.tasks }

// Close erases keys and closes the underlying store.
func (v *Vault) Close() error { return v.store.Close() }

// DefaultDir returns the default data directory following XDG convention.
func DefaultDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "zvault")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "zvault")
	}
	return filepath.Join(home, ".local", "share", "zvault")
}

// PasswordFromEnv reads the vault password from ZVAULT_PASSWORD environment variable.
// Returns empty string if not set.
func PasswordFromEnv() string {
	return os.Getenv("ZVAULT_PASSWORD")
}

// SecretStore wraps a zstore collection for secrets.
type SecretStore struct {
	col *zstore.Collection[secret.Secret]
}

// Add stores a new secret.
func (s *SecretStore) Add(sec secret.Secret) error {
	return s.col.Put(sec.ID, sec)
}

// Get retrieves a secret by ID.
func (s *SecretStore) Get(id string) (secret.Secret, error) {
	return s.col.Get(id)
}

// List returns all secrets.
func (s *SecretStore) List() ([]secret.Secret, error) {
	return s.col.List()
}

// Update overwrites a secret, setting UpdatedAt.
func (s *SecretStore) Update(sec secret.Secret) error {
	sec.UpdatedAt = time.Now()
	return s.col.Put(sec.ID, sec)
}

// Delete removes a secret by ID.
func (s *SecretStore) Delete(id string) error {
	return s.col.Delete(id)
}

// Search returns secrets matching the query against name (case-insensitive
// substring), tags (exact match), or type (exact match).
func (s *SecretStore) Search(query string) ([]secret.Secret, error) {
	all, err := s.col.List()
	if err != nil {
		return nil, err
	}

	q := strings.ToLower(query)
	var results []secret.Secret
	for _, sec := range all {
		if matches(sec, q) {
			results = append(results, sec)
		}
	}
	return results, nil
}

func matches(sec secret.Secret, query string) bool {
	// name: case-insensitive substring
	if strings.Contains(strings.ToLower(sec.Name), query) {
		return true
	}
	// tags: exact match
	if slices.Contains(sec.Tags, query) {
		return true
	}
	// type: exact match
	if string(sec.Type) == query {
		return true
	}
	return false
}

// TaskStore wraps a zstore collection for tasks.
type TaskStore struct {
	col *zstore.Collection[task.Task]
}

// Add stores a new task.
func (s *TaskStore) Add(tk task.Task) error {
	return s.col.Put(tk.ID, tk)
}

// Get retrieves a task by ID.
func (s *TaskStore) Get(id string) (task.Task, error) {
	return s.col.Get(id)
}

// List returns tasks matching the filter. Zero-value filter fields match all.
func (s *TaskStore) List(f task.Filter) ([]task.Task, error) {
	all, err := s.col.List()
	if err != nil {
		return nil, err
	}

	var results []task.Task
	for _, tk := range all {
		if matchesFilter(tk, f) {
			results = append(results, tk)
		}
	}
	return results, nil
}

// Update overwrites a task.
func (s *TaskStore) Update(tk task.Task) error {
	return s.col.Put(tk.ID, tk)
}

// Delete removes a task by ID.
func (s *TaskStore) Delete(id string) error {
	return s.col.Delete(id)
}

// ClearDone removes all completed tasks and returns the count deleted.
func (s *TaskStore) ClearDone() (int, error) {
	all, err := s.col.List()
	if err != nil {
		return 0, err
	}

	count := 0
	for _, tk := range all {
		if tk.Done {
			if err := s.col.Delete(tk.ID); err != nil {
				return count, fmt.Errorf("delete task %s: %w", tk.ID, err)
			}
			count++
		}
	}
	return count, nil
}

func matchesFilter(tk task.Task, f task.Filter) bool {
	// status filter
	switch f.Status {
	case task.FilterPending:
		if tk.Done {
			return false
		}
	case task.FilterDone:
		if !tk.Done {
			return false
		}
	}
	// FilterAll and zero value match everything

	// priority filter
	if f.Priority != task.PriorityNone && tk.Priority != f.Priority {
		return false
	}

	// tag filter
	if f.Tag != "" && !slices.Contains(tk.Tags, f.Tag) {
		return false
	}

	return true
}
