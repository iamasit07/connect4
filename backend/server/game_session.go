package server

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/bot"
	"github.com/iamasit07/4-in-a-row/backend/config"
	"github.com/iamasit07/4-in-a-row/backend/db"
	"github.com/iamasit07/4-in-a-row/backend/game"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

type ConnectionManagerInterface interface {
	SendMessage(userID int64, message models.ServerMessage) error
	RemoveConnection(userID int64)
}

type GameSession struct {
	GameID              string
	Player1ID           int64
	Player1Username     string
	Player2ID           *int64 // NULL for BOT
	Player2Username     string
	Game                *game.Game
	PlayerMapping       map[int64]models.PlayerID // userID → PlayerID
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
}

func (gs *GameSession) GetPlayerID(userID int64) (models.PlayerID, bool) {
	playerID, exists := gs.PlayerMapping[userID]
	return playerID, exists
}

func (gs *GameSession) GetUsername(playerID models.PlayerID) string {
	if playerID == models.Player1 {
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
	return gs.Player2Username == models.BotUsername
}

func (gs *GameSession) cleanupConnections(conn ConnectionManagerInterface) {
	conn.RemoveConnection(gs.Player1ID)
	if !gs.IsBot() && gs.Player2ID != nil {
		conn.RemoveConnection(*gs.Player2ID)
	}
}

// Helper function to convert PlayerID board to int board for database storage
func convertBoardToInts(board [][]models.PlayerID) [][]int {
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
		err := db.SaveGame(gameID, p1ID, p1User, p2ID, p2User,
			winnerID, winnerUser, reason, moves, duration, created, finished, boardState)
		if err != nil {
			log.Printf("[GAME] Error saving game %s: %v", gameID, err)
		} else {
			log.Printf("[GAME] Game %s saved successfully", gameID)
		}
	}()
}

