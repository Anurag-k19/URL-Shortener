package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"url_shortner/storage"
)

// RedirectHandler handles GET requests to redirect short URLs to their original destination
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// Extract short ID from URL path
	shortId := r.URL.Path[1:]

	if r.Method != "GET" {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	val, err := storage.RDB.Get(storage.CTX, shortId).Result()

	if err == nil {
		// cache hit
		http.Redirect(w, r, val, http.StatusFound)
		return
	}

	// Look up short URL in database
	var data struct {
		LongURL   string
		ExpiresAt sql.NullTime
	}
	row := storage.DB.QueryRow("SELECT long_url, expires_at FROM urls WHERE short_id = $1", shortId)
	err1 := row.Scan(&data.LongURL, &data.ExpiresAt)
	if err1 != nil {
		http.Error(w, "URL not Found", http.StatusNotFound)
		return
	}
	fmt.Println("[DEBUG] Short ID:", shortId, "| ExpiresAt Valid:", data.ExpiresAt.Valid, "| ExpiresAt Time:", data.ExpiresAt.Time, "| Now:", time.Now())

	// Check if URL has expired and delete if necessary
	if data.ExpiresAt.Valid && time.Now().UTC().After(data.ExpiresAt.Time) {
		storage.DB.Exec("DELETE FROM urls WHERE short_id = $1", shortId)
		http.Error(w, "Link expired", http.StatusGone)
		return
	}
	storage.RDB.Set(storage.CTX, shortId, data.LongURL, 10*time.Minute)

	http.Redirect(w, r, data.LongURL, http.StatusFound)

}
