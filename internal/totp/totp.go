// Package totp generates time-based one-time passwords per RFC 6238.
package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

const (
	period = 30 // seconds
	digits = 6
)

// now is a function variable for testing.
var now = time.Now

// Generate returns a TOTP code and seconds remaining in the current period.
// The secret must be base32-encoded (standard TOTP format).
func Generate(secret string) (string, int, error) {
	key, err := decodeSecret(secret)
	if err != nil {
		return "", 0, fmt.Errorf("decode totp secret: %w", err)
	}

	t := now().Unix()
	counter := uint64(t / period)
	remaining := period - int(t%period)

	code := hotp(key, counter)
	return code, remaining, nil
}

// decodeSecret strips spaces and decodes a base32 string.
func decodeSecret(s string) ([]byte, error) {
	s = strings.ToUpper(strings.ReplaceAll(s, " ", ""))
	// pad to multiple of 8 for standard base32
	if pad := len(s) % 8; pad != 0 {
		s += strings.Repeat("=", 8-pad)
	}
	return base32.StdEncoding.DecodeString(s)
}

// hotp implements HOTP (RFC 4226) — HMAC-SHA1 with dynamic truncation.
func hotp(key []byte, counter uint64) string {
	// counter as 8-byte big-endian
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf[:])
	sum := mac.Sum(nil)

	// dynamic truncation
	offset := sum[len(sum)-1] & 0x0f
	code := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff

	return fmt.Sprintf("%0*d", digits, code%1000000)
}
