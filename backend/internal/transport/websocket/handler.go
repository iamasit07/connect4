package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/iamasit07/4-in-a-row/backend/internal/config"
	"github.com/iamasit07/4-in-a-row/backend/internal/domain"
	"github.com/iamasit07/4-in-a-row/backend/internal/repository/redis"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/game"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/matchmaking"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/session"
)

const (
	// maxMessageSize is the maximum allowed WebSocket message payload (4KB).
	// Prevents memory exhaustion from oversized messages.
	maxMessageSize = 4096

	// maxConnsPerIP limits concurrent WebSocket connections from a single IP.
	maxConnsPerIP = 5

	// writeTimeout is the deadline for writing a message to the peer.
	writeTimeout = 10 * time.Second
)

// ipConnTracker tracks the number of active WebSocket connections per IP address.
type ipConnTracker struct {
	mu    sync.Mutex
	conns map[string]int
}

func newIPConnTracker() *ipConnTracker {
	return &ipConnTracker{conns: make(map[string]int)}
}

func (t *ipConnTracker) Increment(ip string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conns[ip] >= maxConnsPerIP {
		return false
	}
	t.conns[ip]++
	return true
}

func (t *ipConnTracker) Decrement(ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.conns[ip] > 0 {
		t.conns[ip]--
	}
	if t.conns[ip] == 0 {
		delete(t.conns, ip)
	}
}

// Handler manages WebSocket dependencies
type Handler struct {
	ConnManager    *ConnectionManager
	Matchmaking    *matchmaking.MatchmakingQueue
	SessionManager *game.SessionManager
	GameService    *game.Service
	AuthService    *session.AuthService
	Upgrader       websocket.Upgrader
	ipTracker      *ipConnTracker
}

// NewHandler creates a new WebSocket handler with dependencies
func NewHandler(cm *ConnectionManager, mq *matchmaking.MatchmakingQueue, sm *game.SessionManager, gs *game.Service, as *session.AuthService) *Handler {
	allowedOrigins := config.AppConfig.AllowedOrigins

	return &Handler{
		ConnManager:    cm,
		Matchmaking:    mq,
		SessionManager: sm,
		GameService:    gs,
		AuthService:    as,
		ipTracker:      newIPConnTracker(),
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return false
				}
				for _, allowed := range allowedOrigins {
					if origin == allowed {
						return true
					}
				}
				log.Printf("[WS] Rejected WebSocket from disallowed origin: %s", origin)
				return false
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// extractIP returns the client IP address from the request.
func extractIP(r *http.Request) string {
	// Trust X-Forwarded-For if behind a reverse proxy
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first (client) IP
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// HandleWebSocket is the Gin handler that upgrades the connection
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Connection Exhaustion: enforce per-IP connection limit
	clientIP := extractIP(c.Request)
	if !h.ipTracker.Increment(clientIP) {
		log.Printf("[WS] Connection limit exceeded for IP: %s", clientIP)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many WebSocket connections"})
		return
	}

	conn, err := h.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.ipTracker.Decrement(clientIP)
		log.Printf("[WS] Upgrade error: %v", err)
		return
	}

	// Memory DoS: enforce maximum incoming message size (4KB)
	conn.SetReadLimit(maxMessageSize)

	h.handleConnection(conn, clientIP)
}

