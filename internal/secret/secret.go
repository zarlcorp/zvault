// Package secret defines the typed secret model for zvault.
package secret

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Type identifies the kind of secret.
type Type string

const (
	TypePassword Type = "password"
	TypeAPIKey   Type = "apikey"
	TypeSSHKey   Type = "sshkey"
	TypeNote     Type = "note"
)

// Secret holds an encrypted secret with type-specific fields.
type Secret struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      Type              `json:"type"`
	Fields    map[string]string `json:"fields"`
	Tags      []string          `json:"tags"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// NewPassword creates a password secret.
func NewPassword(name, url, username, password string) Secret {
	now := time.Now()
	return Secret{
		ID:   generateID(),
		Name: name,
		Type: TypePassword,
		Fields: map[string]string{
			"url":      url,
			"username": username,
			"password": password,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewAPIKey creates an API key secret.
func NewAPIKey(name, service, key string) Secret {
	now := time.Now()
	return Secret{
		ID:   generateID(),
		Name: name,
		Type: TypeAPIKey,
		Fields: map[string]string{
			"service": service,
			"key":     key,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewSSHKey creates an SSH key secret.
func NewSSHKey(name, label, privateKey, publicKey string) Secret {
	now := time.Now()
	return Secret{
		ID:   generateID(),
		Name: name,
		Type: TypeSSHKey,
		Fields: map[string]string{
			"label":       label,
			"private_key": privateKey,
			"public_key":  publicKey,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewNote creates a note secret.
func NewNote(name, content string) Secret {
	now := time.Now()
	return Secret{
		ID:   generateID(),
		Name: name,
		Type: TypeNote,
		Fields: map[string]string{
			"content": content,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// field returns a field value or empty string.
func (s Secret) field(key string) string {
	return s.Fields[key]
}

// URL returns the url field (password type).
func (s Secret) URL() string { return s.field("url") }

// Username returns the username field (password type).
func (s Secret) Username() string { return s.field("username") }

// Password returns the password field (password type).
func (s Secret) Password() string { return s.field("password") }

// TOTPSecret returns the totp_secret field (password type).
func (s Secret) TOTPSecret() string { return s.field("totp_secret") }

// Notes returns the notes field.
func (s Secret) Notes() string { return s.field("notes") }

// Service returns the service field (apikey type).
func (s Secret) Service() string { return s.field("service") }

// Key returns the key field (apikey type).
func (s Secret) Key() string { return s.field("key") }

// Label returns the label field (sshkey type).
func (s Secret) Label() string { return s.field("label") }

// PrivateKey returns the private_key field (sshkey type).
func (s Secret) PrivateKey() string { return s.field("private_key") }

// PublicKey returns the public_key field (sshkey type).
func (s Secret) PublicKey() string { return s.field("public_key") }

// Passphrase returns the passphrase field (sshkey type).
func (s Secret) Passphrase() string { return s.field("passphrase") }

// Content returns the content field (note type).
func (s Secret) Content() string { return s.field("content") }

// generateID returns an 8-character hex string from 4 random bytes.
func generateID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand: %v", err))
	}
	return hex.EncodeToString(b)
}
