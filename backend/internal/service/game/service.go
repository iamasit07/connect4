package game

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/iamasit07/connect4/backend/internal/domain"
	"github.com/iamasit07/connect4/backend/internal/service/bot"
	"github.com/iamasit07/connect4/backend/pkg/uid"
)

type GameSession struct {
	GameID              string
	Player1ID           int64
	Player1Username     string
	Player2ID           *int64
	Player2Username     string
	Game                *domain.Game
	PlayerMapping       map[int64]domain.PlayerID
	Spectators          map[int64]bool
	Reason              string
	CreatedAt           time.Time
	FinishedAt          time.Time
	BotDifficulty       string      // "easy", "medium", "hard"
	PostGameTimer       *time.Timer // 30-second window for rematch after game ends
	RematchRequester    *int64      // userID of player who requested rematch
	RematchRequestTimer *time.Timer // 10-second window to accept rematch request
	TurnTimer           *time.Timer // 30-minute turn timer
	DisconnectTimer     *time.Timer      // Shared grace period timer
	DisconnectTime      time.Time        // When the disconnect timer started
	DisconnectedPlayers map[int64]bool   // Set of currently disconnected player IDs
	GracePeriodTimer    *time.Timer      // Short timer (3s) to debounce disconnect events

	mu             sync.Mutex
	repo           GameRepository
	sessionManager *SessionManager

	// Lifecycle management
	Ctx    context.Context
	cancel context.CancelFunc

	// Events channel for decoupled communication
	Events chan domain.GameEvent
}

type GameRepository interface {
	SaveGame(gameID string, player1ID int64, player1Username string, player2ID *int64, player2Username string, winnerID *int64, winnerUsername string, reason string, totalMoves, durationSeconds int, createdAt, finishedAt time.Time, boardState [][]int) error
}

// SessionManager manages active game sessions
type SessionManager struct {
	Session          map[string]*GameSession // gameID → GameSession
	UserToGame       map[int64]string        // userID → gameID (for quick lookup)
	mu               sync.RWMutex
	repo             GameRepository
	onSessionCreated func(*GameSession)
}

func NewSessionManager(repo GameRepository) *SessionManager {
	return &SessionManager{
		Session:    make(map[string]*GameSession),
		UserToGame: make(map[int64]string),
		repo:       repo,
	}
}

func (sm *SessionManager) SetSessionCreatedCallback(cb func(*GameSession)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onSessionCreated = cb
}

func (sm *SessionManager) CreateSession(player1ID int64, player1Username string, player2ID *int64, player2Username string, botDifficulty string) *GameSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := NewGameSession(player1ID, player1Username, player2ID, player2Username, botDifficulty, sm.repo, sm)
	gameID := session.GameID
	sm.Session[gameID] = session
	sm.UserToGame[player1ID] = gameID

	if player2ID != nil {
		sm.UserToGame[*player2ID] = gameID
	}

	if sm.onSessionCreated != nil {
		// Invoke callback (e.g., to start event loop)
		sm.onSessionCreated(session)
	}

	return session
}

func (sm *SessionManager) GetSessionByUserID(userID int64) (*GameSession, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	gameID, exists := sm.UserToGame[userID]
	if !exists {
		return nil, false
	}

	session, exists := sm.Session[gameID]
	return session, exists
}

func (sm *SessionManager) GetSessionByGameID(gameID string) (*GameSession, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.Session[gameID]
	return session, exists
}

func (sm *SessionManager) GetSessionByGameIDAndUserID(gameID string, userID int64) (*GameSession, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.Session[gameID]
	if !exists {
		return nil, false
	}

	if session.Player1ID == userID {
		return session, true
	}
	if session.Player2ID != nil && *session.Player2ID == userID {
		return session, true
	}

	return nil, false
}

func (sm *SessionManager) RemoveSession(gameID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return sm.removeSessionLocked(gameID)
}

// removeSessionLocked removes session from maps without acquiring lock (caller must hold it)
func (sm *SessionManager) removeSessionLocked(gameID string) error {
	session, exists := sm.Session[gameID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	delete(sm.UserToGame, session.Player1ID)
	if session.Player2ID != nil {
		delete(sm.UserToGame, *session.Player2ID)
	}

	delete(sm.Session, gameID)

	session.cancel()

	return nil
}

// RemoveSpectatorFromAll removes a user from the spectators list of all active sessions
func (sm *SessionManager) RemoveSpectatorFromAll(userID int64) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, session := range sm.Session {
		session.RemoveSpectator(userID)
	}
}

