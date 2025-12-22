package websocket

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/iamasit07/4-in-a-row/backend/models"
)

type ConnectionManager struct {
	connections map[int64]*websocket.Conn
	mu          sync.RWMutex
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[int64]*websocket.Conn),
	}
}

// Replaces existing connection if user connects from another device
func (cm *ConnectionManager) AddConnection(userID int64, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if oldConn, exists := cm.connections[userID]; exists {
		log.Printf("[CONNECTION] Replacing existing connection for user %d", userID)
		oldConn.Close()
	}

	cm.connections[userID] = conn
	log.Printf("[CONNECTION] Added connection for user %d", userID)
}

func (cm *ConnectionManager) GetConnection(userID int64) (*websocket.Conn, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[userID]
	return conn, exists
}

func (cm *ConnectionManager) RemoveConnection(userID int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn, exists := cm.connections[userID]; exists {
		conn.Close()
		delete(cm.connections, userID)
		log.Printf("[CONNECTION] Removed connection for user %d", userID)
	}
}

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

func (cm *ConnectionManager) DisconnectUser(userID int64, reason string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if conn, exists := cm.connections[userID]; exists {
		log.Printf("[CONNECTION] Disconnecting user %d, reason: %s", userID, reason)

		conn.WriteJSON(models.ServerMessage{
			Type:    "force_disconnect",
			Message: reason,
		})

		conn.Close()
		delete(cm.connections, userID)
	}
}

func (cm *ConnectionManager) GetAllConnections() []int64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	userIDs := make([]int64, 0, len(cm.connections))
	for userID := range cm.connections {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}
