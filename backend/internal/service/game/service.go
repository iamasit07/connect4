package game

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/internal/config"
	"github.com/iamasit07/4-in-a-row/backend/internal/domain"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/bot"
	"github.com/iamasit07/4-in-a-row/backend/pkg/uid"
)

type GameSession struct {
	GameID              string
	Player1ID           int64
	Player1Username     string
	Player2ID           *int64 // NULL for BOT
	Player2Username     string
	Game                *domain.Game
	PlayerMapping       map[int64]domain.PlayerID // userID → PlayerID
	Reason              string
	DisconnectedPlayers []int64 // userIDs
	ReconnectTimer      *time.Timer
	CreatedAt           time.Time
	FinishedAt          time.Time
	BotDifficulty       string // "easy", "medium", "hard"
	PostGameTimer       *time.Timer // 30-second window for rematch after game ends
	RematchRequester    *int64      // userID of player who requested rematch
	RematchRequestTimer *time.Timer // 10-second window to accept rematch request
	mu                  sync.Mutex
	repo                GameRepository
}

type ConnectionManagerInterface interface {
	SendMessage(userID int64, message domain.ServerMessage) error
	RemoveConnection(userID int64)
}

type GameRepository interface {
	SaveGame(gameID string, player1ID int64, player1Username string, player2ID *int64, player2Username string, winnerID *int64, winnerUsername string, reason string, totalMoves, durationSeconds int, createdAt, finishedAt time.Time, boardState [][]int) error
}

// SessionManager manages active game sessions
type SessionManager struct {
	Session    map[string]*GameSession // gameID → GameSession
	UserToGame map[int64]string        // userID → gameID (for quick lookup)
	mu         sync.RWMutex
	repo       GameRepository
}

func NewSessionManager(repo GameRepository) *SessionManager {
	return &SessionManager{
		Session:    make(map[string]*GameSession),
		UserToGame: make(map[int64]string),
		repo:       repo,
	}
}

func (sm *SessionManager) CreateSession(player1ID int64, player1Username string, player2ID *int64, player2Username string, botDifficulty string, conn ConnectionManagerInterface) *GameSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := NewGameSession(player1ID, player1Username, player2ID, player2Username, botDifficulty, conn, sm.repo)
	gameID := session.GameID
	sm.Session[gameID] = session
	sm.UserToGame[player1ID] = gameID

	if player2Username != domain.BotUsername && player2ID != nil {
		sm.UserToGame[*player2ID] = gameID
	}

	log.Printf("[SESSION] Created session %s: %s (ID: %d) vs %s (ID: %v)\n",
		gameID, player1Username, player1ID, player2Username, player2ID)
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

	log.Printf("[SESSION] Removing session %s", gameID)

	delete(sm.UserToGame, session.Player1ID)
	if session.Player2ID != nil {
		delete(sm.UserToGame, *session.Player2ID)
	}

	delete(sm.Session, gameID)

	return nil
}

func (sm *SessionManager) HasActiveGame(userID int64) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	gameID, exists := sm.UserToGame[userID]
	if !exists {
		return false
	}

	session, exists := sm.Session[gameID]
	if !exists {
		return false
	}

	return !session.Game.IsFinished()
}

func (sm *SessionManager) TerminateSessionForUser(userID int64, conn ConnectionManagerInterface) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Need to look up session manually since getSessionByUserIDLocked logic relies on UserToGame map
	gameID, ok := sm.UserToGame[userID]
	if !ok {
		return
	}
	session, ok := sm.Session[gameID]
	if !ok {
		return
	}

	log.Printf("[SESSION] Force terminating session %s for user %d", session.GameID, userID)
	
	// Notify users
	conn.SendMessage(session.Player1ID, domain.ServerMessage{Type: "game_ended", Reason: "Force termination"})
	if session.Player2ID != nil {
		conn.SendMessage(*session.Player2ID, domain.ServerMessage{Type: "game_ended", Reason: "Force termination"})
	}
	
	sm.removeSessionLocked(session.GameID)
}

