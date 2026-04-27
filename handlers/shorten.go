package handlers

import (
	"database/sql"
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

	// Check if user already shortened this URL
	var existingID string
	err = storage.DB.QueryRow(
		"SELECT short_id from urls WHERE user_id =$1 AND long_url=$2",
		userID, req.URL,
	).Scan(&existingID)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"short_url": "http://localhost:8000/" + existingID})
		return
	}

	// Generate unique short ID by checking for collisions in database
	var shortId string
	for {
		shortId = utils.GenerateShortId()
		var count int
		err = storage.DB.QueryRow("SELECT COUNT(*) FROM urls WHERE short_id = $1", shortId).Scan(&count)
		if err != nil || count == 0 {
			break
		}
	}
	// Set 5-minute expiry for guest URLs, no expiry for registered users
	var expiry sql.NullTime

	if userID == "guest" {
		expiry = sql.NullTime{
			Time:  time.Now().UTC().Add(1 * time.Minute),
			Valid: true,
		}
	} else {
		expiry = sql.NullTime{Valid: false} // no expiry
	}

	var createdAt = time.Now().UTC()
	// Store URL in PostgreSQL database
	_, err = storage.DB.Exec(
		"INSERT INTO urls (short_id, long_url, user_id, expires_at, created_at) VALUES ($1, $2, $3, $4, $5)",
		shortId, req.URL, userID, expiry, createdAt,
	)
	if err != nil {
		http.Error(w, "Failed to store URL", http.StatusInternalServerError)
		return
	}

	fmt.Println("Stored:", shortId, "→", req.URL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"short_url": "http://localhost:8000/" + shortId})
	fmt.Println("Received url: ", req.URL)

}
