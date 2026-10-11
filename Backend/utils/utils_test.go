package utils

import (
	"strings"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"http url", "http://example.com", true},
		{"https url with path and query", "https://example.com/a/b?q=1", true},
		{"empty string", "", false},
		{"no scheme", "example.com", false},
		{"relative path", "/just/a/path", false},
		{"ftp scheme", "ftp://example.com/file", false},
		{"javascript scheme", "javascript:alert(1)", false},
		{"plain text", "not a url", false},
		{"google sheets", "https://docs.google.com/spreadsheets/d/abc", true},
		{"uppercase scheme", "HTTPS://EXAMPLE.COM", true},
		{"script in path is just text", "https://example.com/<script>", true},
		{"data url", "data:text/html,<script>alert(1)</script>", false},
		{"file url", "file:///etc/passwd", false},
		{"mailto", "mailto:a@b.com", false},
		{"no host", "http://", false},
		{"empty host with path", "https:///path", false},
		{"userinfo trick", "https://google.com@evil.com", false},
		{"max length", "https://example.com/" + strings.Repeat("a", 2048-len("https://example.com/")), true},
		{"too long", "https://example.com/" + strings.Repeat("a", 3000), false},
		{"own short link", "http://localhost:8000/redirect?code=abc", false},
		{"own short link by ip", "http://127.0.0.1:8000/abc", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateURL(tt.url); got != tt.want {
				t.Errorf("ValidateURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestGenerateShortId(t *testing.T) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	allowed := make(map[rune]bool, len(charset))
	for _, c := range charset {
		allowed[c] = true
	}

	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateShortId()

		if len(id) != 6 {
			t.Fatalf("GenerateShortId() = %q, want length 6", id)
		}
		for _, c := range id {
			if !allowed[c] {
				t.Fatalf("GenerateShortId() = %q, contains invalid character %q", id, c)
			}
		}
		seen[id] = true
	}

	// 62^6 possible ids, so 1000 draws should essentially never collide.
	if len(seen) < 995 {
		t.Errorf("GenerateShortId() produced only %d unique ids out of 1000", len(seen))
	}
}
