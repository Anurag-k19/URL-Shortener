package utils

import (
	"net/url"
	"strings"
)

const maxURLLength = 2048

// ownHosts are the hosts that serve our own short links; shortening them would create chains or loops.
var ownHosts = map[string]bool{"localhost": true, "127.0.0.1": true}

func ValidateURL(rawURL string) bool {
	if rawURL == "" || len(rawURL) > maxURLLength {
		return false
	}

	parsed, err := url.ParseRequestURI(rawURL)

	if err != nil {
		return false
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	if parsed.Hostname() == "" {
		return false
	}

	// Reject userinfo like https://google.com@evil.com, which hides the real destination.
	if parsed.User != nil {
		return false
	}

	if ownHosts[strings.ToLower(parsed.Hostname())] {
		return false
	}

	return true
}
