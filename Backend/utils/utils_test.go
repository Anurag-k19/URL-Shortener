package utils

import (
	"net/http"
	"net/http/httptest"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateURL(tt.url); got != tt.want {
				t.Errorf("ValidateURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestIsreachable(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{"200 ok", http.StatusOK, true},
		{"301 redirect", http.StatusMovedPermanently, true},
		{"404 not found", http.StatusNotFound, false},
		{"500 server error", http.StatusInternalServerError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			if got := Isreachable(srv.URL); got != tt.want {
				t.Errorf("Isreachable() with status %d = %v, want %v", tt.status, got, tt.want)
			}
		})
	}

	t.Run("unreachable host", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		url := srv.URL
		srv.Close()

		if Isreachable(url) {
			t.Errorf("Isreachable(%q) = true for a closed server, want false", url)
		}
	})
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
