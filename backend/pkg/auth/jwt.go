package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/iamasit07/4-in-a-row/backend/internal/config"
)

// Claims represents JWT claims with user information
type Claims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	SessionID string `json:"session_id"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT token for the user with session ID
func GenerateJWT(userID int64, username, sessionID string) (string, error) {
	// Use Centralized Config
	secret := config.AppConfig.JWTSecret
	// Hardcoded fallback for logic consistency, though Config handles default
	expirationHours := config.GetEnvAsInt("JWT_EXPIRATION_HOURS", 720) 

	claims := &Claims{
		UserID:    userID,
		Username:  username,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expirationHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*Claims, error) {
	// Use Centralized Config
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

// SetupClaims represents JWT claims for the setup phase
type SetupClaims struct {
	Email    string `json:"email"`
	GoogleID string `json:"google_id"`
	Name     string `json:"name"`
	jwt.RegisteredClaims
}

// GenerateSetupToken creates a short-lived token for the signup completion step
func GenerateSetupToken(email, googleID, name string) (string, error) {
	// Use Centralized Config
	secret := config.AppConfig.JWTSecret
	
	claims := &SetupClaims{
		Email:    email,
		GoogleID: googleID,
		Name:     name,
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
	// Use Centralized Config
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
