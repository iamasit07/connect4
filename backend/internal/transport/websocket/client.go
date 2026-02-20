package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iamasit07/connect4/backend/internal/domain"
)

// ConnectionManager handles active WebSocket connections thread-safely
type ConnectionManager struct {
	connections map[int64]*websocket.Conn
	usernames   map[int64]string
	
	// writeMu ensures only one goroutine writes to a specific socket at a time.
	// This is CRITICAL because conn.WriteJSON is not thread-safe.
	writeMu     map[int64]*sync.Mutex 
	
	mu          sync.RWMutex // Protects the maps themselves
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[int64]*websocket.Conn),
		usernames:   make(map[int64]string),
		writeMu:     make(map[int64]*sync.Mutex),
	}
}

// AddConnection registers a new connection and initializes its write lock
func (cm *ConnectionManager) AddConnection(userID int64, conn *websocket.Conn, username string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 1. Close old connection if it exists (Single Session Logic)
	if oldConn, exists := cm.connections[userID]; exists {
		oldConn.Close()
	}

	// 2. Register new connection
	cm.connections[userID] = conn
	cm.usernames[userID] = username
	
	// 3. Initialize the per-user write mutex (The Missing Logic)
	cm.writeMu[userID] = &sync.Mutex{}
}

// RemoveConnection removes a user's connection and cleans up locks
func (cm *ConnectionManager) RemoveConnection(userID int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn, exists := cm.connections[userID]; exists {
		conn.Close()
		delete(cm.connections, userID)
		delete(cm.usernames, userID)
		delete(cm.writeMu, userID)
	}
}

// RemoveConnectionIfMatching avoids race conditions where we might accidentally 
// close a NEW connection when trying to clean up an OLD one.
func (cm *ConnectionManager) RemoveConnectionIfMatching(userID int64, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if currentConn, exists := cm.connections[userID]; exists {
		if currentConn == conn {
			currentConn.Close()
			delete(cm.connections, userID)
			delete(cm.usernames, userID)
			delete(cm.writeMu, userID)
		}
	}
}

func (cm *ConnectionManager) IsCurrentConnection(userID int64, conn *websocket.Conn) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	currentConn, exists := cm.connections[userID]
	return exists && currentConn == conn
}

// SendMessage sends a JSON message to a specific user securely
func (cm *ConnectionManager) SendMessage(userID int64, message domain.ServerMessage) error {
	// 1. Acquire global read lock to find the user's socket & mutex
	cm.mu.RLock()
	conn, exists := cm.connections[userID]
	mu, muExists := cm.writeMu[userID]
	cm.mu.RUnlock()

	if !exists || !muExists {
		return nil // User disconnected, ignore
	}

	// 2. Acquire the PER-USER write lock (Critical Thread-Safety Fix)
	mu.Lock()
	defer mu.Unlock()

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.WriteJSON(message)
}

// BroadcastMessage sends a message to all connected users
func (cm *ConnectionManager) BroadcastMessage(message domain.ServerMessage) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for userID := range cm.connections {
		// We launch goroutines so one slow user doesn't block the broadcast
		go func(uid int64) {
			cm.SendMessage(uid, message)
		}(userID)
	}
}

// DisconnectUser sends a generic disconnect message and closes the socket.
// This satisfies the Disconnector interface used in AuthHandler.
func (cm *ConnectionManager) DisconnectUser(userID int64, reason string) {
	msg := domain.ServerMessage{
		Type:    "force_disconnect",
		Message: reason,
	}
	// Try to send the message (best effort)
	_ = cm.SendMessage(userID, msg)
	
	// Then force close
	cm.RemoveConnection(userID)
}

// GetUsername returns the username for a connected user
func (cm *ConnectionManager) GetUsername(userID int64) (string, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	name, exists := cm.usernames[userID]
	return name, exists
}
