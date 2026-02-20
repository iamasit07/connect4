package auth

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14) // Cost 14 is secure
	return string(bytes), err
}

// CheckPasswordHash checks if a password matches a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength enforces password complexity rules:
// - Minimum 8 characters
// - At least 1 uppercase letter
// - At least 1 lowercase letter
// - At least 1 digit
// - At least 1 special character
func ValidatePasswordStrength(password string) error {
	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	var failures []string

	if len(password) < 8 {
		failures = append(failures, "at least 8 characters")
	}
	if !hasUpper {
		failures = append(failures, "at least 1 uppercase letter")
	}
	if !hasLower {
		failures = append(failures, "at least 1 lowercase letter")
	}
	if !hasDigit {
		failures = append(failures, "at least 1 digit")
	}
	if !hasSpecial {
		failures = append(failures, "at least 1 special character")
	}

	if len(failures) > 0 {
		return fmt.Errorf("password must contain %s", strings.Join(failures, ", "))
	}

	return nil
}
