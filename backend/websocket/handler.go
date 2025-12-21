package websocket

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/iamasit07/4-in-a-row/backend/config"
	"github.com/iamasit07/4-in-a-row/backend/db"
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
		HandleMove(message, conn, sessionManager, connManager, *currentToken, *currentUsername, *isAuthenticated, tokenManager)
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

	// Only set token-username mapping if:
	// 1. Token is new (doesn't exist in manager), OR
	// 2. Token already belongs to this username (safe to overwrite)
	existingMappedUser, tokenMapped := tokenManager.GetUsernameByToken(userToken)
	if !tokenMapped || existingMappedUser == message.Username {
		tokenManager.SetTokenUsername(userToken, message.Username)
	}


	*currentUsername = message.Username
	*currentToken = userToken
	*isAuthenticated = true

	// Check for active session and terminate if abandoning
	session, exists := sessionManager.GetSessionByToken(userToken)
	if exists && !session.Game.IsFinished() {
		log.Printf("[ABANDON] Player %s is abandoning game %s to join new queue", *currentUsername, session.GameID)
		
		// CRITICAL: Remove the abandoning player's connection BEFORE terminating
		// This prevents game_over messages from bleeding into their new game
		opponentToken := session.GetOpponentToken(userToken)
		connManager.RemoveConnection(userToken)
		
		// Terminate the session (only opponent receives game_over now)
		err := session.TerminateSessionByAbandonment(userToken, opponentToken, connManager)
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
	currentToken, currentUsername string, isAuthenticated bool,
	tokenManager *TokenManager) {

	if !isAuthenticated {
		SendErrorMessage(conn, "not_authenticated", "You must join the queue first")
		return
	}

	tokenToValidate := message.UserToken
	if tokenToValidate == "" {
		tokenToValidate = currentToken
	}

	if tokenToValidate != currentToken {
		log.Printf("[CORRUPTION] Token mismatch: message token %s != connection token %s", 
			tokenToValidate, currentToken)
		HandleTokenCorruption(currentToken, currentUsername, sessionManager, connManager, "token_mismatch")
		return
	}

	valid, reason := ValidatePlayerToken(tokenManager, sessionManager, tokenToValidate, currentUsername)
	if !valid {
		HandleTokenCorruption(tokenToValidate, currentUsername, sessionManager, connManager, reason)
		return
	}

	session, exists := sessionManager.GetSessionByToken(tokenToValidate)
	if !exists {
		SendErrorMessage(conn, "no_active_game", "You are not in an active game")
		return
	}

	err := session.HandleMove(tokenToValidate, message.Column, connManager)
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

	// UserToken is always required (primary identifier)
	if userToken == "" {
		SendErrorMessage(conn, "invalid_reconnect", "UserToken is required for reconnection")
		return
	}

	// Must provide at least username OR gameID
	if username == "" && gameID == "" {
		SendErrorMessage(conn, "invalid_reconnect", "Either username or gameID is required")
		return
	}

	var session *server.GameSession
	var exists bool

	// CASE 1: GameID provided (with or without username)
	if gameID != "" {
		session, exists = sessionManager.GetSessionByGameID(gameID)
		if !exists {
			// Session not in memory - check database to see if game is finished
			log.Printf("[RECONNECT] Session %s not found in memory, checking database", gameID)
			gameResult, err := db.GetGameByID(gameID)
			if err != nil {
				log.Printf("[RECONNECT] Database error checking game %s: %v", gameID, err)
				SendErrorMessage(conn, "database_error", "Failed to check game status")
				return
			}
			
			if gameResult != nil {
				// Game exists in database - it's finished
				log.Printf("[RECONNECT] Game %s found in database - already finished", gameID)
				SendErrorMessage(conn, "game_finished", "This game has already ended")
				return
			}
			
			// Game doesn't exist anywhere
			SendErrorMessage(conn, "game_not_found", "Game not found")
			return
		}

		// Verify token belongs to this game
		if session.Player1Token != userToken && session.Player2Token != userToken {
			log.Printf("[RECONNECT] Token %s not found in game %s", userToken, gameID)
			SendErrorMessage(conn, "invalid_token", "Your token is not associated with this game")
			return
		}

		// Derive username from token if not provided
		if username == "" {
			username = session.GetUsernameByToken(userToken)
			log.Printf("[RECONNECT] Derived username %s from token for game %s", username, gameID)
		} else {
			// Username provided - validate it matches
			sessionUsername := session.GetUsernameByToken(userToken)
			if sessionUsername != username {
				log.Printf("[RECONNECT] Username mismatch: provided %s, expected %s", username, sessionUsername)
				SendErrorMessage(conn, "username_mismatch", "Username does not match your token")
				return
			}
		}

	} else if username != "" {
		// CASE 2: Username-only (no gameID)
		session, exists = sessionManager.GetSessionByToken(userToken)
		if !exists {
			log.Printf("[RECONNECT] No active game found for token %s", userToken)
			SendErrorMessage(conn, "game_not_found", "No active game found for your account")
			return
		}

		// Validate username matches token's game
		sessionUsername := session.GetUsernameByToken(userToken)
		if sessionUsername != username {
			log.Printf("[RECONNECT] Username mismatch for token %s: provided %s, expected %s", 
				userToken, username, sessionUsername)
			SendErrorMessage(conn, "username_mismatch", "Username does not match your active game")
			return
		}

		gameID = session.GameID
		log.Printf("[RECONNECT] Found game %s for username %s via token lookup", gameID, username)
	}

	// SECURITY CHECK: Verify game is not finished
	if session.Game.IsFinished() {
		SendErrorMessage(conn, "game_finished", "Cannot reconnect to a finished game")
		return
	}

	// SECURITY CHECK: Remove stale connection if it exists
	if _, exists := connManager.GetConnection(userToken); exists {
		log.Printf("[RECONNECT] Removing stale connection for %s", username)
		connManager.RemoveConnection(userToken)
	}

	// SECURITY CHECK: Verify game doesn't already have 2 active connections
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

	// SECURITY CHECK: Verify player is actually disconnected
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

func ValidatePlayerToken(
	tokenManager *TokenManager,
	sessionManager *server.SessionManager,
	userToken, expectedUsername string,
) (valid bool, reason string) {

	actualUsername, exists := tokenManager.GetUsernameByToken(userToken)
	if !exists {
		return false, "token_not_found"
	}

	if actualUsername != expectedUsername {
		return false, "username_mismatch"
	}

	session, exists := sessionManager.GetSessionByToken(userToken)
	if !exists {
		return false, "no_active_session"
	}

	if session.Player1Token != userToken && session.Player2Token != userToken {
		return false, "token_session_mismatch"
	}

	return true, ""
}

func HandleTokenCorruption(
	corruptedToken, currentUsername string,
	sessionManager *server.SessionManager,
	connManager *ConnectionManager,
	reason string,
) {
	log.Printf("[CORRUPTION] Token corruption detected for %s (token: %s), reason: %s",
		currentUsername, corruptedToken, reason)

	conn, exists := connManager.GetConnection(corruptedToken)
	if exists {
		SendErrorMessage(conn, "token_corrupted", "Your token is invalid. You have been disconnected. Please refresh and rejoin with correct credentials.")
	}

	session, exists := sessionManager.GetSessionByToken(corruptedToken)
	if exists {
		session.HandleDisconnect(corruptedToken, connManager, sessionManager)
	}

	connManager.RemoveConnection(corruptedToken)
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
