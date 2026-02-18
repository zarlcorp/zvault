package vault

import (
	"testing"
	"time"

	"github.com/zarlcorp/core/pkg/zfilesystem"
	"github.com/zarlcorp/zvault/internal/secret"
	"github.com/zarlcorp/zvault/internal/task"
)

func openTestVault(t *testing.T) *Vault {
	t.Helper()
	v, err := OpenFS(zfilesystem.NewMemFS(), "test-password")
	if err != nil {
		t.Fatalf("open vault: %v", err)
	}
	t.Cleanup(func() { v.Close() })
	return v
}

// --- SecretStore tests ---

func TestSecretAddAndGet(t *testing.T) {
	v := openTestVault(t)
	s := secret.NewPassword("github", "https://github.com", "user", "pass")

	if err := v.Secrets().Add(s); err != nil {
		t.Fatalf("add: %v", err)
	}

	got, err := v.Secrets().Get(s.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.Name != "github" {
		t.Fatalf("name = %q, want %q", got.Name, "github")
	}
	if got.Username() != "user" {
		t.Fatalf("username = %q, want %q", got.Username(), "user")
	}
}

func TestSecretGetNotFound(t *testing.T) {
	v := openTestVault(t)

	_, err := v.Secrets().Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestSecretList(t *testing.T) {
	v := openTestVault(t)

	secrets := []secret.Secret{
		secret.NewPassword("github", "url", "u", "p"),
		secret.NewAPIKey("stripe", "stripe.com", "sk_123"),
		secret.NewNote("wifi", "password123"),
	}

	for _, s := range secrets {
		if err := v.Secrets().Add(s); err != nil {
			t.Fatalf("add: %v", err)
		}
	}

	got, err := v.Secrets().List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("list len = %d, want 3", len(got))
	}
}

func TestSecretUpdate(t *testing.T) {
	v := openTestVault(t)
	s := secret.NewPassword("github", "url", "user", "old-pass")

	if err := v.Secrets().Add(s); err != nil {
		t.Fatalf("add: %v", err)
	}

	s.Fields["password"] = "new-pass"
	if err := v.Secrets().Update(s); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := v.Secrets().Get(s.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.Password() != "new-pass" {
		t.Fatalf("password = %q, want %q", got.Password(), "new-pass")
	}
}

func TestSecretDelete(t *testing.T) {
	v := openTestVault(t)
	s := secret.NewNote("temp", "will delete")

	if err := v.Secrets().Add(s); err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := v.Secrets().Delete(s.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := v.Secrets().Get(s.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestSecretDeleteNotFound(t *testing.T) {
	v := openTestVault(t)

	err := v.Secrets().Delete("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestSecretSearchByName(t *testing.T) {
	v := openTestVault(t)

	v.Secrets().Add(secret.NewPassword("GitHub Login", "url", "u", "p"))
	v.Secrets().Add(secret.NewPassword("GitLab Login", "url", "u", "p"))
	v.Secrets().Add(secret.NewAPIKey("Stripe Key", "stripe", "key"))

	tests := []struct {
		query string
		want  int
	}{
		{"github", 1},   // case-insensitive
		{"login", 2},    // substring match
		{"STRIPE", 1},   // case-insensitive
		{"nothing", 0},  // no match
		{"Git", 2},      // prefix match
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got, err := v.Secrets().Search(tt.query)
			if err != nil {
				t.Fatalf("search: %v", err)
			}
			if len(got) != tt.want {
				t.Fatalf("search %q returned %d results, want %d", tt.query, len(got), tt.want)
			}
		})
	}
}

func TestSecretSearchByTag(t *testing.T) {
	v := openTestVault(t)

	s1 := secret.NewPassword("github", "url", "u", "p")
	s1.Tags = []string{"work", "dev"}
	v.Secrets().Add(s1)

	s2 := secret.NewAPIKey("stripe", "stripe", "key")
	s2.Tags = []string{"work", "billing"}
	v.Secrets().Add(s2)

	s3 := secret.NewNote("personal wifi", "pass")
	s3.Tags = []string{"home"}
	v.Secrets().Add(s3)

	tests := []struct {
		query string
		want  int
	}{
		{"work", 2},
		{"dev", 1},
		{"home", 1},
		{"nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got, err := v.Secrets().Search(tt.query)
			if err != nil {
				t.Fatalf("search: %v", err)
			}
			if len(got) != tt.want {
				t.Fatalf("search %q returned %d results, want %d", tt.query, len(got), tt.want)
			}
		})
	}
}

func TestSecretSearchByType(t *testing.T) {
	v := openTestVault(t)

	// use names that won't match the type string
	v.Secrets().Add(secret.NewPassword("login-1", "url", "u", "p"))
	v.Secrets().Add(secret.NewPassword("login-2", "url", "u", "p"))
	v.Secrets().Add(secret.NewAPIKey("key-1", "svc", "k"))
	v.Secrets().Add(secret.NewNote("memo-1", "content"))

	tests := []struct {
		query string
		want  int
	}{
		{"password", 2},
		{"apikey", 1},
		{"note", 1},
		{"sshkey", 0},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got, err := v.Secrets().Search(tt.query)
			if err != nil {
				t.Fatalf("search: %v", err)
			}
			if len(got) != tt.want {
				t.Fatalf("search %q returned %d results, want %d", tt.query, len(got), tt.want)
			}
		})
	}
}

// --- TaskStore tests ---

func TestTaskAddAndGet(t *testing.T) {
	v := openTestVault(t)
	tk := task.New("buy milk")

	if err := v.Tasks().Add(tk); err != nil {
		t.Fatalf("add: %v", err)
	}

	got, err := v.Tasks().Get(tk.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if got.Title != "buy milk" {
		t.Fatalf("title = %q, want %q", got.Title, "buy milk")
	}
}

func TestTaskGetNotFound(t *testing.T) {
	v := openTestVault(t)

	_, err := v.Tasks().Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing task")
	}
}

func TestTaskUpdate(t *testing.T) {
	v := openTestVault(t)
	tk := task.New("draft report")

	if err := v.Tasks().Add(tk); err != nil {
		t.Fatalf("add: %v", err)
	}

	now := time.Now()
	tk.Done = true
	tk.CompletedAt = &now
	if err := v.Tasks().Update(tk); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := v.Tasks().Get(tk.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if !got.Done {
		t.Fatal("task should be done after update")
	}
}

func TestTaskDelete(t *testing.T) {
	v := openTestVault(t)
	tk := task.New("temp task")

	if err := v.Tasks().Add(tk); err != nil {
		t.Fatalf("add: %v", err)
	}

	if err := v.Tasks().Delete(tk.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := v.Tasks().Get(tk.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestTaskListFilterByStatus(t *testing.T) {
	v := openTestVault(t)

	// add 3 tasks: 2 pending, 1 done
	t1 := task.New("pending-1")
	t2 := task.New("pending-2")
	t3 := task.New("done-1")
	now := time.Now()
	t3.Done = true
	t3.CompletedAt = &now

	for _, tk := range []task.Task{t1, t2, t3} {
		if err := v.Tasks().Add(tk); err != nil {
			t.Fatalf("add: %v", err)
		}
	}

	tests := []struct {
		name   string
		status task.FilterStatus
		want   int
	}{
		{"all", task.FilterAll, 3},
		{"pending", task.FilterPending, 2},
		{"done", task.FilterDone, 1},
		{"zero value (all)", "", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := v.Tasks().List(task.Filter{Status: tt.status})
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(got) != tt.want {
				t.Fatalf("list returned %d, want %d", len(got), tt.want)
			}
		})
	}
}

func TestTaskListFilterByPriority(t *testing.T) {
	v := openTestVault(t)

	t1 := task.New("high-1")
	t1.Priority = task.PriorityHigh

	t2 := task.New("high-2")
	t2.Priority = task.PriorityHigh

	t3 := task.New("low-1")
	t3.Priority = task.PriorityLow

	t4 := task.New("none-1")

	for _, tk := range []task.Task{t1, t2, t3, t4} {
		if err := v.Tasks().Add(tk); err != nil {
			t.Fatalf("add: %v", err)
		}
	}

	tests := []struct {
		name     string
		priority task.Priority
		want     int
	}{
		{"all", task.PriorityNone, 4},
		{"high", task.PriorityHigh, 2},
		{"low", task.PriorityLow, 1},
		{"medium (none)", task.PriorityMedium, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := v.Tasks().List(task.Filter{Priority: tt.priority})
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(got) != tt.want {
				t.Fatalf("list returned %d, want %d", len(got), tt.want)
			}
		})
	}
}

func TestTaskListFilterByTag(t *testing.T) {
	v := openTestVault(t)

	t1 := task.New("task-1")
	t1.Tags = []string{"work", "urgent"}

	t2 := task.New("task-2")
	t2.Tags = []string{"work"}

	t3 := task.New("task-3")
	t3.Tags = []string{"personal"}

	for _, tk := range []task.Task{t1, t2, t3} {
		if err := v.Tasks().Add(tk); err != nil {
			t.Fatalf("add: %v", err)
		}
	}

	tests := []struct {
		tag  string
		want int
	}{
		{"work", 2},
		{"urgent", 1},
		{"personal", 1},
		{"nonexistent", 0},
		{"", 3}, // empty = all
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got, err := v.Tasks().List(task.Filter{Tag: tt.tag})
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(got) != tt.want {
				t.Fatalf("list tag=%q returned %d, want %d", tt.tag, len(got), tt.want)
			}
		})
	}
}

func TestTaskListCombinedFilters(t *testing.T) {
	v := openTestVault(t)

	t1 := task.New("high-work")
	t1.Priority = task.PriorityHigh
	t1.Tags = []string{"work"}

	t2 := task.New("high-personal")
	t2.Priority = task.PriorityHigh
	t2.Tags = []string{"personal"}

	t3 := task.New("low-work")
	t3.Priority = task.PriorityLow
	t3.Tags = []string{"work"}

	for _, tk := range []task.Task{t1, t2, t3} {
		if err := v.Tasks().Add(tk); err != nil {
			t.Fatalf("add: %v", err)
		}
	}

	got, err := v.Tasks().List(task.Filter{
		Priority: task.PriorityHigh,
		Tag:      "work",
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("list returned %d, want 1", len(got))
	}
	if got[0].Title != "high-work" {
		t.Fatalf("title = %q, want %q", got[0].Title, "high-work")
	}
}

func TestTaskClearDone(t *testing.T) {
	v := openTestVault(t)

	now := time.Now()

	t1 := task.New("pending")
	t2 := task.New("done-1")
	t2.Done = true
	t2.CompletedAt = &now
	t3 := task.New("done-2")
	t3.Done = true
	t3.CompletedAt = &now
	t4 := task.New("also pending")

	for _, tk := range []task.Task{t1, t2, t3, t4} {
		if err := v.Tasks().Add(tk); err != nil {
			t.Fatalf("add: %v", err)
		}
	}

	count, err := v.Tasks().ClearDone()
	if err != nil {
		t.Fatalf("clear done: %v", err)
	}
	if count != 2 {
		t.Fatalf("cleared %d, want 2", count)
	}

	// only pending tasks should remain
	remaining, err := v.Tasks().List(task.Filter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(remaining) != 2 {
		t.Fatalf("remaining = %d, want 2", len(remaining))
	}
	for _, tk := range remaining {
		if tk.Done {
			t.Fatalf("task %q should not be done", tk.Title)
		}
	}
}

func TestTaskClearDoneEmpty(t *testing.T) {
	v := openTestVault(t)

	count, err := v.Tasks().ClearDone()
	if err != nil {
		t.Fatalf("clear done: %v", err)
	}
	if count != 0 {
		t.Fatalf("cleared %d, want 0", count)
	}
}

func TestTaskClearDoneNoDoneTasks(t *testing.T) {
	v := openTestVault(t)

	v.Tasks().Add(task.New("pending-1"))
	v.Tasks().Add(task.New("pending-2"))

	count, err := v.Tasks().ClearDone()
	if err != nil {
		t.Fatalf("clear done: %v", err)
	}
	if count != 0 {
		t.Fatalf("cleared %d, want 0", count)
	}

	remaining, err := v.Tasks().List(task.Filter{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(remaining) != 2 {
		t.Fatalf("remaining = %d, want 2", len(remaining))
	}
}

// --- Vault lifecycle tests ---

func TestVaultOpenClose(t *testing.T) {
	v, err := OpenFS(zfilesystem.NewMemFS(), "password")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := v.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestVaultWrongPassword(t *testing.T) {
	fs := zfilesystem.NewMemFS()

	v, err := OpenFS(fs, "correct")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	v.Close()

	_, err = OpenFS(fs, "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestVaultPersistence(t *testing.T) {
	fs := zfilesystem.NewMemFS()

	// first session: write data
	v1, err := OpenFS(fs, "password")
	if err != nil {
		t.Fatalf("first open: %v", err)
	}

	sec := secret.NewNote("persist-test", "hello")
	if err := v1.Secrets().Add(sec); err != nil {
		t.Fatalf("add: %v", err)
	}

	tk := task.New("persist-task")
	if err := v1.Tasks().Add(tk); err != nil {
		t.Fatalf("add task: %v", err)
	}
	v1.Close()

	// second session: read data
	v2, err := OpenFS(fs, "password")
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	defer v2.Close()

	got, err := v2.Secrets().Get(sec.ID)
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if got.Content() != "hello" {
		t.Fatalf("content = %q, want %q", got.Content(), "hello")
	}

	gotTask, err := v2.Tasks().Get(tk.ID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if gotTask.Title != "persist-task" {
		t.Fatalf("title = %q, want %q", gotTask.Title, "persist-task")
	}
}

func TestDefaultDir(t *testing.T) {
	dir := DefaultDir()
	if dir == "" {
		t.Fatal("default dir is empty")
	}
}

func TestPasswordFromEnv(t *testing.T) {
	t.Setenv("ZVAULT_PASSWORD", "env-pass")
	if got := PasswordFromEnv(); got != "env-pass" {
		t.Fatalf("got %q, want %q", got, "env-pass")
	}
}

func TestPasswordFromEnvEmpty(t *testing.T) {
	t.Setenv("ZVAULT_PASSWORD", "")
	if got := PasswordFromEnv(); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}
