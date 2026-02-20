package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iamasit07/connect4/backend/internal/config"
)

// Claims represents JWT claims for access tokens
type Claims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// RefreshClaims represents JWT claims for refresh tokens
type RefreshClaims struct {
	UserID    int64  `json:"user_id"`
	SessionID string `json:"session_id"`
	TokenID   string `json:"token_id"` // unique ID to track/revoke this specific refresh token
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a short-lived JWT access token
func GenerateAccessToken(userID int64, username, sessionID string) (string, error) {
	secret := config.AppConfig.JWTSecret
	ttl := time.Duration(config.AppConfig.AccessTokenTTLMinutes) * time.Minute

	claims := &Claims{
		UserID:    userID,
		Username:  username,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateAccessToken validates a JWT access token and returns the claims
func ValidateAccessToken(tokenString string) (*Claims, error) {
	secret := config.AppConfig.JWTSecret

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateRefreshToken creates a long-lived JWT refresh token
func GenerateRefreshToken(userID int64, sessionID, tokenID string) (string, error) {
	secret := config.AppConfig.JWTSecret
	ttl := time.Duration(config.AppConfig.RefreshTokenTTLDays) * 24 * time.Hour

	claims := &RefreshClaims{
		UserID:    userID,
		SessionID: sessionID,
		TokenID:   tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateRefreshToken validates a refresh token JWT and returns its claims
func ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	secret := config.AppConfig.JWTSecret

	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid refresh token")
}

// --- Backward-compatible aliases ---

// GenerateJWT is an alias for GenerateAccessToken (backward compatibility)
func GenerateJWT(userID int64, username, sessionID string) (string, error) {
	return GenerateAccessToken(userID, username, sessionID)
}

// ValidateJWT is an alias for ValidateAccessToken (backward compatibility)
func ValidateJWT(tokenString string) (*Claims, error) {
	return ValidateAccessToken(tokenString)
}

// --- Setup Token (unchanged) ---

// SetupClaims represents JWT claims for the setup phase
type SetupClaims struct {
	Email    string `json:"email"`
	GoogleID string `json:"google_id"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	jwt.RegisteredClaims
}

// GenerateSetupToken creates a short-lived token for the signup completion step
func GenerateSetupToken(email, googleID, name, picture string) (string, error) {
	secret := config.AppConfig.JWTSecret

	claims := &SetupClaims{
		Email:    email,
		GoogleID: googleID,
		Name:     name,
		Picture:  picture,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateSetupToken validates the setup token
func ValidateSetupToken(tokenString string) (*SetupClaims, error) {
	secret := config.AppConfig.JWTSecret

	token, err := jwt.ParseWithClaims(tokenString, &SetupClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*SetupClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