func (gs *GameSession) GetPlayerID(userID int64) (domain.PlayerID, bool) {
	playerID, exists := gs.PlayerMapping[userID]
	return playerID, exists
}

func (gs *GameSession) GetUsername(playerID domain.PlayerID) string {
	if playerID == domain.Player1 {
		return gs.Player1Username
	}
	return gs.Player2Username
}

func (gs *GameSession) GetUsernameByUserID(userID int64) string {
	if userID == gs.Player1ID {
		return gs.Player1Username
	}
	return gs.Player2Username
}

func (gs *GameSession) GetOpponentUsername(userID int64) string {
	if userID == gs.Player1ID {
		return gs.Player2Username
	}
	return gs.Player1Username
}

func (gs *GameSession) GetOpponentID(userID int64) *int64 {
	if userID == gs.Player1ID {
		return gs.Player2ID
	}
	return &gs.Player1ID
}

func (gs *GameSession) IsBot() bool {
	return gs.Player2Username == domain.BotUsername
}

func (gs *GameSession) cleanupConnections(conn ConnectionManagerInterface) {
	conn.RemoveConnection(gs.Player1ID)
	if !gs.IsBot() && gs.Player2ID != nil {
		conn.RemoveConnection(*gs.Player2ID)
	}
}

// Helper function to convert PlayerID board to int board for database storage
func convertBoardToInts(board [][]domain.PlayerID) [][]int {
	intBoard := make([][]int, len(board))
	for i := range board {
		intBoard[i] = make([]int, len(board[i]))
		for j := range board[i] {
			intBoard[i][j] = int(board[i][j])
		}
	}
	return intBoard
}

// Saves game data to database in background to avoid blocking game_over messages
func (gs *GameSession) saveGameAsync(gameID string, p1ID int64, p1User string,
	p2ID *int64, p2User string, winnerID *int64, winnerUser string,
	reason string, moves, duration int, created, finished time.Time, boardState [][]int) {

	go func() {
		err := gs.repo.SaveGame(gameID, p1ID, p1User, p2ID, p2User,
			winnerID, winnerUser, reason, moves, duration, created, finished, boardState)
		if err != nil {
			log.Printf("[GAME] Error saving game %s: %v", gameID, err)
		} else {
			log.Printf("[GAME] Game %s saved successfully", gameID)
		}
	}()
}

func NewGameSession(player1ID int64, player1Username string, player2ID *int64, player2Username string, botDifficulty string, conn ConnectionManagerInterface, repo GameRepository) *GameSession {
	gameID := uid.GenerateGameID()
	newGame := (&domain.Game{}).NewGame()

	mapping := make(map[int64]domain.PlayerID)
	mapping[player1ID] = domain.Player1
	if player2ID != nil {
		mapping[*player2ID] = domain.Player2
	}

	gs := &GameSession{
		GameID:          gameID,
		Player1ID:       player1ID,
		Player1Username: player1Username,
		Player2ID:       player2ID,
		Player2Username: player2Username,
		Game:            newGame,
		PlayerMapping:   mapping,
		BotDifficulty:   botDifficulty,
		CreatedAt:       time.Now(),
		mu:              sync.Mutex{},
		repo:            repo,
	}

	conn.SendMessage(player1ID, domain.ServerMessage{
		Type:        "game_start",
		GameID:      gs.GameID,
		Opponent:    player2Username,
		YourPlayer:  int(domain.Player1),
		CurrentTurn: int(gs.Game.CurrentPlayer),
		Board:       gs.Game.Board,
	})

	if player2Username != domain.BotUsername && player2ID != nil {
		conn.SendMessage(*player2ID, domain.ServerMessage{
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
				count++
			}
		} else {
			if now.Sub(session.CreatedAt) > 24*time.Hour {
				delete(sm.Session, gameID)
				delete(sm.UserToGame, session.Player1ID)
				if session.Player2ID != nil {
					delete(sm.UserToGame, *session.Player2ID)
				}
				count++
			}
		}
	}
	
	if count > 0 {
		log.Printf("[SESSION] Memory cleanup: Removed %d stale game sessions", count)
	}
}

