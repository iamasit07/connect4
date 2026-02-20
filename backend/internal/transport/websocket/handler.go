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
	"github.com/iamasit07/connect4/backend/internal/config"
	"github.com/iamasit07/connect4/backend/internal/domain"
	"github.com/iamasit07/connect4/backend/internal/repository/redis"
	"github.com/iamasit07/connect4/backend/internal/service/game"
	"github.com/iamasit07/connect4/backend/internal/service/matchmaking"
	"github.com/iamasit07/connect4/backend/internal/service/session"
)

const (
	maxMessageSize = 4096
	maxConnsPerIP  = 5
	writeTimeout   = 10 * time.Second
	pongWait       = 60 * time.Second
	pingInterval   = 20 * time.Second
)

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
	
	gameLoops   map[string]bool 
	gameLoopsMu sync.Mutex
}

// NewHandler creates a new WebSocket handler with dependencies
func NewHandler(cm *ConnectionManager, mq *matchmaking.MatchmakingQueue, sm *game.SessionManager, gs *game.Service, as *session.AuthService) *Handler {
	allowedOrigins := config.AppConfig.AllowedOrigins

	h := &Handler{
		ConnManager:    cm,
		Matchmaking:    mq,
		SessionManager: sm,
		GameService:    gs,
		AuthService:    as,
		ipTracker:      newIPConnTracker(),
		gameLoops:      make(map[string]bool),
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
	
	sm.SetSessionCreatedCallback(h.EnsureEventLoopRunning)
	
	return h
}

// EnsureEventLoopRunning checks if an event loop is running for the game session, and starts one if not
func (h *Handler) EnsureEventLoopRunning(gs *game.GameSession) {
	h.gameLoopsMu.Lock()
	defer h.gameLoopsMu.Unlock()
	
	if h.gameLoops[gs.GameID] {
		return
	}
	
	h.gameLoops[gs.GameID] = true
	go h.ConsumeGameEvents(gs)
}

// ConsumeGameEvents listens to the decoupled game event channel and broadcasts to users
func (h *Handler) ConsumeGameEvents(gs *game.GameSession) {
	defer func() {
		h.gameLoopsMu.Lock()
		delete(h.gameLoops, gs.GameID)
		h.gameLoopsMu.Unlock()
		log.Printf("[WS] Event loop stopped for game %s", gs.GameID)
	}()

	log.Printf("[WS] Event loop started for game %s", gs.GameID)

	for {
		select {
		case event, ok := <-gs.Events:
			if !ok {
				log.Printf("[WS] Events channel closed for game %s", gs.GameID)
				return
			}
			msg, ok := event.Payload.(domain.ServerMessage)
			if !ok {
				log.Printf("[WS] Unknown event payload type for game %s", gs.GameID)
				continue
			}

			for _, recipientID := range event.Recipients {
				go func(uid int64, m domain.ServerMessage) {
					h.ConnManager.SendMessage(uid, m)
				}(recipientID, msg)
			}
		case <-gs.Ctx.Done():
			log.Printf("[WS] Game context cancelled for game %s", gs.GameID)
			return
		}
	}
}

// extractIP returns the client IP address from the request.
func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
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

	conn.SetReadLimit(maxMessageSize)

	h.handleConnection(conn, clientIP)
}

