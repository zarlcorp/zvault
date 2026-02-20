package totp

import (
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	// JBSWY3DPEHPK3PXP decodes to "Hello!" — a common TOTP test secret
	tests := []struct {
		name      string
		secret    string
		time      time.Time
		wantCode  string
		wantRemOk bool
	}{
		{
			name:      "known secret at epoch 0",
			secret:    "JBSWY3DPEHPK3PXP",
			time:      time.Unix(0, 0),
			wantCode:  "282760",
			wantRemOk: true,
		},
		{
			name:      "known secret at t=30",
			secret:    "JBSWY3DPEHPK3PXP",
			time:      time.Unix(30, 0),
			wantCode:  "996554",
			wantRemOk: true,
		},
		{
			name:      "known secret at t=60",
			secret:    "JBSWY3DPEHPK3PXP",
			time:      time.Unix(60, 0),
			wantCode:  "602287",
			wantRemOk: true,
		},
		{
			name:      "with spaces in secret",
			secret:    "JBSW Y3DP EHPK 3PXP",
			time:      time.Unix(0, 0),
			wantCode:  "282760",
			wantRemOk: true,
		},
		{
			name:      "lowercase secret",
			secret:    "jbswy3dpehpk3pxp",
			time:      time.Unix(0, 0),
			wantCode:  "282760",
			wantRemOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now = func() time.Time { return tt.time }
			t.Cleanup(func() { now = time.Now })

			code, remaining, err := Generate(tt.secret)
			if err != nil {
				t.Fatalf("Generate() error: %v", err)
			}
			if code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
			if remaining < 1 || remaining > 30 {
				t.Errorf("remaining = %d, want 1-30", remaining)
			}
		})
	}
}

func TestGenerateRemaining(t *testing.T) {
	tests := []struct {
		unix          int64
		wantRemaining int
	}{
		{0, 30},
		{1, 29},
		{15, 15},
		{29, 1},
		{30, 30},
		{31, 29},
	}

	for _, tt := range tests {
		now = func() time.Time { return time.Unix(tt.unix, 0) }
		_, remaining, err := Generate("JBSWY3DPEHPK3PXP")
		if err != nil {
			t.Fatalf("Generate() error: %v", err)
		}
		if remaining != tt.wantRemaining {
			t.Errorf("at t=%d: remaining = %d, want %d", tt.unix, remaining, tt.wantRemaining)
		}
	}
	now = time.Now
}

func TestGenerateInvalidSecret(t *testing.T) {
	_, _, err := Generate("not-valid-base32!!!")
	if err == nil {
		t.Error("expected error for invalid base32 secret")
	}
}

func TestGenerateSixDigitPadding(t *testing.T) {
	// ensure codes are zero-padded to 6 digits
	now = func() time.Time { return time.Unix(0, 0) }
	defer func() { now = time.Now }()

	code, _, err := Generate("JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 6 {
		t.Errorf("code length = %d, want 6", len(code))
	}
}

// RFC 6238 test vectors use SHA1 with the 20-byte ASCII key
// "12345678901234567890"
func TestRFC6238Vectors(t *testing.T) {
	// base32 of "12345678901234567890"
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

	tests := []struct {
		time int64
		want string
	}{
		{59, "287082"},
		{1111111109, "081804"},
		{1111111111, "050471"},
		{1234567890, "005924"},
		{2000000000, "279037"},
	}

	for _, tt := range tests {
		now = func() time.Time { return time.Unix(tt.time, 0) }
		code, _, err := Generate(secret)
		if err != nil {
			t.Fatalf("at t=%d: %v", tt.time, err)
		}
		if code != tt.want {
			t.Errorf("at t=%d: code = %q, want %q", tt.time, code, tt.want)
		}
	}
	now = time.Now
}
