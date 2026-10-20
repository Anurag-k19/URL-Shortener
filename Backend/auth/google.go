package auth

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const googleIssuer = "https://accounts.google.com"

var (
	oauthConfig *oauth2.Config
	verifier    *oidc.IDTokenVerifier
	frontendURL string
)

// InitGoogle sets up the Google OAuth config and ID token verifier from env vars.
func InitGoogle(ctx context.Context) error {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	frontendURL = os.Getenv("FRONTEND_URL")

	if clientID == "" || clientSecret == "" || redirectURL == "" || frontendURL == "" {
		return errors.New("GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_REDIRECT_URL and FRONTEND_URL must be set in .env")
	}

	provider, err := oidc.NewProvider(ctx, googleIssuer)
	if err != nil {
		return fmt.Errorf("connecting to Google: %w", err)
	}

	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}

	verifier = provider.Verifier(&oidc.Config{ClientID: clientID})

	return nil
}
