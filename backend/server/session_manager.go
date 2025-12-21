package server

import (
	"fmt"
	"sync"
)

type SessionManager struct {
	Session    map[string]*GameSession  // gameID → GameSession
	TokenToGame map[string]string        // userToken → gameID
	Mux        *sync.Mutex
}

func (sm *SessionManager) CreateSession(player1Token, player1Username, player2Token, player2Username string, conn ConnectionManagerInterface) *GameSession {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	// Create new game session with tokens
	session := NewGameSession(player1Token, player1Username, player2Token, player2Username, conn)
	gameID := session.GameID
	sm.Session[gameID] = session
	sm.TokenToGame[player1Token] = gameID
	if player2Username != "BOT" {
		sm.TokenToGame[player2Token] = gameID
	}

	fmt.Printf("[SESSION] Created session %s: %s (%s) vs %s (%s)\n",
		gameID, player1Username, player1Token, player2Username, player2Token)
	return session
}

func (sm *SessionManager) GetSessionByToken(userToken string) (*GameSession, bool) {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	gameID, exists := sm.TokenToGame[userToken]
	if !exists {
		return nil, false
	}

	session, exists := sm.Session[gameID]
	return session, exists
}

func (sm *SessionManager) GetSessionByGameID(gameID string) (*GameSession, bool) {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	session, exists := sm.Session[gameID]
	return session, exists
}

func (sm *SessionManager) NewSessionManager() *SessionManager {
	session := &SessionManager{
		Session:    make(map[string]*GameSession),
		TokenToGame: make(map[string]string),
		Mux:        &sync.Mutex{},
	}
	return session
}

func (sm *SessionManager) RemoveSession(gameID string, player1Token, player2Token string) error {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	fmt.Printf("[SESSION] Removing session %s\n", gameID)
	delete(sm.Session, gameID)
	delete(sm.TokenToGame, player1Token)
	if player2Token != "" {
		delete(sm.TokenToGame, player2Token)
	}

	return nil
}

// GetSessionByGameIDAndUsername finds a session where the username matches a player
// Used for token corruption recovery
func (sm *SessionManager) GetSessionByGameIDAndUsername(gameID, username string) (*GameSession, string, bool) {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	session, exists := sm.Session[gameID]
	if !exists {
		return nil, "", false
	}

	// Check if username matches player1 or player2
	if session.Player1Username == username {
		return session, session.Player1Token, true
	}
	if session.Player2Username == username {
		return session, session.Player2Token, true
	}

	return nil, "", false
}

// UpdatePlayerToken updates a player's token in the session and TokenToGame mapping
// Used for token corruption recovery
func (sm *SessionManager) UpdatePlayerToken(session *GameSession, oldToken, newToken, username string) error {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	session.mu.Lock()
	defer session.mu.Unlock()

	delete(sm.TokenToGame, oldToken)
	sm.TokenToGame[newToken] = session.GameID

	var targetToken *string
	if session.Player1Username == username {
		targetToken = &session.Player1Token
	} else if session.Player2Username == username {
		targetToken = &session.Player2Token
	} else {
		return fmt.Errorf("username not found in session")
	}

	playerID := session.PlayerMapping[oldToken]
	delete(session.PlayerMapping, oldToken)
	session.PlayerMapping[newToken] = playerID
	*targetToken = newToken

	fmt.Printf("[SESSION] Updated token for %s in session %s\n", username, session.GameID)
	return nil
}


