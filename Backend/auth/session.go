package auth

import (
	"context"
	"crypto/rand"
	"net/http"
	"time"
	"url_shortner/storage"
)

const (
	sessionCookieName = "session"
	sessionKeyPrefix  = "session:"
	sessionTTL        = 7 * 24 * time.Hour
)

// createSession stores a new session for userID in Redis and returns its random ID.
// Redis deletes the session by itself when sessionTTL runs out.
func createSession(ctx context.Context, userID int64) (string, error) {
	sessionID := rand.Text()
	if err := storage.RDB.Set(ctx, sessionKeyPrefix+sessionID, userID, sessionTTL).Err(); err != nil {
		return "", err
	}
	return sessionID, nil
}

// setSessionCookie sends the session ID to the browser in a cookie JavaScript can't read.
func setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		Secure:   secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})
}

// lookupSession returns the user ID stored for sessionID.
// It returns redis.Nil when the session doesn't exist or has expired.
// It's a variable so tests can replace it without a running Redis.
var lookupSession = func(ctx context.Context, sessionID string) (int64, error) {
	return storage.RDB.Get(ctx, sessionKeyPrefix+sessionID).Int64()
}

// deleteSession removes sessionID from Redis, which signs that browser out immediately.
var deleteSession = func(ctx context.Context, sessionID string) error {
	return storage.RDB.Del(ctx, sessionKeyPrefix+sessionID).Err()
}

// clearSessionCookie tells the browser to delete the session cookie.
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})
}
