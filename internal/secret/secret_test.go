package secret_test

import (
	"encoding/hex"
	"testing"

	"github.com/zarlcorp/zvault/internal/secret"
)

func TestNewPassword(t *testing.T) {
	s := secret.NewPassword("github", "https://github.com", "user", "pass123")

	if s.Name != "github" {
		t.Fatalf("name = %q, want %q", s.Name, "github")
	}
	if s.Type != secret.TypePassword {
		t.Fatalf("type = %q, want %q", s.Type, secret.TypePassword)
	}
	if s.URL() != "https://github.com" {
		t.Fatalf("url = %q, want %q", s.URL(), "https://github.com")
	}
	if s.Username() != "user" {
		t.Fatalf("username = %q, want %q", s.Username(), "user")
	}
	if s.Password() != "pass123" {
		t.Fatalf("password = %q, want %q", s.Password(), "pass123")
	}
	assertValidID(t, s.ID)
	assertTimestampsSet(t, s)
}

func TestNewAPIKey(t *testing.T) {
	s := secret.NewAPIKey("stripe", "stripe.com", "sk_test_123")

	if s.Name != "stripe" {
		t.Fatalf("name = %q, want %q", s.Name, "stripe")
	}
	if s.Type != secret.TypeAPIKey {
		t.Fatalf("type = %q, want %q", s.Type, secret.TypeAPIKey)
	}
	if s.Service() != "stripe.com" {
		t.Fatalf("service = %q, want %q", s.Service(), "stripe.com")
	}
	if s.Key() != "sk_test_123" {
		t.Fatalf("key = %q, want %q", s.Key(), "sk_test_123")
	}
	assertValidID(t, s.ID)
	assertTimestampsSet(t, s)
}

func TestNewSSHKey(t *testing.T) {
	s := secret.NewSSHKey("server", "prod-server", "-----BEGIN...", "ssh-ed25519 AAAA...")

	if s.Name != "server" {
		t.Fatalf("name = %q, want %q", s.Name, "server")
	}
	if s.Type != secret.TypeSSHKey {
		t.Fatalf("type = %q, want %q", s.Type, secret.TypeSSHKey)
	}
	if s.Label() != "prod-server" {
		t.Fatalf("label = %q, want %q", s.Label(), "prod-server")
	}
	if s.PrivateKey() != "-----BEGIN..." {
		t.Fatalf("private_key = %q, want %q", s.PrivateKey(), "-----BEGIN...")
	}
	if s.PublicKey() != "ssh-ed25519 AAAA..." {
		t.Fatalf("public_key = %q, want %q", s.PublicKey(), "ssh-ed25519 AAAA...")
	}
	assertValidID(t, s.ID)
	assertTimestampsSet(t, s)
}

func TestNewNote(t *testing.T) {
	s := secret.NewNote("wifi", "my secret wifi password")

	if s.Name != "wifi" {
		t.Fatalf("name = %q, want %q", s.Name, "wifi")
	}
	if s.Type != secret.TypeNote {
		t.Fatalf("type = %q, want %q", s.Type, secret.TypeNote)
	}
	if s.Content() != "my secret wifi password" {
		t.Fatalf("content = %q, want %q", s.Content(), "my secret wifi password")
	}
	assertValidID(t, s.ID)
	assertTimestampsSet(t, s)
}

func TestFieldGetters(t *testing.T) {
	tests := []struct {
		name   string
		secret secret.Secret
		getter func(secret.Secret) string
		want   string
	}{
		{
			name:   "password/url",
			secret: secret.NewPassword("test", "https://example.com", "u", "p"),
			getter: secret.Secret.URL,
			want:   "https://example.com",
		},
		{
			name:   "password/username",
			secret: secret.NewPassword("test", "url", "admin", "p"),
			getter: secret.Secret.Username,
			want:   "admin",
		},
		{
			name:   "password/password",
			secret: secret.NewPassword("test", "url", "u", "s3cret"),
			getter: secret.Secret.Password,
			want:   "s3cret",
		},
		{
			name:   "apikey/service",
			secret: secret.NewAPIKey("test", "aws", "key"),
			getter: secret.Secret.Service,
			want:   "aws",
		},
		{
			name:   "apikey/key",
			secret: secret.NewAPIKey("test", "svc", "AKIAIOSFODNN7EXAMPLE"),
			getter: secret.Secret.Key,
			want:   "AKIAIOSFODNN7EXAMPLE",
		},
		{
			name:   "sshkey/label",
			secret: secret.NewSSHKey("test", "my-key", "priv", "pub"),
			getter: secret.Secret.Label,
			want:   "my-key",
		},
		{
			name:   "sshkey/private_key",
			secret: secret.NewSSHKey("test", "lbl", "PRIVATE", "pub"),
			getter: secret.Secret.PrivateKey,
			want:   "PRIVATE",
		},
		{
			name:   "sshkey/public_key",
			secret: secret.NewSSHKey("test", "lbl", "priv", "PUBLIC"),
			getter: secret.Secret.PublicKey,
			want:   "PUBLIC",
		},
		{
			name:   "note/content",
			secret: secret.NewNote("test", "hello world"),
			getter: secret.Secret.Content,
			want:   "hello world",
		},
		{
			name:   "missing field returns empty",
			secret: secret.NewNote("test", "content"),
			getter: secret.Secret.URL,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.getter(tt.secret)
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOptionalFields(t *testing.T) {
	s := secret.NewPassword("test", "url", "user", "pass")
	s.Fields["totp_secret"] = "JBSWY3DPEHPK3PXP"
	s.Fields["notes"] = "some notes"

	if s.TOTPSecret() != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("totp_secret = %q, want %q", s.TOTPSecret(), "JBSWY3DPEHPK3PXP")
	}
	if s.Notes() != "some notes" {
		t.Fatalf("notes = %q, want %q", s.Notes(), "some notes")
	}
}

func TestSSHKeyPassphrase(t *testing.T) {
	s := secret.NewSSHKey("test", "lbl", "priv", "pub")
	s.Fields["passphrase"] = "my-passphrase"

	if s.Passphrase() != "my-passphrase" {
		t.Fatalf("passphrase = %q, want %q", s.Passphrase(), "my-passphrase")
	}
}

func TestIDUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for range 100 {
		s := secret.NewNote("test", "content")
		if seen[s.ID] {
			t.Fatalf("duplicate ID: %s", s.ID)
		}
		seen[s.ID] = true
	}
}

func TestIDFormat(t *testing.T) {
	for range 50 {
		s := secret.NewNote("test", "content")
		assertValidID(t, s.ID)
	}
}

func TestTags(t *testing.T) {
	s := secret.NewPassword("test", "url", "user", "pass")
	s.Tags = []string{"work", "github"}

	if len(s.Tags) != 2 {
		t.Fatalf("tags len = %d, want 2", len(s.Tags))
	}
	if s.Tags[0] != "work" || s.Tags[1] != "github" {
		t.Fatalf("tags = %v, want [work github]", s.Tags)
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

func assertTimestampsSet(t *testing.T, s secret.Secret) {
	t.Helper()
	if s.CreatedAt.IsZero() {
		t.Fatal("created_at is zero")
	}
	if s.UpdatedAt.IsZero() {
		t.Fatal("updated_at is zero")
	}
}
