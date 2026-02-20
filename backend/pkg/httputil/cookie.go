package httputil

import (
	"errors"
	"net/http"

	"github.com/iamasit07/connect4/backend/internal/config"
)

const AuthCookieName = "auth_token"
const RefreshCookieName = "refresh_token"

func SetAccessCookie(w http.ResponseWriter, token string) {
	ttlMinutes := config.AppConfig.AccessTokenTTLMinutes
	maxAge := ttlMinutes * 60

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
		cookie.SameSite = http.SameSiteStrictMode
		cookie.Secure = true
	} else {
		cookie.SameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, cookie)
}

func SetRefreshCookie(w http.ResponseWriter, token string) {
	ttlDays := config.AppConfig.RefreshTokenTTLDays
	maxAge := ttlDays * 24 * 60 * 60

	isProduction := config.GetEnv("ENVIRONMENT", "development") == "production"

	cookie := &http.Cookie{
		Name:     RefreshCookieName,
		Value:    token,
		Path:     "/api/auth/refresh", // Only sent on refresh endpoint
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   isProduction,
	}

	if isProduction {
		cookie.SameSite = http.SameSiteStrictMode
		cookie.Secure = true
	} else {
		cookie.SameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, cookie)
}

// SetAuthCookie sets both access and refresh cookies
func SetAuthCookie(w http.ResponseWriter, accessToken string) {
	SetAccessCookie(w, accessToken)
}

// SetTokenPairCookies sets both access token and refresh token cookies
func SetTokenPairCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	SetAccessCookie(w, accessToken)
	SetRefreshCookie(w, refreshToken)
}

func ClearAuthCookie(w http.ResponseWriter) {
	isProduction := config.GetEnv("ENVIRONMENT", "development") == "production"

	// Clear access token cookie
	cookie := &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
	}

	if isProduction {
		cookie.SameSite = http.SameSiteStrictMode
	} else {
		cookie.SameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, cookie)
}

func ClearRefreshCookie(w http.ResponseWriter) {
	isProduction := config.GetEnv("ENVIRONMENT", "development") == "production"

	cookie := &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/api/auth/refresh",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
	}

	if isProduction {
		cookie.SameSite = http.SameSiteStrictMode
	} else {
		cookie.SameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, cookie)
}

// ClearAllAuthCookies clears both access and refresh token cookies
func ClearAllAuthCookies(w http.ResponseWriter) {
	ClearAuthCookie(w)
	ClearRefreshCookie(w)
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

func GetRefreshTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil {
		return "", errors.New("refresh cookie not found")
	}

	if cookie.Value == "" {
		return "", errors.New("refresh cookie is empty")
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
