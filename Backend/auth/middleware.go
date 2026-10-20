package auth

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type contextKey struct{}

var userIDKey = contextKey{}

// Middleware attaches the signed-in user's ID to the request context when the request
// carries a valid session cookie. Requests without one continue as guests.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err == nil && cookie.Value != "" {
			userID, err := lookupSession(r.Context(), cookie.Value)
			switch {
			case err == nil:
				r = r.WithContext(context.WithValue(r.Context(), userIDKey, userID))
			case !errors.Is(err, redis.Nil):
				log.Println("auth: session lookup failed:", err)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// UserID returns the signed-in user's ID from the request context, or false for guests.
func UserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}