// ActiveGameInfo represents a live game for the /watch endpoint
type ActiveGameInfo struct {
	GameID         string `json:"gameId"`
	Player1        string `json:"player1Username"`
	Player2        string `json:"player2Username"`
	MoveCount      int    `json:"moveCount"`
	SpectatorCount int    `json:"spectatorCount"`
	StartedAt      string `json:"startedAt"`
}

// GetActiveGames returns a list of all active (non-finished) PvP game sessions
func (sm *SessionManager) GetActiveGames() []ActiveGameInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	games := make([]ActiveGameInfo, 0)
	for _, session := range sm.Session {
		// Only show active PvP games (not bot games, not finished)
		if session.IsBot() || session.Game.IsFinished() {
			continue
		}
		games = append(games, ActiveGameInfo{
			GameID:         session.GameID,
			Player1:        session.Player1Username,
			Player2:        session.Player2Username,
			MoveCount:      session.Game.MoveCount,
			SpectatorCount: len(session.Spectators),
			StartedAt:      session.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return games
}

// IsSpectator checks whether a user is a spectator in a specific game session
func (sm *SessionManager) IsSpectator(userID int64, gameID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.Session[gameID]
	if !exists {
		return false
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	return session.Spectators[userID]
}

// IsSpectatorAnywhere checks whether a user is spectating any game
func (sm *SessionManager) IsSpectatorAnywhere(userID int64) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, session := range sm.Session {
		session.mu.Lock()
		isSpec := session.Spectators[userID]
		session.mu.Unlock()
		if isSpec {
			return true
		}
	}
	return false
}

func (sm *SessionManager) ForceCleanupForUser(userID int64) {
	session, exists := sm.GetSessionByUserID(userID)
	if !exists {
		return
	}

	gameID := session.GameID

	session.mu.Lock()
	gameFinished := session.Game.IsFinished()

	if gameFinished {
		// Game is finished but session lingers (rematch window) — clean up timers
		if session.PostGameTimer != nil {
			session.PostGameTimer.Stop()
			session.PostGameTimer = nil
		}
		if session.RematchRequestTimer != nil {
			session.RematchRequestTimer.Stop()
			session.RematchRequestTimer = nil
		}
		session.RematchRequester = nil
		session.mu.Unlock()
	} else {
		// Game is still active — abandon it (saves to DB, notifies opponent)
		session.mu.Unlock()
		session.TerminateSessionWithReason(userID, "abandonment")
	}

	sm.RemoveSession(gameID)
}

func NewGameSession(player1ID int64, player1Username string, player2ID *int64, player2Username string, botDifficulty string, repo GameRepository, sm *SessionManager) *GameSession {
	gameID := uid.GenerateGameID()
	newGame := (&domain.Game{}).NewGame()

	mapping := make(map[int64]domain.PlayerID)
	mapping[player1ID] = domain.Player1
	if player2ID != nil {
		mapping[*player2ID] = domain.Player2
	}

	ctx, cancel := context.WithCancel(context.Background())



	gs := &GameSession{
		GameID:          gameID,
		Player1ID:       player1ID,
		Player1Username: player1Username,
		Player2ID:       player2ID,
		Player2Username: player2Username,
		Game:            newGame,
		PlayerMapping:   mapping,
		Spectators:      make(map[int64]bool),
		BotDifficulty:   botDifficulty,
		CreatedAt:       time.Now(),
		mu:              sync.Mutex{},
		repo:            repo,
		sessionManager:  sm,
		DisconnectedPlayers: make(map[int64]bool),
		Events: make(chan domain.GameEvent, 100),
		Ctx:    ctx,
		cancel: cancel,

	}

	gs.startTurnTimer()

	// Notify players directly via event
	recipients := []int64{player1ID}
	if player2ID != nil {
		recipients = append(recipients, *player2ID)
	}

	gs.broadcastEvent(domain.GameEvent{
		Type:       domain.EventGameStart,
		Recipients: recipients,
		Payload: domain.ServerMessage{
			Type:        "game_start",
			GameID:      gs.GameID,
			Opponent:    player2Username,
			YourPlayer:  int(domain.Player1),
			CurrentTurn: int(gs.Game.CurrentPlayer),
			Board:       gs.Game.Board,
		},
	})
	
	// Direct send to Player 1 - OVERRIDE for custom field
	gs.sendEvent(player1ID, domain.ServerMessage{
		Type:        "game_start", 
		GameID:      gs.GameID,
		Opponent:    player2Username,
		YourPlayer:  int(domain.Player1),
		CurrentTurn: int(gs.Game.CurrentPlayer),
		Board:       gs.Game.Board,
	})

	if player2ID != nil {
		gs.sendEvent(*player2ID, domain.ServerMessage{
			Type:        "game_start", 
			GameID:      gs.GameID,
			Opponent:    player1Username,
			YourPlayer:  int(domain.Player2),
			CurrentTurn: int(gs.Game.CurrentPlayer),
			Board:       gs.Game.Board,
		})
	}

	return gs
}

func (sm *SessionManager) CleanupOldSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	count := 0
	now := time.Now()

	for gameID, session := range sm.Session {
		if session.Game.IsFinished() {
			if now.Sub(session.FinishedAt) > 1*time.Hour {
				delete(sm.Session, gameID)
				delete(sm.UserToGame, session.Player1ID)
				if session.Player2ID != nil {
					delete(sm.UserToGame, *session.Player2ID)
				}
				close(session.Events) 
				count++
			}
		} else {
			if now.Sub(session.CreatedAt) > 24*time.Hour {
				delete(sm.Session, gameID)
				delete(sm.UserToGame, session.Player1ID)
				if session.Player2ID != nil {
					delete(sm.UserToGame, *session.Player2ID)
				}
				close(session.Events)
				count++
			}
		}
	}

	if count > 0 {
		log.Printf("[SESSION] Memory cleanup: Removed %d stale game sessions", count)
	}
}

// Helper to send an event to a single user
func (gs *GameSession) sendEvent(userID int64, payload interface{}) {
	select {
	case gs.Events <- domain.GameEvent{
		Type:       domain.EventInfo,
		Recipients: []int64{userID},
		Payload:    payload,
	}:
	case <-gs.Ctx.Done():
		// Session cancelled, stop sending
	}
}

// Helper to broadcast and event to specific users
func (gs *GameSession) broadcastEvent(event domain.GameEvent) {
	select {
	case gs.Events <- event:
	case <-gs.Ctx.Done():
		// Session cancelled, stop sending
	}
}

// Helper to get all participants (players + spectators)
func (gs *GameSession) getAllParticipants() []int64 {
	participants := []int64{gs.Player1ID}
	if gs.Player2ID != nil {
		participants = append(participants, *gs.Player2ID)
	}
	for spectatorID := range gs.Spectators {
		participants = append(participants, spectatorID)
	}
	return participants
}

func (gs *GameSession) HandleMove(userID int64, column int) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	playerID, exists := gs.GetPlayerID(userID)
	if !exists {
		return fmt.Errorf("player not found in game")
	}

	if gs.Game.CurrentPlayer != playerID {
		return fmt.Errorf("not your turn")
	}

	row, err := gs.Game.MakeMove(playerID, column)
	if err != nil {
		return err
	}

	recipients := gs.getAllParticipants()

	if gs.Game.Status == domain.StatusWon {
		if gs.TurnTimer != nil {
			gs.TurnTimer.Stop()
		}

		// 1. Move Made Message
		gs.broadcastEvent(domain.GameEvent{
			Type:       domain.EventMoveMade,
			Recipients: recipients,
			Payload: domain.ServerMessage{
				Type:     "move_made",
				Column:   column,
				Row:      row,
				Player:   int(playerID),
				Board:    gs.Game.Board,
				NextTurn: int(gs.Game.CurrentPlayer),
			},
		})

		gs.FinishedAt = time.Now()
		winnerUsername := gs.GetUsername(gs.Game.Winner)
		winnerID := userID
		gs.Reason = "connect_four"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())
		allowRematch := true

		// 2. Game Over Message
		gs.broadcastEvent(domain.GameEvent{
			Type:       domain.EventGameOver,
			Recipients: recipients,
			Payload: domain.ServerMessage{
				Type:         "game_over",
				Winner:       winnerUsername,
				Reason:       gs.Reason,
				Board:        gs.Game.Board,
				AllowRematch: &allowRematch,
			},
		})

		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			gs.Player2ID, gs.Player2Username, &winnerID, winnerUsername,
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))

		gs.StartPostGameTimer()
		return nil
	}

	if gs.Game.Status == domain.StatusDraw {
		if gs.TurnTimer != nil {
			gs.TurnTimer.Stop()
		}

		// 1. Move Made Message
		gs.broadcastEvent(domain.GameEvent{
			Type:       domain.EventMoveMade,
			Recipients: recipients,
			Payload: domain.ServerMessage{
				Type:     "move_made",
				Column:   column,
				Row:      row,
				Player:   int(playerID),
				Board:    gs.Game.Board,
				NextTurn: int(gs.Game.CurrentPlayer),
			},
		})

		gs.FinishedAt = time.Now()
		gs.Reason = "draw"
		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())
		allowRematch := true

		// 2. Game Over Message
		gs.broadcastEvent(domain.GameEvent{
			Type:       domain.EventGameOver,
			Recipients: recipients,
			Payload: domain.ServerMessage{
				Type:         "game_over",
				Winner:       "draw",
				Reason:       "draw",
				Board:        gs.Game.Board,
				AllowRematch: &allowRematch,
			},
		})

		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			gs.Player2ID, gs.Player2Username, nil, "draw",
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))

		gs.StartPostGameTimer()
		return nil
	}

	// Normal Move
	gs.broadcastEvent(domain.GameEvent{
		Type:       domain.EventMoveMade,
		Recipients: recipients,
		Payload: domain.ServerMessage{
			Type:     "move_made",
			Column:   column,
			Row:      row,
			Player:   int(playerID),
			Board:    gs.Game.Board,
			NextTurn: int(gs.Game.CurrentPlayer),
		},
	})

	gs.startTurnTimer()

	// TRIGGER BOT MOVE if applicable
	if gs.IsBot() && gs.Game.CurrentPlayer == domain.Player2 {
		go func() {
			select {
			case <-time.After(500 * time.Millisecond):
				if err := gs.HandleBotMove(); err != nil {
					log.Printf("[BOT] Error handling bot move: %v", err)
				}
			case <-gs.Ctx.Done():
				// Game cancelled during sleep
				return
			}
		}()
	}

	return nil
}

