package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
	"url_shortner/storage"
)

// RedirectHandler handles GET requests to redirect short URLs to their original destination
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	shortId := r.URL.Query().Get("code")
	if shortId == "" {
		shortId = strings.TrimPrefix(r.URL.Path, "/")
	}
	if shortId == "" || shortId == "redirect" {
		http.Error(w, "Missing code query parameter", http.StatusBadRequest)
		return
	}

	// Look up short URL in database first so expiry rules are enforced before any cache is used.
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

	// Check if URL has expired and delete it from the database and cache if necessary.
	if data.ExpiresAt.Valid && time.Now().UTC().After(data.ExpiresAt.Time) {
		storage.DB.Exec("DELETE FROM urls WHERE short_id = $1", shortId)
		storage.RDB.Del(storage.CTX, shortId)
		http.Error(w, "Link expired", http.StatusGone)
		return
	}

	if val, err := storage.RDB.Get(storage.CTX, shortId).Result(); err == nil {
		// cache hit for a non-expired link
		http.Redirect(w, r, val, http.StatusFound)
		return
	}

	var ttl time.Duration
	if data.ExpiresAt.Valid {
		ttl = time.Until(data.ExpiresAt.Time)
		if ttl < 0 {
			ttl = 0
		}
	} else {
		ttl = 10 * time.Minute
	}

	storage.RDB.Set(storage.CTX, shortId, data.LongURL, ttl)
	http.Redirect(w, r, data.LongURL, http.StatusFound)

}
