package handlers

import (
	"net/http"
	"time"
	"url_shortner/storage"
	"url_shortner/utils"
)

// RedirectHandler handles GET requests to redirect short URLs to their original destination
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// Extract short ID from URL path
	shortId := r.URL.Path[1:]

	if r.Method != "GET" {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	storage.Mu.Lock()
	defer storage.Mu.Unlock()
	// Look up short URL in storage
	data, exists := storage.URLmap[shortId]

	if !exists {
		http.Error(w, "URL not Found", http.StatusNotFound)
		return
	}
	// Check if URL has expired and delete if necessary
	if !data.ExpiresAt.IsZero() && time.Now().After(data.ExpiresAt) {
		delete(storage.URLmap, shortId)
		http.Error(w, "Link expired", http.StatusGone)
		return
	}

	if !utils.Isreachable(data.LongURL) {

		http.Error(w, "Invalid URL: Not Reachable", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, data.LongURL, http.StatusFound)

}
