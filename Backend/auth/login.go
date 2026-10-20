package auth

import (
	"crypto/rand"
	"net/http"
	"strings"
)

const stateCookieName = "oauth_state"

// GoogleLoginHandler sends the user to Google's sign-in page.
// The random state is kept in a short-lived cookie so the callback can check
// that the login was started by this browser (CSRF protection).
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	state := rand.Text()

	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/auth/google",
		MaxAge:   300,
		HttpOnly: true,
		Secure:   secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, oauthConfig.AuthCodeURL(state), http.StatusFound)
}

// secureCookies reports whether cookies should be HTTPS-only, based on the redirect URL's scheme.
func secureCookies() bool {
	return strings.HasPrefix(oauthConfig.RedirectURL, "https://")
}
