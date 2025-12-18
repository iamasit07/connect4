package websocket

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/websocket"
)

type ConnectionManager struct {
	connections map[string]*websocket.Conn
	mu          sync.RWMutex
}

func NewConnectionManager() *ConnectionManager {
	conn := &ConnectionManager{
		connections: make(map[string]*websocket.Conn),
	}
	return conn
}

func AddConnection(username string, conn *websocket.Conn, manager *ConnectionManager) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.connections[username] = conn
}

func RemoveConnection(username string, manager *ConnectionManager) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	delete(manager.connections, username)
}

func GetConnection(username string, manager *ConnectionManager) (*websocket.Conn, bool) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	conn, exists := manager.connections[username]
	return conn, exists
}

func SendMessage(username string, message models.ServerMessage, manager *ConnectionManager) error {
	conn, exists := GetConnection(username, manager)
	if !exists {
		return fmt.Errorf("connection for username %s does not exist", username)
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}

func BroadCastToGame(player1, player2 string, message models.ServerMessage, manager *ConnectionManager) error {
	err1 := SendMessage(player1, message, manager)
	err2 := SendMessage(player2, message, manager)

	if player1 == "BOT" {
		err1 = nil
	} else if player2 == "BOT" {
		err2 = nil
	}

	if err1 != nil && err2 != nil {
		return fmt.Errorf("failed to send message to both players: %v, %v", err1, err2)
	} else if err1 != nil {
		return fmt.Errorf("failed to send message to player1 (%s): %v", player1, err1)
	} else if err2 != nil {
		return fmt.Errorf("failed to send message to player2 (%s): %v", player2, err2)
	} else {
		return nil
	}
}