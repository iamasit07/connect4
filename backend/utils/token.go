package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken creates a cryptographically secure random token
func GenerateToken() string {
	bytes := make([]byte, 16) // 16 bytes = 128 bits
	rand.Read(bytes)
	return "tok_" + hex.EncodeToString(bytes)
}