func (gs *GameSession) HandleMove(userID int64, column int, conn ConnectionManagerInterface) error {
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

	if gs.Game.Status == domain.StatusWon {
		gs.FinishedAt = time.Now()
		winnerUsername := gs.GetUsername(gs.Game.Winner)
		winnerID := userID
		gs.Reason = "connect_four"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		allowRematch := true
		gameOverMsg := domain.ServerMessage{
			Type:         "game_over",
			Winner:       winnerUsername,
			Reason:       gs.Reason,
			Board:        gs.Game.Board,
			AllowRematch: &allowRematch,
		}

		conn.SendMessage(gs.Player1ID, gameOverMsg)
		if !gs.IsBot() && gs.Player2ID != nil {
			conn.SendMessage(*gs.Player2ID, gameOverMsg)
		}

		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			gs.Player2ID, gs.Player2Username, &winnerID, winnerUsername,
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))

		// Start 30-second post-game timer for rematch opportunity
		gs.StartPostGameTimer(conn)

		return nil
	}

	if gs.Game.Status == domain.StatusDraw {
		gs.FinishedAt = time.Now()
		gs.Reason = "draw"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		allowRematch := true
		gameOverMsg := domain.ServerMessage{
			Type:         "game_over",
			Winner:       "draw",
			Reason:       "draw",
			Board:        gs.Game.Board,
			AllowRematch: &allowRematch,
		}

		conn.SendMessage(gs.Player1ID, gameOverMsg)
		if !gs.IsBot() && gs.Player2ID != nil {
			conn.SendMessage(*gs.Player2ID, gameOverMsg)
		}

		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			gs.Player2ID, gs.Player2Username, nil, "draw",
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))

		// Start 30-second post-game timer for rematch opportunity
		gs.StartPostGameTimer(conn)

		return nil
	}

	moveMadeMsg := domain.ServerMessage{
		Type:     "move_made",
		Column:   column,
		Row:      row,
		Player:   int(playerID),
		Board:    gs.Game.Board,
		NextTurn: int(gs.Game.CurrentPlayer),
	}

	conn.SendMessage(gs.Player1ID, moveMadeMsg)
	if !gs.IsBot() && gs.Player2ID != nil {
		conn.SendMessage(*gs.Player2ID, moveMadeMsg)
	}
	
	// TRIGGER BOT MOVE if applicable
	if gs.IsBot() && gs.Game.CurrentPlayer == domain.Player2 {
		go func() {
			// Small delay to feel natural
			time.Sleep(500 * time.Millisecond)
			if err := gs.HandleBotMove(conn); err != nil {
				log.Printf("[BOT] Error handling bot move: %v", err)
			}
		}()
	}

	return nil
}

