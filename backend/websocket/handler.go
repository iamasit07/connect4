package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/server"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

func HandleConnection(conn *websocket.Conn, connManager *ConnectionManager, matchMakingQueue *models.MatchmakingQueue, sessionManager *server.SessionManager) {
	defer conn.Close()

	var currentUserID int64
	var currentUsername string
	isAuthenticated := false

	for {
		var message models.ClientMessage
		err := conn.ReadJSON(&message)
		if err != nil {
			if isAuthenticated {
				log.Printf("[WS] Connection closed for user %d (%s): %v", currentUserID, currentUsername, err)
				HandleDisconnect(currentUserID, connManager, sessionManager)
			} else {
				log.Printf("[WS] Unauthenticated connection closed: %v", err)
			}
			break
		}

		// Validate JWT from every message
		if message.JWT == "" {
			SendErrorMessage(conn, "not_authenticated", "JWT token required")
			continue
		}

		claims, err := utils.ValidateJWT(message.JWT)
		if err != nil {
			SendErrorMessage(conn, "invalid_token", "Invalid or expired JWT token")
			log.Printf("[WS] JWT validation failed: %v", err)
			continue
		}

		if !isAuthenticated {
			currentUserID = claims.UserID
			currentUsername = claims.Username
			isAuthenticated = true

			if _, exists := connManager.GetConnection(currentUserID); exists {
				log.Printf("[WS] User %d (%s) connecting from new device, disconnecting old session", currentUserID, currentUsername)
				connManager.DisconnectUser(currentUserID, "Logged in from another device")
			}

			connManager.AddConnection(currentUserID, conn)
			log.Printf("[WS] User %d (%s) authenticated and connected", currentUserID, currentUsername)
		}

		if claims.UserID != currentUserID {
			SendErrorMessage(conn, "token_mismatch", "JWT token does not match current user")
			continue
		}

		HandleWebSocket(message, conn, connManager, matchMakingQueue, sessionManager, currentUserID, currentUsername)
	}
}

func HandleWebSocket(message models.ClientMessage, conn *websocket.Conn, connManager *ConnectionManager, matchMakingQueue *models.MatchmakingQueue, sessionManager *server.SessionManager, userID int64, username string) {
	switch message.Type {
	case "join_queue":
		HandleJoinQueue(userID, username, connManager, matchMakingQueue, sessionManager)
	case "move":
		HandleMove(message, userID, sessionManager, connManager)
	case "reconnect":
		HandleReconnect(message, userID, sessionManager, connManager)
	default:
		SendErrorMessage(conn, "unknown_message_type", "Unknown message type")
	}
}

func HandleJoinQueue(userID int64, username string, connManager *ConnectionManager, matchMakingQueue *models.MatchmakingQueue, sessionManager *server.SessionManager) {
	log.Printf("[QUEUE] User %d (%s) attempting to join queue", userID, username)

	if sessionManager.HasActiveGame(userID) {
		session, _ := sessionManager.GetSessionByUserID(userID)
		if session != nil {
			log.Printf("[QUEUE] User %d (%s) has active game %s - terminating it", userID, username, session.GameID)

			err := session.TerminateSessionByAbandonment(userID, connManager)
			if err != nil {
				log.Printf("[QUEUE] Failed to terminate user's session: %v", err)
			}

			sessionManager.RemoveSession(session.GameID)
		}
	}

	err := matchMakingQueue.AddPlayerToQueue(userID, username)
	if err != nil {
		log.Printf("[QUEUE] Error adding user to queue: %v", err)
		connManager.SendMessage(userID, models.ServerMessage{
			Type:    "queue_error",
			Message: "Failed to join matchmaking queue",
		})
		return
	}

	connManager.SendMessage(userID, models.ServerMessage{
		Type:    "queue_joined",
		Message: "Joined matchmaking queue",
	})

	log.Printf("[QUEUE] User %d (%s) successfully joined queue", userID, username)
}

func HandleMove(message models.ClientMessage, userID int64, sessionManager *server.SessionManager, connManager *ConnectionManager) {
	session, exists := sessionManager.GetSessionByUserID(userID)
	if !exists {
		connManager.SendMessage(userID, models.ServerMessage{
			Type:    "no_active_game",
			Message: "No active game found",
		})
		return
	}

	err := session.HandleMove(userID, message.Column, connManager)
	if err != nil {
		log.Printf("[MOVE] Error handling move for user %d: %v", userID, err)
		connManager.SendMessage(userID, models.ServerMessage{
			Type:    "invalid_move",
			Message: err.Error(),
		})
		return
	}
}

func HandleReconnect(message models.ClientMessage, userID int64, sessionManager *server.SessionManager, connManager *ConnectionManager) {
	log.Printf("[RECONNECT] User %d attempting to reconnect", userID)

	var gameID string
	var session *server.GameSession
	var isPlayer bool

	if message.GameID == "" {
		log.Printf("[RECONNECT] No gameID provided, looking up user %d's active game", userID)
		session, isPlayer = sessionManager.GetSessionByUserID(userID)
		if !isPlayer {
			connManager.SendMessage(userID, models.ServerMessage{
				Type:    "no_active_game",
				Message: "No active game found. Please start a new game.",
			})
			return
		}
		gameID = session.GameID
		log.Printf("[RECONNECT] Found active game %s for user %d", gameID, userID)
	} else {
		gameID = message.GameID
		session, isPlayer = sessionManager.GetSessionByGameIDAndUserID(gameID, userID)
		if !isPlayer {
			sessionByID, exists := sessionManager.GetSessionByGameID(gameID)
			if !exists {
				connManager.SendMessage(userID, models.ServerMessage{
					Type:    "game_finished",
					Message: "This game has already ended",
				})
				return
			}
			if sessionByID.Game.IsFinished() {
				connManager.SendMessage(userID, models.ServerMessage{
					Type:    "game_finished",
					Message: "This game has already ended",
				})
			} else {
				connManager.SendMessage(userID, models.ServerMessage{
					Type:    "not_in_game",
					Message: "You are not a player in this game",
				})
			}
			return
		}
	}

	if session.Game.IsFinished() {
		connManager.SendMessage(userID, models.ServerMessage{
			Type:    "game_finished",
			Message: "This game has already finished",
		})
		return
	}

	err := session.HandleReconnect(userID, connManager)
	if err != nil {
		log.Printf("[RECONNECT] Error reconnecting user %d: %v", userID, err)
		connManager.SendMessage(userID, models.ServerMessage{
			Type:    "reconnect_failed",
			Message: err.Error(),
		})
		return
	}

	log.Printf("[RECONNECT] User %d successfully reconnected to game %s", userID, gameID)
}

func HandleDisconnect(userID int64, connManager *ConnectionManager, sessionManager *server.SessionManager) {
	session, exists := sessionManager.GetSessionByUserID(userID)
	if !exists {
		log.Printf("[DISCONNECT] User %d disconnected (no active game)", userID)
		return
	}

	err := session.HandleDisconnect(userID, connManager, sessionManager)
	if err != nil {
		log.Printf("[DISCONNECT] Error handling disconnect for user %d: %v", userID, err)
	}
}

func SendErrorMessage(conn *websocket.Conn, errorType, message string) {
	conn.WriteJSON(models.ServerMessage{
		Type:    errorType,
		Message: message,
	})
}

func CreateUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
}