func (gs *GameSession) HandleBotMove() error {
	// Acquire lock since this is entry point from goroutine
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Check if session is still valid
	if gs.Ctx.Err() != nil {
		return nil
	}

	// Verify it's actually bot's turn (race condition check)
	if gs.Game.CurrentPlayer != domain.Player2 || !gs.IsBot() || gs.Game.IsFinished() {
		return nil
	}

	difficulty := gs.BotDifficulty
	if difficulty == "" {
		difficulty = "medium"
	}

	botColumn := bot.CalculateBestMove(gs.Game.Board, domain.Player2, difficulty)
	botRow, err := gs.Game.MakeMove(domain.Player2, botColumn)
	if err != nil {
		return err
	}

	recipients := gs.getAllParticipants()

	if gs.Game.Status == domain.StatusWon {
		if gs.TurnTimer != nil {
			gs.TurnTimer.Stop()
		}

		gs.broadcastEvent(domain.GameEvent{
			Type:       domain.EventMoveMade,
			Recipients: recipients,
			Payload: domain.ServerMessage{
				Type:     "move_made",
				Column:   botColumn,
				Row:      botRow,
				Player:   int(domain.Player2),
				Board:    gs.Game.Board,
				NextTurn: int(gs.Game.CurrentPlayer),
			},
		})

		gs.FinishedAt = time.Now()
		gs.Reason = "connect_four"
		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())
		allowRematch := true

		gs.broadcastEvent(domain.GameEvent{
			Type:       domain.EventGameOver,
			Recipients: recipients,
			Payload: domain.ServerMessage{
				Type:         "game_over",
				Winner:       gs.Player2Username,
				Reason:       gs.Reason,
				Board:        gs.Game.Board,
				AllowRematch: &allowRematch,
			},
		})

		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			nil, gs.Player2Username, nil, gs.Player2Username,
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))

		gs.StartPostGameTimer()
		return nil
	}

	if gs.Game.Status == domain.StatusDraw {
		if gs.TurnTimer != nil {
			gs.TurnTimer.Stop()
		}

		gs.broadcastEvent(domain.GameEvent{
			Type:       domain.EventMoveMade,
			Recipients: recipients,
			Payload: domain.ServerMessage{
				Type:     "move_made",
				Column:   botColumn,
				Row:      botRow,
				Player:   int(domain.Player2),
				Board:    gs.Game.Board,
				NextTurn: int(gs.Game.CurrentPlayer),
			},
		})

		gs.FinishedAt = time.Now()
		gs.Reason = "draw"
		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())
		allowRematch := true

		gs.broadcastEvent(domain.GameEvent{
			Type:       domain.EventGameOver,
			Recipients: recipients,
			Payload: domain.ServerMessage{
				Type:         "game_over",
				Winner:       "draw",
				Reason:       "draw",
				Board:        gs.Game.Board,
				AllowRematch: &allowRematch,
			},
		})

		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			nil, gs.Player2Username, nil, "draw",
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))

		gs.StartPostGameTimer()
		return nil
	}

	gs.broadcastEvent(domain.GameEvent{
		Type:       domain.EventMoveMade,
		Recipients: recipients,
		Payload: domain.ServerMessage{
			Type:     "move_made",
			Column:   botColumn,
			Row:      botRow,
			Player:   int(domain.Player2),
			Board:    gs.Game.Board,
			NextTurn: int(gs.Game.CurrentPlayer),
		},
	})
	
	gs.startTurnTimer()
	return nil
}