func (gs *GameSession) HandleBotMove(conn ConnectionManagerInterface) error {
	// Acquire lock since this is entry point from goroutine
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Verify it's actually bot's turn (race condition check)
	if gs.Game.CurrentPlayer != domain.Player2 || !gs.IsBot() || gs.Game.IsFinished() {
		return nil
	}

	difficulty := gs.BotDifficulty
	if difficulty == "" {
		difficulty = "medium" // Default to medium if not set
	}

	botColumn := bot.CalculateBestMove(gs.Game.Board, domain.Player2, difficulty)
	botRow, err := gs.Game.MakeMove(domain.Player2, botColumn)
	if err != nil {
		return err
	}

	if gs.Game.Status == domain.StatusWon {
		gs.FinishedAt = time.Now()
		gs.Reason = "connect_four"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		allowRematch := true // Bot games allow infinite rematches
		gameOverMsg := domain.ServerMessage{
			Type:         "game_over",
			Winner:       domain.BotUsername,
			Reason:       gs.Reason,
			Board:        gs.Game.Board,
			AllowRematch: &allowRematch,
		}
		conn.SendMessage(gs.Player1ID, gameOverMsg)

		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			nil, domain.BotUsername, nil, domain.BotUsername,
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))

		// Start 30-second post-game timer for rematch opportunity
		gs.StartPostGameTimer(conn)

		return nil
	}

	if gs.Game.Status == domain.StatusDraw {
		gs.FinishedAt = time.Now()
		gs.Reason = "draw"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		allowRematch := true // Bot games allow infinite rematches
		gameOverMsg := domain.ServerMessage{
			Type:         "game_over",
			Winner:       "draw",
			Reason:       "draw",
			Board:        gs.Game.Board,
			AllowRematch: &allowRematch,
		}
		conn.SendMessage(gs.Player1ID, gameOverMsg)

		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			nil, domain.BotUsername, nil, "draw",
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))

		// Start 30-second post-game timer for rematch opportunity
		gs.StartPostGameTimer(conn)

		return nil
	}

	botMoveMsg := domain.ServerMessage{
		Type:     "move_made",
		Column:   botColumn,
		Row:      botRow,
		Player:   int(domain.Player2),
		Board:    gs.Game.Board,
		NextTurn: int(gs.Game.CurrentPlayer),
	}
	conn.SendMessage(gs.Player1ID, botMoveMsg)

	return nil
}

func (gs *GameSession) HandleDisconnect(userID int64, conn ConnectionManagerInterface, sessionManager *SessionManager) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// If game is finished and there's a pending rematch request, cancel it
	if gs.Game.IsFinished() {
		if gs.RematchRequester != nil {
			// If the disconnecting player is the requester or the recipient, cancel rematch
			gs.CancelRematchRequest(conn)

			// Clean up connections
			gs.cleanupConnections(conn)

			// Stop post-game timer
			if gs.PostGameTimer != nil {
				gs.PostGameTimer.Stop()
				gs.PostGameTimer = nil
			}
		}
		return nil
	}

	username := gs.GetUsernameByUserID(userID)
	log.Printf("[DISCONNECT] Player %s (ID: %d) disconnected from game %s", username, userID, gs.GameID)

	gs.DisconnectedPlayers = append(gs.DisconnectedPlayers, userID)

	opponentID := gs.GetOpponentID(userID)
	if !gs.IsBot() && opponentID != nil {
		conn.SendMessage(*opponentID, domain.ServerMessage{
			Type:    "opponent_disconnected",
			Message: fmt.Sprintf("%s disconnected", username),
		})
	}

	// Check if both REAL players disconnected (ignore bot)
	// Bot never disconnects, so only check for 2 disconnected in PvP games
	if !gs.IsBot() && len(gs.DisconnectedPlayers) == 2 {
		log.Printf("[DISCONNECT] Both players disconnected from game %s - game ends as draw", gs.GameID)

		if gs.ReconnectTimer != nil {
			gs.ReconnectTimer.Stop()
			gs.ReconnectTimer = nil
		}

		gs.FinishedAt = time.Now()
		gs.Reason = "both_disconnected"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			gs.Player2ID, gs.Player2Username, nil, "draw",
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))

		gs.cleanupConnections(conn)
		sessionManager.RemoveSession(gs.GameID)
		return nil
	}

	if gs.ReconnectTimer == nil {
		gs.ReconnectTimer = time.AfterFunc(config.AppConfig.ReconnectTimeout, func() {
			gs.HandleReconnectTimeout(userID, conn, sessionManager)
		})
		log.Printf("[DISCONNECT] Started 30s reconnect timer for game %s", gs.GameID)
	}

	return nil
}