// handleConnection manages the lifecycle of a single WebSocket connection
func (h *Handler) handleConnection(conn *websocket.Conn, clientIP string) {
	// Decrement IP connection count when this connection closes
	defer h.ipTracker.Decrement(clientIP)

	// Set read deadline to detect stale connections
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Keep-alive pinger with write deadline
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			conn.SetWriteDeadline(time.Now().Add(writeTimeout))
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

		h.ConnManager.AddConnection(userID, conn, username)
	} else {
		log.Printf("[WS] Missing initialization or token")
		conn.Close()
		return
	}

	// 2. Cleanup on exit
	defer func() {
		h.Matchmaking.RemovePlayer(userID)

		// Notify game session if active
		gameSession, exists := h.SessionManager.GetSessionByUserID(userID)
		if exists {
			gameSession.HandleDisconnect(userID, h.ConnManager, h.SessionManager)
		}

		// Clean up spectator status across all sessions
		h.SessionManager.RemoveSpectatorFromAll(userID)

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
		// Security: Ensure only the currently active session can send messages.
		// If a new device logs in, old connections will be disconnected on next message.
		if msg.JWT != "" {
			claims, err := h.AuthService.ValidateToken(msg.JWT)
			if err != nil {
				log.Printf("[WS] Invalid token for user %d: %v", userID, err)
				h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Session invalidated"})
				return // Break loop and disconnect
			}
			// Update context if user somehow changed (unlikely but safe)
			if claims.UserID != userID {
				log.Printf("[WS] User mismatch: expected %d, got %d", userID, claims.UserID)
				h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "User mismatch"})
				return
			}
			sessionID = claims.SessionID
		} else {
			// Verify the session is still the active one in DB
			// This catches if user logged in from another device
			sess, err := h.AuthService.GetSession(sessionID)
			if err != nil {
				log.Printf("[WS] Session lookup error for sessionID=%s, userID=%d: %v", sessionID, userID, err)
				h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Session lookup failed"})
				return
			}
			if sess == nil {
				log.Printf("[WS] Session not found: sessionID=%s, userID=%d", sessionID, userID)
				h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Session not found"})
				return
			}
			if !sess.IsActive {
				log.Printf("[WS] Session inactive: sessionID=%s, userID=%d, IsActive=%v", sessionID, userID, sess.IsActive)
				h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Session expired or logged out"})
				return
			}
			// Additional check: Verify this session is still the user's active session
			activeSession, err := h.AuthService.GetActiveSession(userID)
			if err != nil {
				log.Printf("[WS] Active session lookup error for userID=%d: %v", userID, err)
			} else if activeSession == nil {
				log.Printf("[WS] No active session found for userID=%d", userID)
				h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "No active session"})
				return
			} else if activeSession.SessionID != sessionID {
				log.Printf("[WS] Session mismatch: current=%s, active=%s, userID=%d - User logged in from another device",
					sessionID, activeSession.SessionID, userID)
				h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Logged in from another device"})
				return
			}
		}
		// ---------------------------------------------

		h.processMessage(userID, msg)
	}
}

// processMessage routes specific actions
func (h *Handler) processMessage(userID int64, msg domain.ClientMessage) {
	// Security: Block spectators from sending game-affecting messages
	switch msg.Type {
	case "make_move", "request_rematch", "rematch_response", "abandon_game":
		if h.SessionManager.IsSpectatorAnywhere(userID) {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{
				Type:    "error",
				Message: "Spectators cannot perform game actions",
			})
			return
		}
	}
	switch msg.Type {
	case "find_match":
		difficulty := msg.Difficulty

		// Rate-limit find_match: max 1 request per 3 seconds per user
		rateLimitKey := fmt.Sprintf("ratelimit:find_match:%d", userID)
		if !h.checkRateLimit(rateLimitKey, 3*time.Second) {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Too many requests. Please wait."})
			return
		}

		// Enforce one active queue/game per user
		if _, inGame := h.SessionManager.GetSessionByUserID(userID); inGame {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "You are already in a game"})
			return
		}

		// Clean up any existing game session before joining queue
		// If game is active, it's abandoned. If finished (rematch window), it's cleaned up.
		h.SessionManager.ForceCleanupForUser(userID, h.ConnManager)

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

	case "watch_game":
		gameSession, exists := h.SessionManager.GetSessionByGameID(msg.GameID)
		if !exists {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Game not found or ended"})
			return
		}

		// Ensure the user isn't already playing in this game
		if gameSession.Player1ID == userID || (gameSession.Player2ID != nil && *gameSession.Player2ID == userID) {
			return
		}

		gameSession.AddSpectator(userID, h.ConnManager)

	case "leave_spectate":
		if gameSession, exists := h.SessionManager.GetSessionByGameID(msg.GameID); exists {
			gameSession.RemoveSpectator(userID)
		}
	}
}

// checkRateLimit uses Redis to enforce a per-key cooldown.
// Returns true if the action is allowed, false if rate-limited.
func (h *Handler) checkRateLimit(key string, window time.Duration) bool {
	if redis.RedisClient == nil || !redis.IsRedisEnabled() {
		return true // If Redis is unavailable, allow (fail open)
	}
	ctx := context.Background()
	ok, err := redis.RedisClient.SetNX(ctx, key, 1, window).Result()
	if err != nil {
		log.Printf("[WS] Rate limit check failed for key %s: %v", key, err)
		return true // Fail open on Redis error
	}
	return ok
}