func (gs *GameSession) HandleDisconnect(userID int64, sessionManager *SessionManager) error {
	gs.mu.Lock()

	if gs.Game.IsFinished() {
		if gs.DisconnectTimer != nil {
			gs.DisconnectTimer.Stop()
			gs.DisconnectTimer = nil
		}
		if gs.RematchRequester != nil {
			gs.CancelRematchRequest()
		}
		
		gameID := gs.GameID
		gs.mu.Unlock() // Unlock before calling SessionManager
		sessionManager.RemoveSession(gameID)
		return nil
	}

	username := gs.GetUsernameByUserID(userID)
	gs.DisconnectedPlayers[userID] = true
	log.Printf("[DISCONNECT] Player %s (ID: %d) disconnected from game %s", username, userID, gs.GameID)

	if gs.DisconnectTimer != nil {
		gs.mu.Unlock()
		return nil
	}

	gs.DisconnectTime = time.Now()
	
	opponentID := gs.GetOpponentID(userID)
	if opponentID != nil {
		gs.sendEvent(*opponentID, domain.ServerMessage{
			Type:              "opponent_disconnected",
			Message:           "Opponent disconnected, waiting for reconnect...",
			DisconnectTimeout: 60,
		})
	}

	// Start 60-second grace period timer
	gs.DisconnectTimer = time.AfterFunc(60*time.Second, func() {
		gs.mu.Lock()
		defer gs.mu.Unlock()

		// Verify still disconnected
		if !gs.DisconnectedPlayers[userID] {
			return
		}

		gs.cleanupTimers()

		winnerID := gs.GetOpponentID(userID)
		winnerName := gs.GetOpponentUsername(userID)
		
		gs.Game.Status = domain.StatusWon
		gs.Reason = "abandonment"
		gs.FinishedAt = time.Now()
		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())
		
		recipients := gs.getAllParticipants()
		
		allowRematch := false
		gs.broadcastEvent(domain.GameEvent{
			Type: domain.EventGameOver,
			Recipients: recipients,
			Payload: domain.ServerMessage{
				Type: "game_over",
				Winner: winnerName,
				Reason: "abandonment",
				Board: gs.Game.Board,
				AllowRematch: &allowRematch,
			},
		})

		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			gs.Player2ID, gs.Player2Username, winnerID, winnerName,
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))
	})

	gs.mu.Unlock()
	return nil
}

