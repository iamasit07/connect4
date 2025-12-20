package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/iamasit07/4-in-a-row/backend/config"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/server"
)

func HandleConnection(conn *websocket.Conn, connManager *ConnectionManager, 
	matchMakingQueue *models.MatchmakingQueue, sessionManager *server.SessionManager, 
	tokenManager *TokenManager) {
	
	defer conn.Close()

	var currentUsername string
	var currentToken string
	var isAuthenticated bool

	for {
		var message models.ClientMessage
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println("Error during message reading:", err)
			if isAuthenticated {
				HandleDisconnectCleanUp(currentToken, sessionManager, connManager)
			}
			break
		}

		HandleWebSocket(message, conn, connManager, matchMakingQueue, sessionManager, tokenManager, &currentUsername, &currentToken, &isAuthenticated)
	}
}

func HandleWebSocket(message models.ClientMessage, conn *websocket.Conn,
	connManager *ConnectionManager, matchMakingQueue *models.MatchmakingQueue,
	sessionManager *server.SessionManager, tokenManager *TokenManager,
	currentUsername *string, currentToken *string, isAuthenticated *bool) {

	switch message.Type {
	case "join_queue":
		HandleJoinQueue(message, conn, connManager, matchMakingQueue, sessionManager, tokenManager, currentUsername, currentToken, isAuthenticated)
	case "move":
		HandleMove(message, conn, sessionManager, connManager, *currentToken, *isAuthenticated)
	case "reconnect":
		HandleReconnect(message, conn, sessionManager, connManager, currentUsername, currentToken, isAuthenticated)
	default:
		SendErrorMessage(conn, "unknown_message_type", "Unknown message type")
	}
}

func HandleJoinQueue(message models.ClientMessage, conn *websocket.Conn,
	connManager *ConnectionManager, matchMakingQueue *models.MatchmakingQueue,
	sessionManager *server.SessionManager, tokenManager *TokenManager,
	currentUsername *string, currentToken *string, isAuthenticated *bool) {

	if message.Username == "" {
		SendErrorMessage(conn, "invalid_username", "Username cannot be empty")
		return
	}

	if message.Username == models.BotUsername {
		SendErrorMessage(conn, "invalid_username", "Username cannot be 'BOT'")
		return
	}

	// Token validation: prevent token theft by verifying ownership
	userToken := message.UserToken
	if userToken == "" {
		userToken = tokenManager.GenerateUserToken()
		log.Printf("[TOKEN] Generated new user token for %s: %s", message.Username, userToken)
	} else {
		existingUser, exists := tokenManager.GetUsernameByToken(userToken)
		if exists && existingUser != message.Username {
			log.Printf("[SECURITY] User %s attempted to use token owned by %s - forcing new token generation",
				message.Username, existingUser)
			userToken = tokenManager.GenerateUserToken()
		}
	}

	tokenManager.SetTokenUsername(userToken, message.Username)

	*currentUsername = message.Username
	*currentToken = userToken
	*isAuthenticated = true

	// Check for active session and terminate if abandoning
	session, exists := sessionManager.GetSessionByToken(userToken)
	if exists && !session.Game.IsFinished() {
		log.Printf("[ABANDON] Player %s is abandoning game %s to join new queue", *currentUsername, session.GameID)
		err := session.TerminateSession(userToken, "abandoned", connManager)
		if err != nil {
			log.Printf("Failed to terminate abandoned session: %v", err)
		}
		sessionManager.RemoveSession(
			session.GameID,
			session.Player1Token,
			session.Player2Token,
		)
	}

	err := connManager.AddConnection(userToken, *currentUsername, conn)
	if err != nil {
		SendErrorMessage(conn, "token_taken", "This token is already connected")
		*isAuthenticated = false
		return
	}

	err = matchMakingQueue.AddPlayerToQueue(userToken, *currentUsername)
	if err != nil {
		SendErrorMessage(conn, "queue_error", "Failed to join matchmaking queue")
		connManager.RemoveConnection(userToken)
		*isAuthenticated = false
		return
	}

	successMessage := models.ServerMessage{
		Type:      "queue_joined",
		Message:   "Successfully joined matchmaking queue",
		UserToken: userToken,
	}
	responseData, _ := json.Marshal(successMessage)
	conn.WriteMessage(websocket.TextMessage, responseData)
}

