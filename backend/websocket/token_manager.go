package websocket

import (
	"sync"

	"github.com/iamasit07/4-in-a-row/backend/utils"
)

// TokenManager manages persistent user tokens
type TokenManager struct {
	tokenToUsername map[string]string // userToken -> username
	usernameToToken map[string]string // username -> userToken
	mu              sync.RWMutex
}

func NewTokenManager() *TokenManager {
	return &TokenManager{
		tokenToUsername: make(map[string]string),
		usernameToToken: make(map[string]string),
	}
}

// GenerateUserToken creates a new persistent user token
func (tm *TokenManager) GenerateUserToken() string {
	return "tkn_" + utils.GenerateToken()
}

// GetUsernameByToken returns the username associated with a token
func (tm *TokenManager) GetUsernameByToken(token string) (string, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	username, exists := tm.tokenToUsername[token]
	return username, exists
}

// GetTokenByUsername returns the token associated with a username
func (tm *TokenManager) GetTokenByUsername(username string) (string, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	token, exists := tm.usernameToToken[username]
	return token, exists
}

// SetTokenUsername creates or updates the mapping between token and username
func (tm *TokenManager) SetTokenUsername(token, username string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	// Remove old USERNAME mapping for this TOKEN (if exists and different)
	if oldUsername, exists := tm.tokenToUsername[token]; exists && oldUsername != username {
		delete(tm.usernameToToken, oldUsername)
	}
	
	// Remove old TOKEN mapping for this USERNAME (if exists and different)
	if oldToken, exists := tm.usernameToToken[username]; exists && oldToken != token {
		delete(tm.tokenToUsername, oldToken)
	}
	
	tm.tokenToUsername[token] = username
	tm.usernameToToken[username] = token
}

// RemoveToken removes a token and its associated username
func (tm *TokenManager) RemoveToken(token string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if username, exists := tm.tokenToUsername[token]; exists {
		delete(tm.usernameToToken, username)
	}
	delete(tm.tokenToUsername, token)
}
