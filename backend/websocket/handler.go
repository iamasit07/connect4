package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/server"
	"github.com/joho/godotenv"
)

func (cm *ConnectionManager) HandleWebSocket(w http.ResponseWriter, r *http.Request, sessionManager *server.SessionManager, matchMakingQueue *models.MatchmakingQueue) {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origins := r.Header.Get("Origin")
			return origins == frontendURL || origins == "http://localhost:3000"
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
			HandleMakeMove(clientMessage, currentUsername, &isAuthenticated, conn, sessionManager, cm)
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

func HandleMakeMove(message models.ClientMessage, currentUsername string, isAuthenticated *bool,
	conn *websocket.Conn, sessionManager *server.SessionManager, connManager *ConnectionManager) {
	if !*isAuthenticated {
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
	gameID := message.GameID
	username := message.Username

	// User must provide at least one: username or gameID
	if gameID == "" && username == "" {
		SendErrorMessage(conn, "invalid_input", "Please provide either username or game ID")
		return
	}

	var session *server.GameSession
	var exists bool
	var playerUsername string

	// Try to find session by gameID first, then by username
	if gameID != "" {
		session, exists = sessionManager.GetSessionByGameID(gameID)
		if exists && username != "" {
			// Verify username is in this game
			_, isPlayer := session.GetPlayerID(username)
			if !isPlayer {
				SendErrorMessage(conn, "not_in_game", "Username does not match this game")
				return
			}
			playerUsername = username
		} else if exists {
			// GameID provided but no username, need to determine which player
			if session.Player1Username != models.BotUsername {
				playerUsername = session.Player1Username
			} else {
				playerUsername = session.Player2Username
			}
		}
	} else {
		// Only username provided
		session, exists = sessionManager.GetSessionByPlayer(username)
		playerUsername = username
	}

	if !exists {
		SendErrorMessage(conn, "no_active_game", "No active game found for reconnection")
		return
	}

	*currentUsername = playerUsername
	*isAuthenticated = true
	connManager.AddConnection(playerUsername, conn)

	err := session.HandleReconnect(playerUsername, connManager)
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
