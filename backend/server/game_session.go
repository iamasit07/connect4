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

// ConnectionManagerInterface defines the interface for sending messages
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
	mu                  sync.Mutex
}

// Helper methods
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

// saveGameAsync saves game data to database asynchronously
// This prevents blocking when sending game_over messages to players
// All parameters are passed by value to avoid race conditions
func (gs *GameSession) saveGameAsync(gameID string, p1ID int64, p1User string,
	p2ID *int64, p2User string, winnerID *int64, winnerUser string,
	reason string, moves, duration int, created, finished time.Time) {
	
	go func() {
		err := db.SaveGame(gameID, p1ID, p1User, p2ID, p2User,
			winnerID, winnerUser, reason, moves, duration, created, finished)
		if err != nil {
			log.Printf("[GAME] Error saving game %s: %v", gameID, err)
		} else {
			log.Printf("[GAME] Game %s saved successfully", gameID)
		}
	}()
}

// NewGameSession creates a  new game session
func NewGameSession(player1ID int64, player1Username string, player2ID *int64, player2Username string, conn ConnectionManagerInterface) *GameSession {
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
		CreatedAt:       time.Now(),
		mu:              sync.Mutex{},
	}

	// Send game_start to player1
	conn.SendMessage(player1ID, models.ServerMessage{
		Type:        "game_start",
		GameID:      gs.GameID,
		Opponent:    player2Username,
		YourPlayer:  int(models.Player1),
		CurrentTurn: int(gs.Game.CurrentPlayer),
	})

	// Send game_start to player2 if not bot
	if player2Username != models.BotUsername && player2ID != nil {
		conn.SendMessage(*player2ID, models.ServerMessage{
			Type:        "game_start",
			GameID:      gs.GameID,
			Opponent:    player1Username,
			YourPlayer:  int(models.Player2),
			CurrentTurn: int(gs.Game.CurrentPlayer),
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

	// Check for win
	if gs.Game.Status == models.StatusWon {
		gs.FinishedAt = time.Now()
		winnerUsername := gs.GetUsername(gs.Game.Winner)
		winnerID := userID
		gs.Reason = "connect_four"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		gameOverMsg := models.ServerMessage{
			Type:   "game_over",
			Winner: winnerUsername,
			Reason: gs.Reason,
			Board:  gs.Game.Board,
		}

		conn.SendMessage(gs.Player1ID, gameOverMsg)
		if !gs.IsBot() && gs.Player2ID != nil {
			conn.SendMessage(*gs.Player2ID, gameOverMsg)
		}

		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			gs.Player2ID, gs.Player2Username, &winnerID, winnerUsername,
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt)

		return nil
	}

	// Check for draw
	if gs.Game.Status == models.StatusDraw {
		gs.FinishedAt = time.Now()
		gs.Reason = "draw"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		gameOverMsg := models.ServerMessage{
			Type:   "game_over",
			Winner: "draw",
			Reason: "draw",
			Board:  gs.Game.Board,
		}

		conn.SendMessage(gs.Player1ID, gameOverMsg)
		if !gs.IsBot() && gs.Player2ID != nil {
			conn.SendMessage(*gs.Player2ID, gameOverMsg)
		}

		// Save game asynchronously (non-blocking)
		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			gs.Player2ID, gs.Player2Username, nil, "draw",
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt)

		return nil
	}

	// Send move_made to both players
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

	// Handle bot move if playing against bot
	if gs.IsBot() && gs.Game.CurrentPlayer == models.Player2 {
		return gs.HandleBotMove(conn)
	}

	return nil
}

