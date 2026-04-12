package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"url_shortner/storage"
	"url_shortner/utils"
)

// Request represents the JSON payload for URL shortening
type Request struct {
	URL string `json:"url"`
}

// Shortenhandler handles POST requests to create short URLs with optional expiry for guests
func Shortenhandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from header, default to "guest" if not provided
	userID := r.Header.Get("X-User-ID")

	if userID == "" {
		userID = "guest"
	}

	// Parse JSON request body
	var req Request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if !utils.ValidateURL(req.URL) {
		http.Error(w, "Invalid URL ", http.StatusBadRequest)
		return
	}

	if !utils.Isreachable(req.URL) {
		http.Error(w, "Invalid URL: Not Reachable", http.StatusBadRequest)
		return
	}

	storage.Mu.Lock()
	defer storage.Mu.Unlock()
	// Create unique key combining userID and original URL
	key := userID + ":" + req.URL
	// Return existing short URL if user already shortened this URL
	if existingId, exists := storage.Reversemap[key]; exists {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"short_url": "http://localhost:8000/" + existingId})
		return
	}

	// Generate unique short ID by checking for collisions
	var shortId string
	for {
		shortId = utils.GenerateShortId()
		if _, exists := storage.URLmap[shortId]; !exists {
			break
		}

	}
	// Set 5-minute expiry for guest URLs, no expiry for registered users
	var expiry time.Time

	if userID == "guest" {
		expiry = time.Now().Add(5 * time.Minute)
	} else {
		expiry = time.Time{} // no expiry
	}

	// Store the URL mapping and reverse mapping for future lookups
	storage.URLmap[shortId] = storage.URLData{
		LongURL:   req.URL,
		ExpiresAt: expiry,
	}
	storage.Reversemap[key] = shortId
	fmt.Println("Stored:", shortId, "→", req.URL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"short_url": "http://localhost:8000/" + shortId})
	fmt.Println("Received url: ", req.URL)

}
