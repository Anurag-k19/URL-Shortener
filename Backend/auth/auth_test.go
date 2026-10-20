package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
)

// fakeSessions replaces the Redis lookups for one test and restores them afterwards.
func fakeSessions(t *testing.T, sessions map[string]int64) *[]string {
	t.Helper()
	oldLookup, oldDelete, oldConfig := lookupSession, deleteSession, oauthConfig
	t.Cleanup(func() { lookupSession, deleteSession, oauthConfig = oldLookup, oldDelete, oldConfig })

	oauthConfig = &oauth2.Config{RedirectURL: "http://localhost:8000/auth/google/callback"}
	deleted := &[]string{}
	lookupSession = func(_ context.Context, sessionID string) (int64, error) {
		if userID, ok := sessions[sessionID]; ok {
			return userID, nil
		}
		return 0, redis.Nil
	}
	deleteSession = func(_ context.Context, sessionID string) error {
		*deleted = append(*deleted, sessionID)
		delete(sessions, sessionID)
		return nil
	}
	return deleted
}

// userIDSeenBy runs req through Middleware and reports what UserID the next handler saw.
func userIDSeenBy(req *http.Request) (int64, bool) {
	var gotID int64
	var gotOK bool
	Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, gotOK = UserID(r.Context())
	})).ServeHTTP(httptest.NewRecorder(), req)
	return gotID, gotOK
}

func TestMiddleware(t *testing.T) {
	fakeSessions(t, map[string]int64{"valid-session": 42})

	tests := []struct {
		name   string
		cookie *http.Cookie
		wantID int64
		wantOK bool
	}{
		{"no cookie is a guest", nil, 0, false},
		{"valid session is signed in", &http.Cookie{Name: sessionCookieName, Value: "valid-session"}, 42, true},
		{"unknown session is a guest", &http.Cookie{Name: sessionCookieName, Value: "made-up"}, 0, false},
		{"empty cookie is a guest", &http.Cookie{Name: sessionCookieName, Value: ""}, 0, false},
		{"X-User-ID header is ignored", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("X-User-ID", "7")
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			gotID, gotOK := userIDSeenBy(req)
			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Errorf("UserID = (%d, %v), want (%d, %v)", gotID, gotOK, tt.wantID, tt.wantOK)
			}
		})
	}
}

func TestMeHandler(t *testing.T) {
	oldFind := findUser
	t.Cleanup(func() { findUser = oldFind })
	findUser = func(_ context.Context, userID int64) (user, error) {
		if userID == 42 {
			return user{ID: 42, Email: "anurag@example.com", Name: "Anurag"}, nil
		}
		if userID == 99 {
			return user{}, errors.New("database down")
		}
		return user{}, sql.ErrNoRows
	}

	signedInAs := func(userID int64) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		return req.WithContext(context.WithValue(req.Context(), userIDKey, userID))
	}

	t.Run("guest gets 401", func(t *testing.T) {
		rec := httptest.NewRecorder()
		MeHandler(rec, httptest.NewRequest(http.MethodGet, "/auth/me", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", rec.Code)
		}
	})

	t.Run("signed-in user gets their profile", func(t *testing.T) {
		rec := httptest.NewRecorder()
		MeHandler(rec, signedInAs(42))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		var got user
		if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		want := user{ID: 42, Email: "anurag@example.com", Name: "Anurag"}
		if got != want {
			t.Errorf("body = %+v, want %+v", got, want)
		}
	})

	t.Run("deleted user gets 401", func(t *testing.T) {
		rec := httptest.NewRecorder()
		MeHandler(rec, signedInAs(7))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", rec.Code)
		}
	})

	t.Run("database error gets 500", func(t *testing.T) {
		rec := httptest.NewRecorder()
		MeHandler(rec, signedInAs(99))
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", rec.Code)
		}
	})

	t.Run("POST is not allowed", func(t *testing.T) {
		rec := httptest.NewRecorder()
		MeHandler(rec, httptest.NewRequest(http.MethodPost, "/auth/me", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want 405", rec.Code)
		}
	})
}

func TestLogoutHandler(t *testing.T) {
	t.Run("deletes the session and clears the cookie", func(t *testing.T) {
		sessions := map[string]int64{"valid-session": 42}
		deleted := fakeSessions(t, sessions)

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
		rec := httptest.NewRecorder()
		LogoutHandler(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want 204", rec.Code)
		}
		if len(*deleted) != 1 || (*deleted)[0] != "valid-session" {
			t.Errorf("deleted sessions = %v, want [valid-session]", *deleted)
		}
		cookies := rec.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Name != sessionCookieName || cookies[0].MaxAge >= 0 {
			t.Errorf("cookies = %v, want the session cookie cleared", cookies)
		}

		// The old cookie no longer signs anyone in.
		again := httptest.NewRequest(http.MethodGet, "/", nil)
		again.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
		if _, ok := userIDSeenBy(again); ok {
			t.Error("session still valid after logout")
		}
	})

	t.Run("works without a cookie", func(t *testing.T) {
		deleted := fakeSessions(t, map[string]int64{})
		rec := httptest.NewRecorder()
		LogoutHandler(rec, httptest.NewRequest(http.MethodPost, "/auth/logout", nil))
		if rec.Code != http.StatusNoContent || len(*deleted) != 0 {
			t.Errorf("status = %d, deleted = %v; want 204 and nothing deleted", rec.Code, *deleted)
		}
	})

	t.Run("GET is not allowed", func(t *testing.T) {
		fakeSessions(t, map[string]int64{})
		rec := httptest.NewRecorder()
		LogoutHandler(rec, httptest.NewRequest(http.MethodGet, "/auth/logout", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want 405", rec.Code)
		}
	})
}