func (gs *GameSession) HandleBotMove(conn ConnectionManagerInterface) error {
	botColumn := bot.CalculateBestMove(gs.Game.Board, models.Player2)
	botRow, err := gs.Game.MakeMove(models.Player2, botColumn)
	if err != nil {
		return err
	}

	// Check for bot win
	if gs.Game.Status == models.StatusWon {
		gs.FinishedAt = time.Now()
		gs.Reason = "connect_four"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		gameOverMsg := models.ServerMessage{
			Type:   "game_over",
			Winner: models.BotUsername,
			Reason: gs.Reason,
			Board:  gs.Game.Board,
		}
		conn.SendMessage(gs.Player1ID, gameOverMsg)
		// Don't close connection - let player see game over screen

		// Save game asynchronously (bot wins)
		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			nil, models.BotUsername, nil, models.BotUsername,
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt)
		return nil
	}

	// Check for draw
	if gs.Game.Status == models.StatusDraw {
		gs.FinishedAt = time.Now()
		gs.Reason = "draw"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		gameOverMsg := models.ServerMessage{
			Type:   "game_over",
			Winner: "draw",
			Reason: "draw",
			Board:  gs.Game.Board,
		}
		conn.SendMessage(gs.Player1ID, gameOverMsg)
		// Don't close connection - let player see game over screen

		// Save game asynchronously (draw)
		gs.saveGameAsync(gs.GameID, gs.Player1ID, gs.Player1Username,
			nil, models.BotUsername, nil, "draw",
			gs.Reason, gs.Game.MoveCount, duration, gs.CreatedAt, gs.FinishedAt)
		return nil
	}

	// Send bot move to player1
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

	if gs.Game.IsFinished() {
		return nil
	}

	username := gs.GetUsernameByUserID(userID)
	log.Printf("[DISCONNECT] Player %s (ID: %d) disconnected from game %s", username, userID, gs.GameID)

	// Add to disconnected list
	gs.DisconnectedPlayers = append(gs.DisconnectedPlayers, userID)

	// Notify opponent
	opponentID := gs.GetOpponentID(userID)
	if !gs.IsBot() && opponentID != nil {
		conn.SendMessage(*opponentID, models.ServerMessage{
			Type:    "opponent_disconnected",
			Message: fmt.Sprintf("%s disconnected", username),
		})
	}

	// Check if both disconnected
	if len(gs.DisconnectedPlayers) == 2 {
		log.Printf("[DISCONNECT] Both players disconnected from game %s", gs.GameID)
		gs.FinishedAt = time.Now()
		gs.Reason = "both_disconnected"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		// Save game as draw
		err := db.SaveGame(
			gs.GameID,
			gs.Player1ID, gs.Player1Username,
			gs.Player2ID, gs.Player2Username,
			nil, "draw",
			gs.Reason,
			gs.Game.MoveCount,
			duration,
			gs.CreatedAt,
			gs.FinishedAt,
		)
		if err != nil {
			log.Printf("[GAME] Error saving game: %v", err)
		}

		gs.cleanupConnections(conn)
		sessionManager.RemoveSession(gs.GameID)
		return nil
	}

	// Start reconnect timer
	gs.ReconnectTimer = time.AfterFunc(config.AppConfig.ReconnectTimeout, func() {
		gs.HandleReconnectTimeout(userID, conn, sessionManager)
	})

	return nil
}

func (gs *GameSession) HandleReconnect(userID int64, conn ConnectionManagerInterface) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	username := gs.GetUsernameByUserID(userID)
	log.Printf("[RECONNECT] Player %s (ID: %d) reconnecting to game %s", username, userID, gs.GameID)

	// Remove from disconnected list
	newDisconnected := []int64{}
	for _, uid := range gs.DisconnectedPlayers {
		if uid != userID {
			newDisconnected = append(newDisconnected, uid)
		}
	}
	gs.DisconnectedPlayers = newDisconnected

	// Stop timer
	if gs.ReconnectTimer != nil {
		gs.ReconnectTimer.Stop()
		gs.ReconnectTimer = nil
	}

	// Send reconnect message
	playerID, _ := gs.GetPlayerID(userID)
	conn.SendMessage(userID, models.ServerMessage{
		Type:        "reconnect_success",
		GameID:      gs.GameID,
		Opponent:    gs.GetOpponentUsername(userID),
		YourPlayer:  int(playerID),
		CurrentTurn: int(gs.Game.CurrentPlayer),
		Board:       gs.Game.Board,
	})

	// Notify opponent
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

	// Save game - opponent wins
	err := db.SaveGame(
		gs.GameID,
		gs.Player1ID, gs.Player1Username,
		gs.Player2ID, gs.Player2Username,
		opponentID, opponentUsername,
		gs.Reason,
		gs.Game.MoveCount,
		duration,
		gs.CreatedAt,
		gs.FinishedAt,
	)
	if err != nil {
		log.Printf("[GAME] Error saving game: %v", err)
	}

	// Send game_over to both
	gameOverMsg := models.ServerMessage{
		Type:   "game_over",
		Winner: opponentUsername,
		Reason: "timeout",
	}

	conn.SendMessage(userID, gameOverMsg)
	if !gs.IsBot() && opponentID != nil {
		conn.SendMessage(*opponentID, gameOverMsg)
	}

	// Clean up
	conn.RemoveConnection(userID)
	if !gs.IsBot() && opponentID != nil {
		conn.RemoveConnection(*opponentID)
	}

	sessionManager.RemoveSession(gs.GameID)
}

func (gs *GameSession) TerminateSessionByAbandonment(abandoningUserID int64, conn ConnectionManagerInterface) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	abandoningUsername := gs.GetUsernameByUserID(abandoningUserID)
	opponentUsername := gs.GetOpponentUsername(abandoningUserID)
	opponentID := gs.GetOpponentID(abandoningUserID)

	log.Printf("[TERMINATE] Game %s terminated by abandonment from %s (ID: %d)",
		gs.GameID, abandoningUsername, abandoningUserID)

	if gs.ReconnectTimer != nil {
		gs.ReconnectTimer.Stop()
		gs.ReconnectTimer = nil
		log.Printf("[TERMINATE] Canceled reconnect timer for game %s", gs.GameID)
	}

	gs.FinishedAt = time.Now()
	gs.Reason = "abandoned"

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
		gs.FinishedAt,
	)
	if err != nil {
		log.Printf("[GAME] Error saving game: %v", err)
	}

	gameOverMsg := models.ServerMessage{
		Type:   "game_over",
		Winner: opponentUsername,
		Reason: "abandoned",
	}

	if !gs.IsBot() && opponentID != nil {
		conn.SendMessage(*opponentID, gameOverMsg)
	}

	return nil
}