func (gs *GameSession) HandleReconnect(userID int64) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	wasDisconnected := gs.DisconnectedPlayers[userID]
	if wasDisconnected {
		delete(gs.DisconnectedPlayers, userID)
		allConnected := true
		for _, disconnected := range gs.DisconnectedPlayers {
			if disconnected {
				allConnected = false
				break
			}
		}

		if allConnected && gs.DisconnectTimer != nil {
			gs.DisconnectTimer.Stop()
			gs.DisconnectTimer = nil
			log.Printf("[RECONNECT] Disconnect timer stopped for game %s", gs.GameID)
			
			opponentID := gs.GetOpponentID(userID)
			if opponentID != nil && !gs.Game.IsFinished() {
				gs.sendEvent(*opponentID, domain.ServerMessage{
					Type: "opponent_reconnected",
				})
			}
		}
	}

	yourPlayer := domain.Player2
	if gs.Player1ID == userID {
		yourPlayer = domain.Player1
	}

	var winner string
	var reason string
	var allowRematch *bool
	var rematchRequester string

	if gs.Game.IsFinished() {
		switch gs.Game.Status {
		case domain.StatusDraw:
			winner = "draw"
		case domain.StatusWon:
			if gs.Reason == "abandonment" || gs.Reason == "timeout" {
				winner = gs.GetUsername(gs.Game.Winner)
				if winner == "" {
					winnerID := gs.GetOpponentID(userID)
					if winnerID != nil {
						winner = gs.GetUsernameByUserID(*winnerID)
					}
				}
			} else {
				winner = gs.GetUsername(gs.Game.Winner)
			}
		}
		reason = gs.Reason
		allowR := (gs.Reason != "abandonment" && gs.Reason != "timeout")
		allowRematch = &allowR
	}

	if gs.RematchRequester != nil {
		rematchRequester = gs.GetUsernameByUserID(*gs.RematchRequester)
	}

	// Send current state to reconnected user
	gs.sendEvent(userID, domain.ServerMessage{
		Type:             "game_state", 
		GameID:           gs.GameID,
		Opponent:         gs.GetOpponentUsername(userID),
		YourPlayer:       int(yourPlayer),
		CurrentTurn:      int(gs.Game.CurrentPlayer),
		Board:            gs.Game.Board,
		Winner:           winner,
		Reason:           reason,
		AllowRematch:     allowRematch,
		RematchRequester: rematchRequester,
	})

	return nil
}

