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
	connections  map[string]*websocket.Conn  // username -> websocket connection
	writeMutexes map[string]*sync.Mutex      // username -> write mutex for that connection
	mu           sync.RWMutex
}

func NewConnectionManager() *ConnectionManager {
	conn := &ConnectionManager{
		connections:  make(map[string]*websocket.Conn),
		writeMutexes: make(map[string]*sync.Mutex),
		mu:           sync.RWMutex{},
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
	cm.writeMutexes[username] = &sync.Mutex{} // Create write mutex for this connection
	return nil
}

func (cm *ConnectionManager) RemoveConnection(username string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.connections, username)
	delete(cm.writeMutexes, username)
}

func (cm *ConnectionManager) GetConnection(username string) (*websocket.Conn, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	conn, exists := cm.connections[username]
	return conn, exists
}

func (cm *ConnectionManager) SendMessage(username string, message models.ServerMessage) error {
	cm.mu.RLock()
	conn, exists := cm.connections[username]
	writeMutex, mutexExists := cm.writeMutexes[username]
	cm.mu.RUnlock()
	
	if !exists || !mutexExists {
		return fmt.Errorf("connection for username %s does not exist", username)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Use per-connection write mutex to prevent concurrent writes
	writeMutex.Lock()
	defer writeMutex.Unlock()
	
	return conn.WriteMessage(websocket.TextMessage, data)
}