package auth

import (
	"crypto/subtle"
	"log"
	"net/http"
)

// GoogleCallbackHandler finishes Google sign-in: it checks the state, swaps the code for an
// ID token, verifies it, saves the user, starts a session and sends the browser to the frontend.
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()

	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil || query.Get("state") == "" ||
		subtle.ConstantTimeCompare([]byte(stateCookie.Value), []byte(query.Get("state"))) != 1 {
		http.Error(w, "Invalid login state, please try signing in again", http.StatusBadRequest)
		return
	}
	clearStateCookie(w)

	if query.Get("error") != "" {
		http.Error(w, "Google sign-in was cancelled", http.StatusUnauthorized)
		return
	}

	code := query.Get("code")
	if code == "" {
		http.Error(w, "Missing code from Google", http.StatusBadRequest)
		return
	}

	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Println("google: code exchange failed:", err)
		http.Error(w, "Could not complete Google sign-in", http.StatusBadGateway)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Println("google: token response had no id_token")
		http.Error(w, "Could not complete Google sign-in", http.StatusBadGateway)
		return
	}

	idToken, err := verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		log.Println("google: id token verification failed:", err)
		http.Error(w, "Could not verify Google sign-in", http.StatusUnauthorized)
		return
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		log.Println("google: reading id token claims failed:", err)
		http.Error(w, "Could not complete Google sign-in", http.StatusBadGateway)
		return
	}
	if claims.Email == "" || !claims.EmailVerified {
		http.Error(w, "Your Google email address is not verified", http.StatusForbidden)
		return
	}

	userID, err := upsertUser(r.Context(), idToken.Subject, claims.Email, claims.Name)
	if err != nil {
		log.Println("auth: saving user failed:", err)
		http.Error(w, "Could not save your account", http.StatusInternalServerError)
		return
	}

	sessionID, err := createSession(r.Context(), userID)
	if err != nil {
		log.Println("auth: creating session failed:", err)
		http.Error(w, "Could not start your session", http.StatusInternalServerError)
		return
	}
	setSessionCookie(w, sessionID)

	http.Redirect(w, r, frontendURL, http.StatusFound)
}

// clearStateCookie deletes the oauth_state cookie, so each state value can only be used once.
func clearStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		Path:     "/auth/google",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secureCookies(),
		SameSite: http.SameSiteLaxMode,
	})
}