func (gs *GameSession) HandleRematchRequest(userID int64, sessionManager *SessionManager) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if !gs.Game.IsFinished() {
		return fmt.Errorf("game is not finished")
	}

	if gs.RematchRequester != nil {
		return fmt.Errorf("rematch already requested")
	}

	gs.RematchRequester = &userID
	requesterName := gs.GetUsernameByUserID(userID)

	opponentID := gs.GetOpponentID(userID)
	if opponentID == nil {
		// Rematch with bot?
		if gs.IsBot() {
			// Instant rematch for bot
			gs.mu.Unlock() 
			sessionManager.CreateRematchSession(gs.Player1ID, gs.Player1Username, nil, "", gs.BotDifficulty)
			gs.mu.Lock()
			return nil
		}
		return fmt.Errorf("cannot rematch with no opponent")
	}

	gs.broadcastEvent(domain.GameEvent{
		Type: domain.EventInfo,
		Recipients: []int64{*opponentID},
		Payload: domain.ServerMessage{
			Type: "rematch_requested",
			Message: fmt.Sprintf("%s wants a rematch!", requesterName),
		},
	})
	
	gs.startRematchTimer()
	return nil
}

func (gs *GameSession) HandleRematchResponse(userID int64, accept bool, sessionManager *SessionManager) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.RematchRequester == nil {
		return fmt.Errorf("no rematch requested")
	}

	if *gs.RematchRequester == userID {
		return fmt.Errorf("cannot respond to own request")
	}

	if !accept {
		gs.RematchRequester = nil
		if gs.RematchRequestTimer != nil {
			gs.RematchRequestTimer.Stop()
			gs.RematchRequestTimer = nil
		}
		
		gs.broadcastEvent(domain.GameEvent{
			Type: domain.EventInfo,
			Recipients: []int64{*gs.RematchRequester},
			Payload: domain.ServerMessage{
				Type: "rematch_declined",
				Message: "Opponent declined rematch",
			},
		})
		return nil
	}

	// Accepted!
	p1ID := gs.Player1ID
	p1Name := gs.Player1Username
	p2ID := gs.Player2ID
	p2Name := gs.Player2Username
	botDiff := gs.BotDifficulty

	// Create new session
	gs.mu.Unlock() // Unlock to call SessionManager (which locks)
	sessionManager.CreateRematchSession(p1ID, p1Name, p2ID, p2Name, botDiff)
	gs.mu.Lock()
	
	return nil
}