func (gs *GameSession) HandleReconnect(userID int64, conn ConnectionManagerInterface) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	username := gs.GetUsernameByUserID(userID)
	log.Printf("[RECONNECT] Player %s (ID: %d) reconnecting to game %s", username, userID, gs.GameID)

	newDisconnected := []int64{}
	for _, uid := range gs.DisconnectedPlayers {
		if uid != userID {
			newDisconnected = append(newDisconnected, uid)
		}
	}
	gs.DisconnectedPlayers = newDisconnected

	if gs.ReconnectTimer != nil {
		gs.ReconnectTimer.Stop()
		gs.ReconnectTimer = nil
	}

	playerID, _ := gs.GetPlayerID(userID)
	conn.SendMessage(userID, domain.ServerMessage{
		Type:        "reconnect_success",
		GameID:      gs.GameID,
		Opponent:    gs.GetOpponentUsername(userID),
		YourPlayer:  int(playerID),
		CurrentTurn: int(gs.Game.CurrentPlayer),
		Board:       gs.Game.Board,
	})

	opponentID := gs.GetOpponentID(userID)
	if !gs.IsBot() && opponentID != nil {
		conn.SendMessage(*opponentID, domain.ServerMessage{
			Type:    "opponent_reconnected",
			Message: fmt.Sprintf("%s reconnected", username),
		})
	}

	return nil
}

func (gs *GameSession) HandleReconnectTimeout(userID int64, conn ConnectionManagerInterface, sessionManager *SessionManager) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	_, exists := sessionManager.GetSessionByGameID(gs.GameID)
	if !exists {
		log.Printf("[TIMEOUT] Game %s already removed, ignoring timeout", gs.GameID)
		return
	}

	if gs.Game.IsFinished() {
		log.Printf("[TIMEOUT] Game %s already finished, ignoring reconnect timeout", gs.GameID)
		return
	}

	username := gs.GetUsernameByUserID(userID)
	opponentID := gs.GetOpponentID(userID)
	opponentUsername := gs.GetOpponentUsername(userID)

	log.Printf("[TIMEOUT] Player %s (ID: %d) reconnect timeout in game %s", username, userID, gs.GameID)

	gs.FinishedAt = time.Now()
	gs.Reason = "timeout"

	duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

	err := gs.repo.SaveGame(
		gs.GameID,
		gs.Player1ID, gs.Player1Username,
		gs.Player2ID, gs.Player2Username,
		opponentID, opponentUsername,
		gs.Reason,
		gs.Game.MoveCount,
		duration,
		gs.CreatedAt,
		gs.FinishedAt, convertBoardToInts(gs.Game.Board),
	)
	if err != nil {
		log.Printf("[GAME] Error saving game: %v", err)
	}

	allowRematch := false
	gameOverMsg := domain.ServerMessage{
		Type:         "game_over",
		Winner:       opponentUsername,
		Reason:       "timeout",
		AllowRematch: &allowRematch,
	}

	conn.SendMessage(userID, gameOverMsg)
	if !gs.IsBot() && opponentID != nil {
		conn.SendMessage(*opponentID, gameOverMsg)
	}

	conn.RemoveConnection(userID)
	if !gs.IsBot() && opponentID != nil {
		conn.RemoveConnection(*opponentID)
	}

	sessionManager.RemoveSession(gs.GameID)
}

func (gs *GameSession) TerminateSessionByAbandonment(abandoningUserID int64, conn ConnectionManagerInterface) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.ReconnectTimer != nil {
		gs.ReconnectTimer.Stop()
		gs.ReconnectTimer = nil
		log.Printf("[CLEANUP] Stopped reconnect timer for game %s (player abandoning)", gs.GameID)
	}

	abandoningUsername := gs.GetUsernameByUserID(abandoningUserID)
	opponentUsername := gs.GetOpponentUsername(abandoningUserID)
	opponentID := gs.GetOpponentID(abandoningUserID)

	log.Printf("[TERMINATE] Game %s terminated by abandonment from %s (ID: %d)",
		gs.GameID, abandoningUsername, abandoningUserID)

	gs.FinishedAt = time.Now()
	gs.Reason = "surrender"

	duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

	allowRematch := false // Abandoned games don't allow rematch
	gameOverMsg := domain.ServerMessage{
		Type:         "game_over",
		Winner:       opponentUsername,
		Reason:       "surrender",
		AllowRematch: &allowRematch,
	}

	// Notify BOTH players
	conn.SendMessage(abandoningUserID, gameOverMsg)
	if !gs.IsBot() && opponentID != nil {
		conn.SendMessage(*opponentID, gameOverMsg)
	}

	gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
		gs.Player2ID, gs.Player2Username, opponentID, opponentUsername,
		gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))

	return nil
}

