package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/server"
)

func (cm *ConnectionManager) HandleWebSocket(w http.ResponseWriter, r *http.Request,
	sessionManager *server.SessionManager, matchMakingQueue *models.MatchmakingQueue) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}

	defer conn.Close()

	var currentUsername string
	var isAuthenticated bool = false

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				fmt.Println("WebSocket closed normally")
			} else {
				fmt.Println("Error reading message:", err)
			}

			if isAuthenticated {
				HandleDisconnectCleanUp(currentUsername, sessionManager, cm)
			}
			break
		}

		var clientMessage models.ClientMessage
		if err := json.Unmarshal(message, &clientMessage); err != nil {
			SendErrorMessage(conn, "invalid_message", "Failed to parse message")
			continue
		}

		switch clientMessage.Type {
		case "join_queue":
			HandleJoinQueue(clientMessage, conn, cm, matchMakingQueue, &currentUsername, &isAuthenticated)
		case "make_move":
			HandleMakeMove(clientMessage, currentUsername, isAuthenticated, conn, sessionManager, cm)
		case "reconnect":
			HandleReconnect(clientMessage, conn, sessionManager, cm, &currentUsername, &isAuthenticated)
		default:
			SendErrorMessage(conn, "unknown_message_type", "Unknown message type")
		}
	}
}

func HandleJoinQueue(message models.ClientMessage, conn *websocket.Conn,
	connManager *ConnectionManager, matchMakingQueue *models.MatchmakingQueue, currentUsername *string, isAuthenticated *bool) {
	if message.Username == "" {
		SendErrorMessage(conn, "invalid_username", "Username cannot be empty")
		return
	}

	if message.Username == models.BotUsername {
		SendErrorMessage(conn, "invalid_username", "Username cannot be 'BOT'")
		return
	}

	*currentUsername = message.Username
	*isAuthenticated = true

	err := connManager.AddConnection(*currentUsername, conn)
	if err != nil {
		SendErrorMessage(conn, "username_taken", "Username is already taken")
		*isAuthenticated = false
		return
	}

	err = matchMakingQueue.AddPlayerToQueue(*currentUsername)
	if err != nil {
		SendErrorMessage(conn, "queue_error", "Failed to join matchmaking queue")
		connManager.RemoveConnection(*currentUsername)
		*isAuthenticated = false
		return
	}

	response := models.ServerMessage{
		Type:    "queue_joined",
		Message: "Successfully joined the matchmaking queue... (Bot will join automatically if no opponent found in 10 seconds)",
	}

	responseData, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, responseData)
}

func HandleMakeMove(message models.ClientMessage, currentUsername string, isAuthenticated bool,
	conn *websocket.Conn, sessionManager *server.SessionManager, connManager *ConnectionManager) {
	if !isAuthenticated {
		SendErrorMessage(conn, "unauthenticated", "You must join the queue before making a move")
		return
	}

	session, exist := sessionManager.GetSessionByPlayer(currentUsername)
	if !exist {
		SendErrorMessage(conn, "no_active_game", "You are not in an active game")
		return
	}

	if message.Column < 0 || message.Column >= models.Columns {
		SendErrorMessage(conn, "invalid_move", "Column out of bounds")
		return
	}

	err := session.HandleMove(currentUsername, message.Column, connManager)
	if err != nil {
		SendErrorMessage(conn, "invalid_move", err.Error())
		return
	}
}

func HandleReconnect(message models.ClientMessage, conn *websocket.Conn, sessionManager *server.SessionManager,
	connManager *ConnectionManager, currentUsername *string, isAuthenticated *bool) {
	username := message.Username
	if username == "" {
		SendErrorMessage(conn, "invalid_username", "Username cannot be empty")
		return
	}

	session, exists := sessionManager.GetSessionByPlayer(username)
	if !exists {
		SendErrorMessage(conn, "no_active_game", "No active game found for reconnection")
		return
	}

	*currentUsername = username
	*isAuthenticated = true
	connManager.AddConnection(username, conn)

	err := session.HandleReconnect(username, connManager)
	if err != nil {
		SendErrorMessage(conn, "reconnect_failed", err.Error())
		*isAuthenticated = false
		return
	}
}

func HandleDisconnectCleanUp(username string, sessionManager *server.SessionManager, connManager *ConnectionManager) {
	session, exists := sessionManager.GetSessionByPlayer(username)
	if exists {
		session.HandleDisconnect(username, connManager)
	}
	connManager.RemoveConnection(username)
}

func SendErrorMessage(conn *websocket.Conn, errorType, errorMsg string) {
	errMsg := models.ErrorMessage{
		Type:    errorType,
		Message: errorMsg,
	}
	data, _ := json.Marshal(errMsg)
	conn.WriteMessage(websocket.TextMessage, data)
}
