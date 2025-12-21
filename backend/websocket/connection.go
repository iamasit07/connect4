package websocket

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/iamasit07/4-in-a-row/backend/models"
)

// Connection holds a WebSocket connection with its username
type Connection struct {
	Username   string
	Conn       *websocket.Conn
	WriteMutex *sync.Mutex
}

// ConnectionManager manages websocket connections by userToken
type ConnectionManager struct {
	connections map[string]*Connection  // userToken → Connection
	mu          sync.RWMutex
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*Connection),
		mu:          sync.RWMutex{},
	}
}

func (cm *ConnectionManager) AddConnection(userToken, username string, conn *websocket.Conn) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if _, exists := cm.connections[userToken]; exists {
		return fmt.Errorf("token already connected")
	}
	
	cm.connections[userToken] = &Connection{
		Username:   username,
		Conn:       conn,
		WriteMutex: &sync.Mutex{},
	}
	return nil
}

func (cm *ConnectionManager) RemoveConnection(userToken string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.connections, userToken)
}

func (cm *ConnectionManager) GetConnection(userToken string) (*websocket.Conn, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	connection, exists := cm.connections[userToken]
	if !exists {
		return nil, false
	}
	return connection.Conn, true
}

func (cm *ConnectionManager) SendMessage(userToken string, message models.ServerMessage) error {
	cm.mu.RLock()
	connection, exists := cm.connections[userToken]
	cm.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("connection for token %s does not exist", userToken)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Use per-connection write mutex to prevent concurrent writes
	connection.WriteMutex.Lock()
	defer connection.WriteMutex.Unlock()
	
	return connection.Conn.WriteMessage(websocket.TextMessage, data)
}