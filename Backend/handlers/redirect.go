package handlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
	"url_shortner/storage"

	"github.com/redis/go-redis/v9"
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

	// Serve from the Redis cache first. Cached entries expire when the link does (see ttl below),
	// so a cache hit is always a live link and needs no database query.
	longURL, err := storage.RDB.Get(r.Context(), shortId).Result()
	if err == nil {
		http.Redirect(w, r, longURL, http.StatusFound)
		return
	}
	if !errors.Is(err, redis.Nil) {
		log.Println("redirect: redis get failed:", err)
	}

	// Cache miss: look the link up in PostgreSQL.
	var expiresAt sql.NullTime
	err = storage.DB.QueryRowContext(r.Context(),
		"SELECT long_url, expires_at FROM short_urls WHERE short_id = $1", shortId,
	).Scan(&longURL, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "URL not Found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Println("redirect: database lookup failed:", err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	// Block expired links. The row stays in the database for monthly analysis.
	if expiresAt.Valid && time.Now().UTC().After(expiresAt.Time) {
		http.Error(w, "Link expired", http.StatusGone)
		return
	}

	// Guest links are cached only until they expire; signed-in users' links for 10 minutes.
	// A TTL of 0 would mean "never expire" in Redis, so only cache when there's time left.
	ttl := 10 * time.Minute
	if expiresAt.Valid {
		ttl = time.Until(expiresAt.Time)
	}
	if ttl > 0 {
		if err := storage.RDB.Set(r.Context(), shortId, longURL, ttl).Err(); err != nil {
			log.Println("redirect: redis set failed:", err)
		}
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}
