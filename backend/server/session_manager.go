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

