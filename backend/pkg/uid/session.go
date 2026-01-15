package uid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateSessionID generates a cryptographically secure random session ID
func GenerateSessionID() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session ID:  %v", err)
	}
	return hex.EncodeToString(bytes), nil
}