func HandleMove(message models.ClientMessage, conn *websocket.Conn,
	sessionManager *server.SessionManager, connManager *ConnectionManager,
	currentToken string, isAuthenticated bool) {

	if !isAuthenticated {
		SendErrorMessage(conn, "not_authenticated", "You must join the queue first")
		return
	}

	session, exists := sessionManager.GetSessionByToken(currentToken)
	if !exists {
		SendErrorMessage(conn, "no_active_game", "You are not in an active game")
		return
	}

	err := session.HandleMove(currentToken, message.Column, connManager)
	if err != nil {
		SendErrorMessage(conn, "invalid_move", err.Error())
		return
	}
}

func HandleReconnect(message models.ClientMessage, conn *websocket.Conn,
	sessionManager *server.SessionManager, connManager *ConnectionManager,
	currentUsername *string, currentToken *string, isAuthenticated *bool) {

	username := message.Username
	gameID := message.GameID
	userToken := message.UserToken

	// SECURITY CHECK 1: Verify gameID exists
	session, exists := sessionManager.GetSessionByGameID(gameID)
	if !exists {
		SendErrorMessage(conn, "game_not_found", "Game not found")
		return
	}

	// SECURITY CHECK 2: Verify game is not finished
	if session.Game.IsFinished() {
		SendErrorMessage(conn, "game_finished", "Cannot reconnect to a finished game")
		return
	}

	// SECURITY CHECK 3: Verify username matches a player in the game
	isPlayer1 := username == session.Player1Username
	isPlayer2 := username == session.Player2Username
	if !isPlayer1 && !isPlayer2 {
		SendErrorMessage(conn, "not_in_game", "You are not a player in this game")
		return
	}

	// SECURITY CHECK 4: Verify user token matches
	var expectedToken string
	if isPlayer1 {
		expectedToken = session.Player1Token
	} else {
		expectedToken = session.Player2Token
	}

	if expectedToken != userToken {
		log.Printf("[RECONNECT] Token mismatch for %s in game %s", username, gameID)
		SendErrorMessage(conn, "invalid_token", "Invalid user token")
		return
	}

	// SECURITY CHECK 5: Verify game doesn't already have 2 active connections
	activeConnections := 0
	if _, exists := connManager.GetConnection(session.Player1Token); exists {
		activeConnections++
	}
	if !session.IsBot() {
		if _, exists := connManager.GetConnection(session.Player2Token); exists {
			activeConnections++
		}
	}
	if activeConnections >= 2 {
		log.Printf("[RECONNECT] Game %s already has 2 active connections", gameID)
		SendErrorMessage(conn, "game_full", "Both players are already connected to this game")
		return
	}

	// SECURITY CHECK 6: Verify token is not currently connected
	_, isConnected := connManager.GetConnection(userToken)
	if isConnected {
		SendErrorMessage(conn, "already_connected", "This token is already connected")
		return
	}

	// SECURITY CHECK 7: Verify player is actually disconnected
	isDisconnected := false
	for _, disconnectedToken := range session.DisconnectedPlayers {
		if disconnectedToken == userToken {
			isDisconnected = true
			break
		}
	}

	if !isDisconnected {
		SendErrorMessage(conn, "not_disconnected", "You are not disconnected from this game")
		return
	}

	// All security checks passed - proceed with reconnection
	*currentUsername = username
	*currentToken = userToken
	*isAuthenticated = true
	connManager.AddConnection(userToken, username, conn)

	err := session.HandleReconnect(userToken, connManager)
	if err != nil {
		SendErrorMessage(conn, "reconnect_failed", err.Error())
		*isAuthenticated = false
		return
	}
}

func HandleDisconnectCleanUp(userToken string, sessionManager *server.SessionManager, connManager *ConnectionManager) {
	session, exists := sessionManager.GetSessionByToken(userToken)
	if !exists {
		return
	}

	if !session.Game.IsFinished() {
		session.HandleDisconnect(userToken, connManager, sessionManager)
	}
}

func SendErrorMessage(conn *websocket.Conn, errorType, message string) {
	errorMessage := models.ErrorMessage{
		Type:    errorType,
		Message: message,
	}
	responseData, _ := json.Marshal(errorMessage)
	conn.WriteMessage(websocket.TextMessage, responseData)
}

// CreateUpgrader creates a WebSocket upgrader with proper CORS settings
func CreateUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			for _, allowed := range config.AppConfig.AllowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
	}
}
