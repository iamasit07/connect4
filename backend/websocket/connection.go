package websocket

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/iamasit07/4-in-a-row/backend/models"
)

// important part to manage websocket connections & operations
type ConnectionManager struct {
	connections map[string]*websocket.Conn  // username -> websocket connection
	mu          sync.RWMutex
}

func NewConnectionManager() *ConnectionManager {
	conn := &ConnectionManager{
		connections: make(map[string]*websocket.Conn),
		mu:          sync.RWMutex{},
	}
	return conn
}

func (cm *ConnectionManager) AddConnection(username string, conn *websocket.Conn) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if _, exists := cm.connections[username]; exists {
		return fmt.Errorf("username already connected")
	}
	
	cm.connections[username] = conn
	return nil
}

func (cm *ConnectionManager) RemoveConnection(username string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.connections, username)
}

func (cm *ConnectionManager) GetConnection(username string) (*websocket.Conn, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	conn, exists := cm.connections[username]
	return conn, exists
}

func (cm *ConnectionManager) SendMessage(username string, message models.ServerMessage) error {
	conn, exists := cm.GetConnection(username)
	if !exists {
		return fmt.Errorf("connection for username %s does not exist", username)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}