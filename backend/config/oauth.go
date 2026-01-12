package config

import (
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleOAuthConfig *oauth2.Config

func LoadOAuthConfig() {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	log.Printf("[OAUTH] Loading Config -> ClientID: %s... (%d chars), Secret: %s... (%d chars), RedirectURL: %s",
		safePrefix(clientID, 5), len(clientID),
		safePrefix(clientSecret, 5), len(clientSecret),
		redirectURL)

	GoogleOAuthConfig = &oauth2.Config{
		RedirectURL:  redirectURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func safePrefix(s string, l int) string {
	if len(s) < l {
		return s
	}
	return s[:l]
}