func NewGameSession(player1ID int64, player1Username string, player2ID *int64, player2Username string, botDifficulty string, conn ConnectionManagerInterface) *GameSession {
	gameID := utils.GenerateGameID()
	newGame := (&game.Game{}).NewGame()

	mapping := make(map[int64]models.PlayerID)
	mapping[player1ID] = models.Player1
	if player2ID != nil {
		mapping[*player2ID] = models.Player2
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
	}

	conn.SendMessage(player1ID, models.ServerMessage{
		Type:        "game_start",
		GameID:      gs.GameID,
		Opponent:    player2Username,
		YourPlayer:  int(models.Player1),
		CurrentTurn: int(gs.Game.CurrentPlayer),
		Board:       gs.Game.Board,
	})

	if player2Username != models.BotUsername && player2ID != nil {
		conn.SendMessage(*player2ID, models.ServerMessage{
			Type:        "game_start",
			GameID:      gs.GameID,
			Opponent:    player1Username,
			YourPlayer:  int(models.Player2),
			CurrentTurn: int(gs.Game.CurrentPlayer),
			Board:       gs.Game.Board,
		})
	}

	return gs
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

	if gs.Game.Status == models.StatusWon {
		gs.FinishedAt = time.Now()
		winnerUsername := gs.GetUsername(gs.Game.Winner)
		winnerID := userID
		gs.Reason = "connect_four"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		allowRematch := true
		gameOverMsg := models.ServerMessage{
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

	if gs.Game.Status == models.StatusDraw {
		gs.FinishedAt = time.Now()
		gs.Reason = "draw"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		allowRematch := true
		gameOverMsg := models.ServerMessage{
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

	moveMadeMsg := models.ServerMessage{
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

	if gs.IsBot() && gs.Game.CurrentPlayer == models.Player2 {
		return gs.HandleBotMove(conn)
	}

	return nil
}

func (gs *GameSession) HandleBotMove(conn ConnectionManagerInterface) error {
	difficulty := gs.BotDifficulty
	if difficulty == "" {
		difficulty = "medium" // Default to medium if not set
	}
	botColumn := bot.CalculateBestMove(gs.Game.Board, models.Player2, difficulty)
	botRow, err := gs.Game.MakeMove(models.Player2, botColumn)
	if err != nil {
		return err
	}

	if gs.Game.Status == models.StatusWon {
		gs.FinishedAt = time.Now()
		gs.Reason = "connect_four"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		allowRematch := true // Bot games allow infinite rematches
		gameOverMsg := models.ServerMessage{
			Type:         "game_over",
			Winner:       models.BotUsername,
			Reason:       gs.Reason,
			Board:        gs.Game.Board,
			AllowRematch: &allowRematch,
		}
		conn.SendMessage(gs.Player1ID, gameOverMsg)

		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			nil, models.BotUsername, nil, models.BotUsername,
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))

		// Start 30-second post-game timer for rematch opportunity
		gs.StartPostGameTimer(conn)

		return nil
	}

	if gs.Game.Status == models.StatusDraw {
		gs.FinishedAt = time.Now()
		gs.Reason = "draw"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		allowRematch := true // Bot games allow infinite rematches
		gameOverMsg := models.ServerMessage{
			Type:         "game_over",
			Winner:       "draw",
			Reason:       "draw",
			Board:        gs.Game.Board,
			AllowRematch: &allowRematch,
		}
		conn.SendMessage(gs.Player1ID, gameOverMsg)

		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			nil, models.BotUsername, nil, "draw",
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt, convertBoardToInts(gs.Game.Board))

		// Start 30-second post-game timer for rematch opportunity
		gs.StartPostGameTimer(conn)

		return nil
	}

	botMoveMsg := models.ServerMessage{
		Type:     "move_made",
		Column:   botColumn,
		Row:      botRow,
		Player:   int(models.Player2),
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
		conn.SendMessage(*opponentID, models.ServerMessage{
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
	conn.SendMessage(userID, models.ServerMessage{
		Type:        "reconnect_success",
		GameID:      gs.GameID,
		Opponent:    gs.GetOpponentUsername(userID),
		YourPlayer:  int(playerID),
		CurrentTurn: int(gs.Game.CurrentPlayer),
		Board:       gs.Game.Board,
	})

	opponentID := gs.GetOpponentID(userID)
	if !gs.IsBot() && opponentID != nil {
		conn.SendMessage(*opponentID, models.ServerMessage{
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

	err := db.SaveGame(
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

	allowRematch := true
	gameOverMsg := models.ServerMessage{
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
	gs.Reason = "abandoned"

	duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

	allowRematch := false // Abandoned games don't allow rematch
	gameOverMsg := models.ServerMessage{
		Type:         "game_over",
		Winner:       opponentUsername,
		Reason:       "abandoned",
		AllowRematch: &allowRematch,
	}

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
	conn.SendMessage(*opponentID, models.ServerMessage{
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
		conn.SendMessage(userID, models.ServerMessage{
			Type:         "rematch_timeout",
			Message:      "Rematch request timed out",
			AllowRematch: &allowRematch,
		})
		
		conn.SendMessage(*opponentID, models.ServerMessage{
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
		conn.SendMessage(requesterID, models.ServerMessage{
			Type:    "rematch_accepted",
			Message: "Rematch accepted",
		})
		conn.SendMessage(userID, models.ServerMessage{
			Type:    "rematch_accepted",
			Message: "Rematch accepted",
		})

		// Create new game
		return gs.CreateRematchGame(conn, sessionManager)
	} else {
		log.Printf("[REMATCH] %s (ID: %d) declined rematch in game %s", responderUsername, userID, gs.GameID)
		
		allowRematch := false // Declined, can't rematch
		// Notify both players
		conn.SendMessage(requesterID, models.ServerMessage{
			Type:         "rematch_declined",
			Message:      "Rematch declined",
			AllowRematch: &allowRematch,
		})
		conn.SendMessage(userID, models.ServerMessage{
			Type:         "rematch_declined",
			Message:      "Rematch declined",
			AllowRematch: &allowRematch,
		})

		// Clean up connections
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

// CancelRematchRequest cancels a pending rematch request when requester leaves
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
		conn.SendMessage(*opponentID, models.ServerMessage{
			Type:    "rematch_cancelled",
			Message: "Rematch request cancelled",
		})
	}

	gs.RematchRequester = nil
}

// CreateRematchGame creates a new game session with the same players
func (gs *GameSession) CreateRematchGame(conn ConnectionManagerInterface, sessionManager *SessionManager) error {
	log.Printf("[REMATCH] Creating rematch for game %s", gs.GameID)

	// Stop all timers
	if gs.PostGameTimer != nil {
		gs.PostGameTimer.Stop()
		gs.PostGameTimer = nil
	}
	if gs.RematchRequestTimer != nil {
		gs.RematchRequestTimer.Stop()
		gs.RematchRequestTimer = nil
	}

	// Remove old session BEFORE creating new one
	// This prevents RemoveSession from deleting the new session's UserToGame mappings
	sessionManager.RemoveSession(gs.GameID)

	// Create new game with same players and same color assignment
	newSession := NewGameSession(
		gs.Player1ID,
		gs.Player1Username,
		gs.Player2ID,
		gs.Player2Username,
		gs.BotDifficulty,
		conn,
	)

	// Add new session to session manager manually
	sessionManager.Mux.Lock()
	sessionManager.Session[newSession.GameID] = newSession
	sessionManager.UserToGame[gs.Player1ID] = newSession.GameID
	if !newSession.IsBot() && gs.Player2ID != nil {
		sessionManager.UserToGame[*gs.Player2ID] = newSession.GameID
	}
	sessionManager.Mux.Unlock()

	log.Printf("[REMATCH] Created new game %s (rematch of %s)", newSession.GameID, gs.GameID)

	return nil
}