package server

import (
	"fmt"
	"log"
	"sync"
)

// SessionManager manages all active game sessions
type SessionManager struct {
	Session    map[string]*GameSession // gameID → GameSession
	UserToGame map[int64]string        // userID → gameID (for quick lookup)
	Mux        *sync.Mutex
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		Session:    make(map[string]*GameSession),
		UserToGame: make(map[int64]string),
		Mux:        &sync.Mutex{},
	}
}

// CreateSession creates a new game session
func (sm *SessionManager) CreateSession(player1ID int64, player1Username string, player2ID *int64, player2Username string, conn ConnectionManagerInterface) *GameSession {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	// Create new game session with user IDs
	session := NewGameSession(player1ID, player1Username, player2ID, player2Username, conn)
	gameID := session.GameID
	sm.Session[gameID] = session
	sm.UserToGame[player1ID] = gameID

	if player2Username != "BOT" && player2ID != nil {
		sm.UserToGame[*player2ID] = gameID
	}

	log.Printf("[SESSION] Created session %s: %s (ID: %d) vs %s (ID: %v)\\n",
		gameID, player1Username, player1ID, player2Username, player2ID)
	return session
}

// GetSessionByUserID retrieves a session by user ID
func (sm *SessionManager) GetSessionByUserID(userID int64) (*GameSession, bool) {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	gameID, exists := sm.UserToGame[userID]
	if !exists {
		return nil, false
	}

	session, exists := sm.Session[gameID]
	return session, exists
}

// GetSessionByGameID retrieves a session by game ID
func (sm *SessionManager) GetSessionByGameID(gameID string) (*GameSession, bool) {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	session, exists := sm.Session[gameID]
	return session, exists
}

// GetSessionByGameIDAndUserID retrieves a session and verifies the user is a player
func (sm *SessionManager) GetSessionByGameIDAndUserID(gameID string, userID int64) (*GameSession, bool) {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	session, exists := sm.Session[gameID]
	if !exists {
		return nil, false
	}

	// Verify user is actually in this game
	if session.Player1ID == userID {
		return session, true
	}
	if session.Player2ID != nil && *session.Player2ID == userID {
		return session, true
	}

	return nil, false
}

// RemoveSession removes a session by game ID
func (sm *SessionManager) RemoveSession(gameID string) error {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	session, exists := sm.Session[gameID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	log.Printf("[SESSION] Removing session %s", gameID)

	// Remove from UserToGame mapping
	delete(sm.UserToGame, session.Player1ID)
	if session.Player2ID != nil {
		delete(sm.UserToGame, *session.Player2ID)
	}

	// Remove session
	delete(sm.Session, gameID)

	return nil
}

// HasActiveGame checks if a user has an active game
func (sm *SessionManager) HasActiveGame(userID int64) bool {
	sm.Mux.Lock()
	defer sm.Mux.Unlock()

	gameID, exists := sm.UserToGame[userID]
	if !exists {
		return false
	}

	session, exists := sm.Session[gameID]
	if !exists {
		// Clean up orphaned mapping
		delete(sm.UserToGame, userID)
		return false
	}

	return !session.Game.IsFinished()
}
