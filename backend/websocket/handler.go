package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/iamasit07/4-in-a-row/backend/config"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/server"
)

func (cm *ConnectionManager) HandleWebSocket(w http.ResponseWriter, r *http.Request, sessionManager *server.SessionManager, matchMakingQueue *models.MatchmakingQueue) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins configured in config (production URL + localhost)
			origin := r.Header.Get("Origin")
			for _, allowedOrigin := range config.AppConfig.AllowedOrigins {
				if allowedOrigin == origin {
					return true
				}
			}
			// Reject if origin not in allowed list
			return false
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		// Don't call http.Error here - upgrade already wrote headers
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
			HandleJoinQueue(clientMessage, conn, cm, matchMakingQueue, sessionManager, &currentUsername, &isAuthenticated)
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
	connManager *ConnectionManager, matchMakingQueue *models.MatchmakingQueue, sessionManager *server.SessionManager, currentUsername *string, isAuthenticated *bool) {
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

	// Check for existing active session
	session, exists := sessionManager.GetSessionByPlayer(*currentUsername)
	if exists && !session.Game.IsFinished() {
		// Player is abandoning an active game - terminate it
		err := session.TerminateSession(*currentUsername, "abandoned", connManager)
		if err != nil {
			log.Printf("Failed to terminate abandoned session: %v", err)
		}
		// Remove from session manager
		sessionManager.RemoveSession(
			session.GameID,
			session.Player1Username,
			session.Player2Username,
		)
	}

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
	token := message.Token

	// SECURITY CHECK 1: Require gameID, username, AND token
	if gameID == "" || username == "" || token == "" {
		SendErrorMessage(conn, "invalid_input", "Game ID, username, and session token are required for reconnection")
		return
	}

	// Find session by gameID
	session, exists := sessionManager.GetSessionByGameID(gameID)
	if !exists {
		SendErrorMessage(conn, "no_active_game", "No active game found with this ID")
		return
	}

	// SECURITY CHECK 2: Verify game is not finished
	if session.Game.IsFinished() {
		SendErrorMessage(conn, "game_finished", "Cannot reconnect to a finished game")
		return
	}

	// SECURITY CHECK 3: Verify username is actually a player in this game
	_, isPlayer := session.GetPlayerID(username)
	if !isPlayer {
		SendErrorMessage(conn, "not_in_game", "You are not a player in this game")
		return
	}

	// SECURITY CHECK 4: Verify username is not currently connected
	_, isConnected := connManager.GetConnection(username)
	if isConnected {
		SendErrorMessage(conn, "already_connected", "This username is already connected")
		return
	}

	// SECURITY CHECK 5: Verify player is actually disconnected
	isDisconnected := false
	for _, disconnectedUser := range session.DisconnectedPlayers {
		if disconnectedUser == username {
			isDisconnected = true
			break
		}
	}

	if !isDisconnected {
		SendErrorMessage(conn, "not_disconnected", "You are not disconnected from this game")
		return
	}

	// SECURITY CHECK 6: Verify session token matches
	expectedToken, hasToken := session.SessionTokens[username]
	if !hasToken || expectedToken != token {
		SendErrorMessage(conn, "invalid_token", "Invalid session token")
		return
	}

	// All security checks passed - proceed with reconnection
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
