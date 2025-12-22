package websocket

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/iamasit07/4-in-a-row/backend/models"
)

// ConnectionManager manages WebSocket connections by user ID
type ConnectionManager struct {
	connections map[int64]*websocket.Conn
	mu          sync.RWMutex
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[int64]*websocket.Conn),
	}
}

// AddConnection adds or updates a connection for a user
// If user already has a connection, it replaces it (for multi-device handling)
func (cm *ConnectionManager) AddConnection(userID int64, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Close old connection if exists
	if oldConn, exists := cm.connections[userID]; exists {
		log.Printf("[CONNECTION] Replacing existing connection for user %d", userID)
		oldConn.Close()
	}

	cm.connections[userID] = conn
	log.Printf("[CONNECTION] Added connection for user %d", userID)
}

// GetConnection retrieves a connection by user ID
func (cm *ConnectionManager) GetConnection(userID int64) (*websocket.Conn, bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, exists := cm.connections[userID]
	return conn, exists
}

// RemoveConnection removes a connection by user ID
func (cm *ConnectionManager) RemoveConnection(userID int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn, exists := cm.connections[userID]; exists {
		conn.Close()
		delete(cm.connections, userID)
		log.Printf("[CONNECTION] Removed connection for user %d", userID)
	}
}

// SendMessage sends a message to a user by user ID
func (cm *ConnectionManager) SendMessage(userID int64, message models.ServerMessage) error {
	cm.mu.RLock()
	conn, exists := cm.connections[userID]
	cm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no connection found for user %d", userID)
	}

	err := conn.WriteJSON(message)
	if err != nil {
		log.Printf("[CONNECTION] Error sending message to user %d: %v", userID, err)
		return err
	}

	return nil
}

// DisconnectUser forcefully disconnects a user and removes their connection
func (cm *ConnectionManager) DisconnectUser(userID int64, reason string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn, exists := cm.connections[userID]; exists {
		log.Printf("[CONNECTION] Disconnecting user %d, reason: %s", userID, reason)

		// Send disconnect message before closing
		conn.WriteJSON(models.ServerMessage{
			Type:    "force_disconnect",
			Message: reason,
		})

		conn.Close()
		delete(cm.connections, userID)
	}
}

// GetAllConnections returns all active user IDs with connections
func (cm *ConnectionManager) GetAllConnections() []int64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	userIDs := make([]int64, 0, len(cm.connections))
	for userID := range cm.connections {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}
