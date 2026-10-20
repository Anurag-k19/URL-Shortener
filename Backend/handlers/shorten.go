package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"url_shortner/auth"
	"url_shortner/storage"
	"url_shortner/utils"
)

// Request represents the JSON payload for URL shortening
type Request struct {
	URL string `json:"url"`
}

// shortURL builds the public short link, e.g. http://localhost:8000/fkI30S.
// Set BASE_URL in .env to your short domain in production.
func shortURL(id string) string {
	base := strings.TrimSuffix(os.Getenv("BASE_URL"), "/")
	if base == "" {
		base = "http://localhost:8000"
	}
	return base + "/" + id
}

// Shortenhandler handles POST requests to create short URLs with optional expiry for guests
func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Signed-in users come from the session cookie (auth.Middleware); everyone else is a guest
	userID := "guest"
	if id, ok := auth.UserID(r.Context()); ok {
		userID = strconv.FormatInt(id, 10)
	}

	// Parse JSON request body, capped at 4 KB so oversized payloads are rejected
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var req Request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	if !utils.ValidateURL(req.URL) {
		http.Error(w, "Invalid URL ", http.StatusBadRequest)
		return
	}

	// Reuse the user's existing short link for this URL if it hasn't expired
	var existingID string
	err = storage.DB.QueryRow(
		"SELECT short_id FROM short_urls WHERE user_id = $1 AND long_url = $2 AND (expires_at IS NULL OR expires_at > NOW())",
		userID, req.URL,
	).Scan(&existingID)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"short_url": shortURL(existingID)})
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		log.Println("shorten: checking for existing link failed:", err)
		http.Error(w, "Failed to store URL", http.StatusInternalServerError)
		return
	}

	// Generate a short ID that isn't taken yet. With 62^6 possible IDs a collision is very rare,
	// so a few attempts are plenty; a database error stops the request instead of skipping the check.
	var shortId string
	const maxAttempts = 5
	for attempt := 1; ; attempt++ {
		shortId = utils.GenerateShortId()
		var count int
		err = storage.DB.QueryRow("SELECT COUNT(*) FROM short_urls WHERE short_id = $1", shortId).Scan(&count)
		if err != nil {
			log.Println("shorten: checking short id failed:", err)
			http.Error(w, "Failed to store URL", http.StatusInternalServerError)
			return
		}
		if count == 0 {
			break
		}
		if attempt == maxAttempts {
			log.Println("shorten: no free short id after", maxAttempts, "attempts")
			http.Error(w, "Failed to store URL", http.StatusInternalServerError)
			return
		}
	}
	// Set 2-minute expiry for guest URLs, no expiry for registered users
	var expiry sql.NullTime

	if userID == "guest" {
		expiry = sql.NullTime{
			Time:  time.Now().UTC().Add(2 * time.Minute),
			Valid: true,
		}
	} else {
		expiry = sql.NullTime{Valid: false} // no expiry
	}

	var createdAt = time.Now().UTC()
	// Store URL in PostgreSQL database
	_, err = storage.DB.Exec(
		"INSERT INTO short_urls (short_id, long_url, user_id, expires_at, created_at) VALUES ($1, $2, $3, $4, $5)",
		shortId, req.URL, userID, expiry, createdAt,
	)
	if err != nil {
		http.Error(w, "Failed to store URL", http.StatusInternalServerError)
		return
	}

	fmt.Println("Stored:", shortId, "→", req.URL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"short_url": shortURL(shortId)})
	fmt.Println("Received url: ", req.URL)

}