// handleConnection manages the lifecycle of a single WebSocket connection
func (h *Handler) handleConnection(conn *websocket.Conn, clientIP string) {
	defer h.ipTracker.Decrement(clientIP)

	// Use a context to cleanly shut down the ping goroutine
	ctx, cancelPing := context.WithCancel(context.Background())
	defer cancelPing() // This kills the ping goroutine when handleConnection exits

	conn.SetReadDeadline(time.Now().Add(pongWait))

	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return // Clean exit when connection closes
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(writeTimeout))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
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
		sessionID = claims.SessionID
		h.ConnManager.AddConnection(userID, conn, username)

		if session, exists := h.SessionManager.GetSessionByUserID(userID); exists {
			// Ensure event loop is running for this session
			h.EnsureEventLoopRunning(session)
			
			if err := session.HandleReconnect(userID); err != nil {
				log.Printf("[WS] Reconnect failed for user %d: %v", userID, err)
			}
		}
	} else {
		log.Printf("[WS] Missing initialization or token")
		conn.Close()
		return
	}

	// 2. Cleanup on exit
	defer func() {
		h.Matchmaking.RemovePlayer(userID)

		if h.ConnManager.IsCurrentConnection(userID, conn) {
			gameSession, exists := h.SessionManager.GetSessionByUserID(userID)
			if exists {
				gameSession.HandleDisconnect(userID, h.SessionManager)
			}
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

		// --- STRICT PER-MESSAGE SESSION VALIDATION (OPTIMIZED) ---
		// Use Offline Validation (Signature + Blocklist) to avoid DB hits
		if msg.JWT != "" {
			claims, err := h.AuthService.ValidateTokenOffline(msg.JWT)
			if err != nil {
				log.Printf("[WS] Session revoked for user %d: %v", userID, err)
				h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Session invalidated/replaced"})
				return // Break loop and disconnect
			}
			// Sanity check
			if claims.UserID != userID || claims.SessionID != sessionID {
				log.Printf("[WS] Session mismatch or user spoofing attempt: %d vs %d", userID, claims.UserID)
				return
			}
		} else {
			
			if h.AuthService.IsSessionBlocked(sessionID) {
				log.Printf("[WS] Session blocked (Redis check): %s", sessionID)
				h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Session invalidated/replaced"})
				return
			}
		}

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

		// Clean up any existing game session before joining queue
		h.SessionManager.ForceCleanupForUser(userID)

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
		
		h.EnsureEventLoopRunning(gameSession)

		err := gameSession.HandleMove(userID, msg.Column)
		if err != nil {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: err.Error()})
		}

	case "request_rematch":
		gameSession, exists := h.SessionManager.GetSessionByUserID(userID)
		if !exists {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Game not found"})
			return
		}
		h.EnsureEventLoopRunning(gameSession)
		
		err := gameSession.HandleRematchRequest(userID, h.SessionManager)
		if err != nil {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: err.Error()})
		}

	case "rematch_response":
		gameSession, exists := h.SessionManager.GetSessionByUserID(userID)
		if !exists {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Game not found"})
			return
		}
		h.EnsureEventLoopRunning(gameSession)
		
		accept := msg.RematchResponse == "accept"
		err := gameSession.HandleRematchResponse(userID, accept, h.SessionManager)
		if err != nil {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: err.Error()})
		}

	case "abandon_game":
		gameSession, exists := h.SessionManager.GetSessionByUserID(userID)
		if !exists {
			return
		}
		h.EnsureEventLoopRunning(gameSession)
		
		gameSession.TerminateSessionWithReason(userID, "surrender")

	case "watch_game":
		gameSession, exists := h.SessionManager.GetSessionByGameID(msg.GameID)
		if !exists {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{Type: "error", Message: "Game not found or ended"})
			return
		}

		if gameSession.Player1ID == userID || (gameSession.Player2ID != nil && *gameSession.Player2ID == userID) {
			return
		}

		h.EnsureEventLoopRunning(gameSession)
		gameSession.AddSpectator(userID)

	case "leave_spectate":
		if gameSession, exists := h.SessionManager.GetSessionByGameID(msg.GameID); exists {
			gameSession.RemoveSpectator(userID)
		}

	case "get_game_state":
		if gameSession, exists := h.SessionManager.GetSessionByUserID(userID); exists {
			h.EnsureEventLoopRunning(gameSession)
			gameSession.HandleGetState(userID)
		} else {
			h.ConnManager.SendMessage(userID, domain.ServerMessage{
				Type:    "no_active_game",
				Message: "No active game session found",
			})
		}

	}
}

func (h *Handler) checkRateLimit(key string, window time.Duration) bool {
	if redis.RedisClient == nil || !redis.IsRedisEnabled() {
		return true
	}
	ctx := context.Background()
	ok, err := redis.RedisClient.SetNX(ctx, key, 1, window).Result()
	if err != nil {
		log.Printf("[WS] Rate limit check failed for key %s: %v", key, err)
		return true
	}
	return ok
}