// StartPostGameTimer starts a 30-second timer to allow rematch requests
func (gs *GameSession) StartPostGameTimer(conn ConnectionManagerInterface) {
	gs.PostGameTimer = time.AfterFunc(30*time.Second, func() {
		gs.mu.Lock()
		defer gs.mu.Unlock()

		log.Printf("[POST_GAME] 30-second window expired for game %s, closing connections", gs.GameID)

		// Stop any pending rematch request timer
		if gs.RematchRequestTimer != nil {
			gs.RematchRequestTimer.Stop()
			gs.RematchRequestTimer = nil
		}

		// Clean up connections silently
		gs.cleanupConnections(conn)
	})
	log.Printf("[POST_GAME] Started 30-second post-game timer for game %s", gs.GameID)
}

// HandleRematchRequest processes a rematch request from a player
func (gs *GameSession) HandleRematchRequest(userID int64, conn ConnectionManagerInterface, sessionManager *SessionManager) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Verify game is finished
	if !gs.Game.IsFinished() {
		return fmt.Errorf("cannot request rematch - game still in progress")
	}

	// Verify requester is a player in this game
	_, isPlayer := gs.GetPlayerID(userID)
	if !isPlayer {
		return fmt.Errorf("you are not a player in this game")
	}

	// Check if there's already a pending rematch request
	if gs.RematchRequester != nil {
		return fmt.Errorf("rematch already requested")
	}

	requesterUsername := gs.GetUsernameByUserID(userID)

	// For bot games, immediately create rematch
	if gs.IsBot() {
		log.Printf("[REMATCH] Bot game rematch requested by %s (ID: %d)", requesterUsername, userID)
		return gs.CreateRematchGame(conn, sessionManager)
	}

	// For PvP games, notify opponent and start 10-second timer
	gs.RematchRequester = &userID
	opponentID := gs.GetOpponentID(userID)

	if opponentID == nil {
		return fmt.Errorf("opponent not found")
	}

	log.Printf("[REMATCH] %s (ID: %d) requested rematch in game %s", requesterUsername, userID, gs.GameID)

	// Send rematch request to opponent
	conn.SendMessage(*opponentID, domain.ServerMessage{
		Type:             "rematch_request",
		RematchRequester: requesterUsername,
		RematchTimeout:   10,
	})

	// Start 10-second timer for acceptance
	gs.RematchRequestTimer = time.AfterFunc(10*time.Second, func() {
		gs.mu.Lock()
		defer gs.mu.Unlock()

		log.Printf("[REMATCH] Request timeout for game %s", gs.GameID)

		allowRematch := false // Timer ended, can't rematch
		// Notify both players that request timed out
		conn.SendMessage(userID, domain.ServerMessage{
			Type:         "rematch_timeout",
			Message:      "Rematch request timed out",
			AllowRematch: &allowRematch,
		})

		conn.SendMessage(*opponentID, domain.ServerMessage{
			Type:         "rematch_timeout",
			Message:      "Rematch request timed out",
			AllowRematch: &allowRematch,
		})

		// Clean up
		gs.RematchRequester = nil
		gs.cleanupConnections(conn)

		// Stop post-game timer if still running
		if gs.PostGameTimer != nil {
			gs.PostGameTimer.Stop()
			gs.PostGameTimer = nil
		}
	})

	return nil
}

