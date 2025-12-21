package websocket

import (
	"sync"

	"github.com/iamasit07/4-in-a-row/backend/utils"
)

type TokenManager struct {
	tokenToUsername map[string]string
	usernameToToken map[string]string
	mu              sync.RWMutex
}

func NewTokenManager() *TokenManager {
	return &TokenManager{
		tokenToUsername: make(map[string]string),
		usernameToToken: make(map[string]string),
	}
}

func (tm *TokenManager) GenerateUserToken() string {
	return "tkn_" + utils.GenerateToken()
}

func (tm *TokenManager) GetUsernameByToken(token string) (string, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	username, exists := tm.tokenToUsername[token]
	return username, exists
}

func (tm *TokenManager) GetTokenByUsername(username string) (string, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	token, exists := tm.usernameToToken[username]
	return token, exists
}

// SetTokenUsername maintains bidirectional token-username mapping with cleanup of old associations
func (tm *TokenManager) SetTokenUsername(token, username string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if oldUsername, exists := tm.tokenToUsername[token]; exists && oldUsername != username {
		delete(tm.usernameToToken, oldUsername)
	}
	
	if oldToken, exists := tm.usernameToToken[username]; exists && oldToken != token {
		delete(tm.tokenToUsername, oldToken)
	}
	
	tm.tokenToUsername[token] = username
	tm.usernameToToken[username] = token
}

func (tm *TokenManager) RemoveToken(token string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if username, exists := tm.tokenToUsername[token]; exists {
		delete(tm.usernameToToken, username)
	}
	delete(tm.tokenToUsername, token)
}
