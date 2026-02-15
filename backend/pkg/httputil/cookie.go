package httputil

import (
	"errors"
	"net/http"

	"github.com/iamasit07/4-in-a-row/backend/internal/config"
)

const AuthCookieName = "auth_token"

func SetAuthCookie(w http.ResponseWriter, token string) {
	expirationHours := config.GetEnvAsInt("JWT_EXPIRATION_HOURS", 720)
	maxAge := expirationHours * 60 * 60

	isProduction := config.GetEnv("ENVIRONMENT", "development") == "production"

	cookie := &http.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   isProduction,
	}

	if isProduction {
		cookie.SameSite = http.SameSiteNoneMode
		cookie.Secure = true
	} else {
		cookie.SameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, cookie)
}

func ClearAuthCookie(w http.ResponseWriter) {
	isProduction := config.GetEnv("ENVIRONMENT", "development") == "production"

	cookie := &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
	}

	if isProduction {
		cookie.SameSite = http.SameSiteNoneMode
	} else {
		cookie.SameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, cookie)
}

func GetTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(AuthCookieName)
	if err != nil {
		return "", errors.New("auth cookie not found")
	}

	if cookie.Value == "" {
		return "", errors.New("auth cookie is empty")
	}

	return cookie.Value, nil
}

func GetTokenFromRequest(r *http.Request) (string, error) {
	token, err := GetTokenFromCookie(r)
	if err == nil && token != "" {
		return token, nil
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			return authHeader[7:], nil
		}
		return authHeader, nil
	}

	return "", errors.New("no auth token found in cookie or header")
}
