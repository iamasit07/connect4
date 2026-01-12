package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iamasit07/4-in-a-row/backend/config"
)

// Claims represents JWT claims with user information
type Claims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	SessionID string `json:"session_id"` // Links JWT to active session
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT token for the user with session ID
func GenerateJWT(userID int64, username, sessionID string) (string, error) {
	// Get JWT secret and expiration from config
	secret := config.GetEnv("JWT_SECRET", "your-secret-key-change-this-in-production")
	expirationHours := config.GetEnvAsInt("JWT_EXPIRATION_HOURS", 720) // 30 days default

	// Create claims
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expirationHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*Claims, error) {
	secret := config.GetEnv("JWT_SECRET", "your-secret-key-change-this-in-production")

	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Extract claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
// SetupClaims represents JWT claims for the setup phase (OAuth -> Complete Signup)
type SetupClaims struct {
	Email    string `json:"email"`
	GoogleID string `json:"google_id"`
	jwt.RegisteredClaims
}

// GenerateSetupToken creates a short-lived token for the signup completion step
func GenerateSetupToken(email, googleID string) (string, error) {
	secret := config.GetEnv("JWT_SECRET", "your-secret-key-change-this-in-production")
	
	claims := &SetupClaims{
		Email:    email,
		GoogleID: googleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 mins to complete signup
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateSetupToken validates the setup token
func ValidateSetupToken(tokenString string) (*SetupClaims, error) {
	secret := config.GetEnv("JWT_SECRET", "your-secret-key-change-this-in-production")

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

// ValidateJWTWithSession validates JWT and checks if the session is still active
func ValidateJWTWithSession(tokenString string) (*Claims, error) {
	// First validate the JWT itself
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return nil, err
	}

	// Check if session is still active
	session, err := GetSession(claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session validation failed: %v", err)
	}

	if session == nil || !session.IsActive {
		return nil, errors.New("session invalidated")
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session expired")
	}

	return claims, nil
}