// TerminateSessionWithReason ends game immediately (abandonment/surrender)
func (gs *GameSession) TerminateSessionWithReason(userID int64, reason string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.Game.IsFinished() {
		return
	}

	gs.FinishedAt = time.Now()
	gs.Reason = reason
	gs.Game.Status = domain.StatusWon // For abandonment
	
	winnerID := gs.GetOpponentID(userID)
	winnerName := gs.GetOpponentUsername(userID)

	// Broadacst
	recipients := gs.getAllParticipants()
	allowRematch := false
	
	gs.broadcastEvent(domain.GameEvent{
		Type: domain.EventGameOver,
		Recipients: recipients,
		Payload: domain.ServerMessage{
			Type: "game_over",
			Winner: winnerName,
			Reason: reason,
			Board: gs.Game.Board,
			AllowRematch: &allowRematch,
		},
	})
	
	duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())
	gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username, gs.Player2ID, gs.Player2Username,
		winnerID, winnerName, reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))
}

func (gs *GameSession) GetPlayerID(userID int64) (domain.PlayerID, bool) {
	playerID, exists := gs.PlayerMapping[userID]
	return playerID, exists
}
func (gs *GameSession) GetUsername(playerID domain.PlayerID) string {
	if playerID == domain.Player1 { return gs.Player1Username }
	return gs.Player2Username
}
func (gs *GameSession) GetUsernameByUserID(userID int64) string {
	if userID == gs.Player1ID { return gs.Player1Username }
	return gs.Player2Username
}
func (gs *GameSession) GetOpponentUsername(userID int64) string {
	if userID == gs.Player1ID { return gs.Player2Username }
	return gs.Player1Username
}
func (gs *GameSession) GetOpponentID(userID int64) *int64 {
	if userID == gs.Player1ID { return gs.Player2ID }
	return &gs.Player1ID
}
func (gs *GameSession) IsBot() bool { return gs.Player2ID == nil }
func (gs *GameSession) AddSpectator(userID int64) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.Spectators[userID] = true
	
	// Send initial state
	gs.sendEvent(userID, domain.ServerMessage{
		Type:        "spectate_start",
		GameID:      gs.GameID,
		Player1:     gs.Player1Username,
		Player2:     gs.Player2Username,
		CurrentTurn: int(gs.Game.CurrentPlayer),
		Board:       gs.Game.Board,
	})
}
func (gs *GameSession) RemoveSpectator(userID int64) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	delete(gs.Spectators, userID)
}
func (gs *GameSession) saveGameAsync(gameID string, p1ID int64, p1User string,
	p2ID *int64, p2User string, winnerID *int64, winnerUser string,
	reason string, moves, duration int, created, finished time.Time, boardState [][]int) {
	go func() {
		err := gs.repo.SaveGame(gameID, p1ID, p1User, p2ID, p2User,
			winnerID, winnerUser, reason, moves, duration, created, finished, boardState)
		if err != nil {
			log.Printf("[GAME] Error saving game %s: %v", gameID, err)
		}
	}()
}
func (gs *GameSession) startTurnTimer() {
	if gs.TurnTimer != nil {
		gs.TurnTimer.Stop()
	}
	gs.TurnTimer = time.AfterFunc(15*time.Minute, func() {
		gs.mu.Lock()
		defer gs.mu.Unlock()
		if gs.Game.IsFinished() { return }
		
		currentPlayer := gs.Game.CurrentPlayer
		var winnerName string
		if currentPlayer == domain.Player1 {
			if gs.Player2ID != nil { winnerName = gs.Player2Username } else { winnerName = "Bot" }
		} else {
			winnerName = gs.Player1Username
		}
		
		gs.Game.Status = domain.StatusWon
		allowRematch := true
		
		gs.broadcastEvent(domain.GameEvent{
			Type: domain.EventGameOver,
			Recipients: gs.getAllParticipants(),
			Payload: domain.ServerMessage{
				Type: "game_over",
				Winner: winnerName,
				Reason: "timeout",
				Board: gs.Game.Board,
				AllowRematch: &allowRematch,
			},
		})
		gs.StartPostGameTimer()
	})
}
func (gs *GameSession) StartPostGameTimer() {
	if gs.PostGameTimer != nil { gs.PostGameTimer.Stop() }
	gs.PostGameTimer = time.AfterFunc(30*time.Second, func() {
		// Cleanup
	})
}
func (gs *GameSession) startRematchTimer() {
	if gs.RematchRequestTimer != nil { gs.RematchRequestTimer.Stop() }
	gs.RematchRequestTimer = time.AfterFunc(10*time.Second, func() {
		gs.mu.Lock()
		defer gs.mu.Unlock()
		if gs.RematchRequester == nil { return }
		gs.RematchRequester = nil
		// Notify timeout
	})
}
func (gs *GameSession) CancelRematchRequest() {
	gs.RematchRequester = nil
	if gs.RematchRequestTimer != nil {
		gs.RematchRequestTimer.Stop()
		gs.RematchRequestTimer = nil
	}
}
func (gs *GameSession) cleanupTimers() {
	if gs.TurnTimer != nil { gs.TurnTimer.Stop() }
	if gs.DisconnectTimer != nil { gs.DisconnectTimer.Stop() }
	if gs.PostGameTimer != nil { gs.PostGameTimer.Stop() }
	if gs.RematchRequestTimer != nil { gs.RematchRequestTimer.Stop() }
}
func (gs *GameSession) HandleGetState(userID int64) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	
	yourPlayer := domain.Player2
	if gs.Player1ID == userID {
		yourPlayer = domain.Player1
	}

	var winner string
	var reason string
	var allowRematch *bool
	var rematchRequester string

	if gs.Game.IsFinished() {
		switch gs.Game.Status {
		case domain.StatusDraw:
			winner = "draw"
		case domain.StatusWon:
			if gs.Reason == "abandonment" || gs.Reason == "timeout" {
				winner = gs.GetUsername(gs.Game.Winner)
				if winner == "" {
					winnerID := gs.GetOpponentID(userID)
					if winnerID != nil {
						winner = gs.GetUsernameByUserID(*winnerID)
					}
				}
			} else {
				winner = gs.GetUsername(gs.Game.Winner)
			}
		}
		reason = gs.Reason
		allowR := (gs.Reason != "abandonment" && gs.Reason != "timeout")
		allowRematch = &allowR
	}

	if gs.RematchRequester != nil {
		rematchRequester = gs.GetUsernameByUserID(*gs.RematchRequester)
	}

	gs.sendEvent(userID, domain.ServerMessage{
		Type:             "game_state",
		GameID:           gs.GameID,
		Opponent:         gs.GetOpponentUsername(userID),
		YourPlayer:       int(yourPlayer),
		CurrentTurn:      int(gs.Game.CurrentPlayer),
		Board:            gs.Game.Board,
		Winner:           winner,
		Reason:           reason,
		AllowRematch:     allowRematch,
		RematchRequester: rematchRequester,
	})
}

// Add CreateRematchSession to SessionManager
func (sm *SessionManager) CreateRematchSession(p1ID int64, p1User string, p2ID *int64, p2User string, botDiff string) *GameSession {
	// Logic to start new game
	session := sm.CreateSession(p1ID, p1User, p2ID, p2User, botDiff)
	return session
}

// Convert board to ints (helper)
func convertBoardToInts(board [][]domain.PlayerID) [][]int {
	rows := len(board)
	cols := len(board[0])
	result := make([][]int, rows)
	for i := range board {
		result[i] = make([]int, cols)
		for j := range board[i] {
			result[i][j] = int(board[i][j])
		}
	}
	return result
}
