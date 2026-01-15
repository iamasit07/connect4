package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iamasit07/4-in-a-row/backend/internal/domain"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/game"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/matchmaking"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/session"
)

// Handler manages WebSocket dependencies
type Handler struct {
	ConnManager    *ConnectionManager
	Matchmaking    *matchmaking.MatchmakingQueue
	SessionManager *game.SessionManager
	GameService    *game.Service
	AuthService    *session.AuthService
	Upgrader       websocket.Upgrader
}

// NewHandler creates a new WebSocket handler with dependencies
func NewHandler(cm *ConnectionManager, mq *matchmaking.MatchmakingQueue, sm *game.SessionManager, gs *game.Service, as *session.AuthService) *Handler {
	return &Handler{
		ConnManager:    cm,
		Matchmaking:    mq,
		SessionManager: sm,
		GameService:    gs,
		AuthService:    as,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// HandleWebSocket is the HTTP handler that upgrades the connection
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade error: %v", err)
		return
	}

	h.handleConnection(conn)
}

// handleConnection manages the lifecycle of a single WebSocket connection
func (h *Handler) handleConnection(conn *websocket.Conn) {
	// Set read deadline to detect stale connections
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	
	// Keep-alive pinger
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}()

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	var userID int64
	var username string
	var sessionID string

	// 1. Wait for Initialization (Auth)
	_, data, err := conn.ReadMessage()
	if err != nil {
		log.Printf("[WS] Read error during init: %v", err)
		conn.Close()
		return
	}

	var message domain.ClientMessage
	if err := json.Unmarshal(data, &message); err != nil {
		log.Printf("[WS] Invalid JSON during init: %v", err)
		conn.Close()
		return
	}

	if message.Type == "init" && message.JWT != "" {
		// Validate JWT using AuthService (Stateful DB Check)
		claims, err := h.AuthService.ValidateToken(message.JWT)
		if err != nil {
			log.Printf("[WS] Invalid token during init: %v", err)
			conn.WriteJSON(domain.ErrorMessage{Type: "error", Message: "Invalid token or session expired"})
			conn.Close()
			return
		}
		userID = claims.UserID
		username = claims.Username
		sessionID = claims.SessionID // Store session ID for subsequent checks

		log.Printf("[WS] Connection initialized for user: %s (ID: %d)", username, userID)
		
		h.ConnManager.AddConnection(userID, conn, username)
		
		// Check for existing active game to reconnect
		activeGame, exists := h.SessionManager.GetSessionByUserID(userID)
		if exists {
			log.Printf("[WS] Found active game for user %s: %s", username, activeGame.GameID)
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "reconnect_found", GameID: activeGame.GameID})
		}
	} else {
		log.Printf("[WS] Missing initialization or token")
		conn.Close()
		return
	}

	// 2. Cleanup on exit
	defer func() {
		log.Printf("[WS] Connection closed for user %s", username)
		h.Matchmaking.RemovePlayer(userID)
		
		// Notify game session if active
		gameSession, exists := h.SessionManager.GetSessionByUserID(userID)
		if exists {
			gameSession.HandleDisconnect(userID, h.ConnManager, h.SessionManager)
		}
		
		h.ConnManager.RemoveConnectionIfMatching(userID, conn)
	}()

	// 3. Main Message Loop
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] User disconnected unexpectedly: %v", err)
			}
			break
		}

		var msg domain.ClientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("[WS] Invalid message format: %v", err)
			continue
		}

		// --- STRICT PER-MESSAGE SESSION VALIDATION ---
		// If the client sends a new JWT (e.g. refresh), validate it fully.
		// If not, validate the existing Session ID against the DB to ensure it hasn't been revoked.
		if msg.JWT != "" {
			claims, err := h.AuthService.ValidateToken(msg.JWT)
			if err != nil {
				h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Session invalidated"})
				return // Break loop and disconnect
			}
			// Update context if user somehow changed (unlikely but safe)
			if claims.UserID != userID {
				h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "User mismatch"})
				return
			}
		} else {
			// Fast Check: Verify the session is still active in DB
			// This catches "Force Logout" or "Ban" events immediately
			sess, err := h.AuthService.GetSession(sessionID)
			if err != nil || sess == nil || !sess.IsActive {
				log.Printf("[WS] Session %s invalidated during connection for user %d", sessionID, userID)
				h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Session expired or logged out"})
				return // Break loop and disconnect
			}
		}
		// ---------------------------------------------

		h.processMessage(userID, msg)
	}
}

// processMessage routes specific actions
func (h *Handler) processMessage(userID int64, msg domain.ClientMessage) {
	switch msg.Type {
	case "find_match":
		difficulty := msg.Difficulty
		// Default to medium if not specified
		if difficulty == "" {
			difficulty = "medium" 
		}

		if h.SessionManager.HasActiveGame(userID) {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "You are already in a game"})
			return
		}

		username, _ := h.ConnManager.GetUsername(userID)
		err := h.Matchmaking.AddPlayerToQueue(userID, username, difficulty)
		if err != nil {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Failed to join queue"})
		} else {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "queue_joined"})
		}

	case "cancel_search":
		h.Matchmaking.RemovePlayer(userID)
		h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "queue_left"})

	case "make_move":
		gameSession, exists := h.SessionManager.GetSessionByUserID(userID)
		if !exists {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Game not found"})
			return
		}
		
		err := gameSession.HandleMove(userID, msg.Column, h.ConnManager)
		if err != nil {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: err.Error()})
		}

	case "reconnect":
		gameSession, exists := h.SessionManager.GetSessionByUserID(userID)
		if !exists {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "No active game found"})
			return
		}
		gameSession.HandleReconnect(userID, h.ConnManager)

	case "request_rematch":
		gameSession, exists := h.SessionManager.GetSessionByUserID(userID)
		if !exists {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Game not found"})
			return
		}
		// Pass SessionManager to handle rematch logic (creating new game)
		err := gameSession.HandleRematchRequest(userID, h.ConnManager, h.SessionManager)
		if err != nil {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: err.Error()})
		}

	case "rematch_response":
		gameSession, exists := h.SessionManager.GetSessionByUserID(userID)
		if !exists {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Game not found"})
			return
		}
		err := gameSession.HandleRematchResponse(userID, msg.RematchResponse, h.ConnManager, h.SessionManager)
		if err != nil {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: err.Error()})
		}

	case "abandon_game":
		gameSession, exists := h.SessionManager.GetSessionByUserID(userID)
		if !exists {
			return
		}
		gameSession.TerminateSessionByAbandonment(userID, h.ConnManager)
		h.SessionManager.RemoveSession(gameSession.GameID)
	}
}