// HandleRematchResponse processes accept/decline response to a rematch request
func (gs *GameSession) HandleRematchResponse(userID int64, response string, conn ConnectionManagerInterface, sessionManager *SessionManager) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Verify there's a pending rematch request
	if gs.RematchRequester == nil {
		return fmt.Errorf("no pending rematch request")
	}

	// Verify responder is the opponent (not the requester)
	if userID == *gs.RematchRequester {
		return fmt.Errorf("cannot respond to your own rematch request")
	}

	// Verify responder is a player in this game
	_, isPlayer := gs.GetPlayerID(userID)
	if !isPlayer {
		return fmt.Errorf("you are not a player in this game")
	}

	// Stop the rematch request timer
	if gs.RematchRequestTimer != nil {
		gs.RematchRequestTimer.Stop()
		gs.RematchRequestTimer = nil
	}

	requesterID := *gs.RematchRequester
	responderUsername := gs.GetUsernameByUserID(userID)

	if response == "accept" {
		log.Printf("[REMATCH] %s (ID: %d) accepted rematch in game %s", responderUsername, userID, gs.GameID)

		// Notify both players
		conn.SendMessage(requesterID, domain.ServerMessage{
			Type:    "rematch_accepted",
			Message: "Rematch accepted",
		})

		conn.SendMessage(userID, domain.ServerMessage{
			Type:    "rematch_accepted",
			Message: "Rematch accepted",
		})

		return gs.CreateRematchGame(conn, sessionManager)
	} else {
		log.Printf("[REMATCH] %s (ID: %d) declined rematch in game %s", responderUsername, userID, gs.GameID)

		allowRematch := false
		// Notify players
		conn.SendMessage(requesterID, domain.ServerMessage{
			Type:         "rematch_declined",
			Message:      "Rematch request declined",
			AllowRematch: &allowRematch,
		})
		
		conn.SendMessage(userID, domain.ServerMessage{
			Type:         "rematch_declined", // Notify decliner too so UI updates
			Message:      "Rematch request declined",
			AllowRematch: &allowRematch,
		})

		gs.RematchRequester = nil
		gs.cleanupConnections(conn)
		
		// Stop post-game timer
		if gs.PostGameTimer != nil {
			gs.PostGameTimer.Stop()
			gs.PostGameTimer = nil
		}
	}

	return nil
}

func (gs *GameSession) CancelRematchRequest(conn ConnectionManagerInterface) {
	if gs.RematchRequester == nil {
		return
	}

	log.Printf("[REMATCH] Cancelling rematch request for game %s", gs.GameID)

	// Stop timers
	if gs.RematchRequestTimer != nil {
		gs.RematchRequestTimer.Stop()
		gs.RematchRequestTimer = nil
	}

	// Notify opponent
	requesterID := *gs.RematchRequester
	opponentID := gs.GetOpponentID(requesterID)

	if opponentID != nil && !gs.IsBot() {
		conn.SendMessage(*opponentID, domain.ServerMessage{
			Type:    "rematch_cancelled",
			Message: "Rematch request cancelled",
		})
	}

	gs.RematchRequester = nil
}

func (gs *GameSession) CreateRematchGame(conn ConnectionManagerInterface, sessionManager *SessionManager) error {
	// Stop existing timers
	if gs.PostGameTimer != nil {
		gs.PostGameTimer.Stop()
		gs.PostGameTimer = nil
	}
	if gs.RematchRequestTimer != nil {
		gs.RematchRequestTimer.Stop()
		gs.RematchRequestTimer = nil
	}

	log.Printf("[REMATCH] Starting new game for %s and %s", gs.Player1Username, gs.Player2Username)

	// Keep same player order
	newP1ID := gs.Player1ID
	newP1Username := gs.Player1Username
	newP2ID := gs.Player2ID
	newP2Username := gs.Player2Username
	
	sessionManager.RemoveSession(gs.GameID)
	sessionManager.CreateSession(newP1ID, newP1Username, newP2ID, newP2Username, gs.BotDifficulty, conn)
	
	return nil
}