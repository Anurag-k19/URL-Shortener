package utils

import (
	"net/http"
	"net/url"
	"time"
)

func ValidateURL(rawUrl string) bool {

	parsed, err := url.ParseRequestURI(rawUrl)

	if err != nil {
		return false
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	return true
}

func Isreachable(url string) bool {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Head(url)

	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}
