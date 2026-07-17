package health

import (
	"strings"
	"testing"
)

func TestRedact(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		mustMask []string // substrings that must NOT survive
		mustKeep []string // substrings that must survive
	}{
		{
			name:     "url credentials",
			in:       "dial postgres://admin:s3cr3t@db.internal:5432/app failed",
			mustMask: []string{"s3cr3t"},
			mustKeep: []string{"admin", "db.internal"},
		},
		{
			name:     "password kv",
			in:       `connection failed password="hunter2" host=db`,
			mustMask: []string{"hunter2"},
			mustKeep: []string{"host=db"},
		},
		{
			name:     "json token field",
			in:       `{"token":"abcdef123456","user":"bob"}`,
			mustMask: []string{"abcdef123456"},
			mustKeep: []string{"bob"},
		},
		{
			name:     "bearer header",
			in:       "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			mustMask: []string{"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
		},
		{
			name:     "api_key kv",
			in:       "api_key=AKIAIOSFODNN7EXAMPLE region=us-east-1",
			mustMask: []string{"AKIAIOSFODNN7EXAMPLE"},
			mustKeep: []string{"region=us-east-1"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := redact(tc.in)
			for _, m := range tc.mustMask {
				if strings.Contains(got, m) {
					t.Errorf("secret %q leaked through redaction: %q", m, got)
				}
			}
			for _, k := range tc.mustKeep {
				if !strings.Contains(got, k) {
					t.Errorf("expected %q to survive redaction: %q", k, got)
				}
			}
		})
	}
}

func TestRedactTruncates(t *testing.T) {
	got := redact(strings.Repeat("x", maxVerboseLen+500))
	if len(got) <= maxVerboseLen {
		t.Fatalf("expected truncation marker, got len %d", len(got))
	}
	if !strings.HasSuffix(got, "(truncated)") {
		t.Errorf("expected truncation suffix, got tail %q", got[len(got)-20:])
	}
}

func TestHTMLTitle(t *testing.T) {
	equal(t, "title", htmlTitle([]byte("<html><head><TITLE> Sign in -  Okta </TITLE></head>")), "Sign in - Okta")
	equal(t, "no title", htmlTitle([]byte("<html></html>")), "")
}
