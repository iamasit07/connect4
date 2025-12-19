package server

import (
	"sync"
)

type SessionManager struct {
	Session    map[string]*GameSession  // gameID -> GameSession
	UserToGame map[string]string        // username -> gameID
	Mux        *sync.Mutex
}

func (sm *SessionManager) CreateSession(player1Username, player2Username string, conn ConnectionManagerInterface) *GameSession {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	session := NewGameSession(player1Username, player2Username, conn)
	gameID := session.GameID
	sm.Session[gameID] = session
	sm.UserToGame[player1Username] = gameID
	if player2Username != "BOT" {
		sm.UserToGame[player2Username] = gameID
	}

	return session
}

func (sm *SessionManager) GetSessionByPlayer(username string) (*GameSession, bool) {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	gameID, exists := sm.UserToGame[username]
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
		UserToGame: make(map[string]string),
		Mux:        &sync.Mutex{},
	}
	return session
}

