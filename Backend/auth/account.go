package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// MeHandler returns the signed-in user as JSON, or 401 for guests.
func MeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := UserID(r.Context())
	if !ok {
		http.Error(w, "Not signed in", http.StatusUnauthorized)
		return
	}

	u, err := findUser(r.Context(), userID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Not signed in", http.StatusUnauthorized)
		return
	}
	if err != nil {
		log.Println("auth: loading user failed:", err)
		http.Error(w, "Could not load your account", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// LogoutHandler deletes the session from Redis and clears the session cookie.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	if cookie, err := r.Cookie(sessionCookieName); err == nil && cookie.Value != "" {
		if err := deleteSession(r.Context(), cookie.Value); err != nil {
			log.Println("auth: deleting session failed:", err)
		}
	}
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}
