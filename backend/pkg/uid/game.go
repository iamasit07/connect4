package uid

import (
	"crypto/rand"
	"encoding/hex"
)

// randonmly generates a unique game ID
// copied from the internet
func GenerateGameID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
