package server

import (
	"fmt"
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
	SendMessage(username string, message models.ServerMessage) error
	RemoveConnection(username string)
}

type GameSession struct {
	GameID              string
	Player1Username     string
	Player2Username     string
	Game                *game.Game
	PlayerMapping       map[string]models.PlayerID // username -> PlayerID (1 or 2)
	Reason              string
	DisconnectedPlayers []string // usernames
	ReconnectTimer      *time.Timer
	CreatedAt           time.Time
	FinishedAt          time.Time
	UserTokens          map[string]string // username -> userToken for auth
	mu                  sync.Mutex
}

// Helper methods to convert between username and PlayerID
func (gs *GameSession) GetPlayerID(username string) (models.PlayerID, bool) {
	playerID, exists := gs.PlayerMapping[username]
	return playerID, exists
}

func (gs *GameSession) GetUsername(playerID models.PlayerID) string {
	if playerID == models.Player1 {
		return gs.Player1Username
	}
	return gs.Player2Username
}

func (gs *GameSession) GetOpponentUsername(username string) string {
	if username == gs.Player1Username {
		return gs.Player2Username
	}
	return gs.Player1Username
}

func (gs *GameSession) IsBot() bool {
	return gs.Player2Username == models.BotUsername
}

func NewGameSession(player1Username, player2Username string, userTokens map[string]string, conn ConnectionManagerInterface) *GameSession {
	gameID := utils.GenerateGameID()
	newGame := (&game.Game{}).NewGame()

	// Create username -> PlayerID mapping
	mapping := make(map[string]models.PlayerID)
	mapping[player1Username] = models.Player1
	mapping[player2Username] = models.Player2

	gs := &GameSession{
		GameID:          gameID,
		Player1Username: player1Username,
		Player2Username: player2Username,
		Game:            newGame,
		PlayerMapping:   mapping,
		UserTokens:      userTokens, // Use userTokens for authentication
		CreatedAt:       time.Now(),
		mu:              sync.Mutex{},
	}

	// Send game start message to player1
	player1Message := models.ServerMessage{
		Type:        "game_start",
		GameID:      gs.GameID,
		Opponent:    player2Username,
		YourPlayer:  int(models.Player1),
		CurrentTurn: int(gs.Game.CurrentPlayer),
	}
	conn.SendMessage(player1Username, player1Message)

	// Send game start message to player2 (if not bot)
	if player2Username != models.BotUsername {
		player2Message := models.ServerMessage{
			Type:        "game_start",
			GameID:      gs.GameID,
			Opponent:    player1Username,
			YourPlayer:  int(models.Player2),
			CurrentTurn: int(gs.Game.CurrentPlayer),
		}
		conn.SendMessage(player2Username, player2Message)
	}

	return gs
}

func (gs *GameSession) HandleMove(username string, column int, conn ConnectionManagerInterface) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	playerID, exists := gs.GetPlayerID(username)
	if !exists {
		return fmt.Errorf("player not in this game")
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
		gs.Reason = "connect_four"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())
		err := db.SaveGame(
			gs.GameID,
			gs.Player1Username,
			gs.Player2Username,
			winnerUsername,
			gs.Reason,
			gs.Game.MoveCount,
			duration,
			gs.CreatedAt,
			gs.FinishedAt,
		)
		if err != nil {
			return err
		}

		game_over_message := models.ServerMessage{
			Type:   "game_over",
			Winner: winnerUsername,
			Reason: gs.Reason,
			Board:  gs.Game.Board,
		}

		conn.SendMessage(gs.Player1Username, game_over_message)
		if !gs.IsBot() {
			conn.SendMessage(gs.Player2Username, game_over_message)
		}

		// Clean up connections
		conn.RemoveConnection(gs.Player1Username)
		if !gs.IsBot() {
			conn.RemoveConnection(gs.Player2Username)
		}
		return nil
	}

	// Check for draw
	if gs.Game.Status == models.StatusDraw {
		gs.FinishedAt = time.Now()
		game_over_message := models.ServerMessage{
			Type:   "game_over",
			Winner: "draw",
			Reason: "draw",
			Board:  gs.Game.Board,
		}
		conn.SendMessage(gs.Player1Username, game_over_message)
		if !gs.IsBot() {
			conn.SendMessage(gs.Player2Username, game_over_message)
		}

		// Clean up connections
		conn.RemoveConnection(gs.Player1Username)
		if !gs.IsBot() {
			conn.RemoveConnection(gs.Player2Username)
		}
		return nil
	}

	// Send move_made to both players
	move_made_message := models.ServerMessage{
		Type:     "move_made",
		Column:   column,
		Row:      row,
		Player:   int(playerID),
		Board:    gs.Game.Board,
		NextTurn: int(gs.Game.CurrentPlayer),
	}
	conn.SendMessage(gs.Player1Username, move_made_message)
	if !gs.IsBot() {
		conn.SendMessage(gs.Player2Username, move_made_message)
	}

	// Handle bot move if playing against bot
	if gs.IsBot() && gs.Game.CurrentPlayer == models.Player2 {
		botColumn := bot.CalculateBestMove(gs.Game.Board, models.Player2)
		botRow, err := gs.Game.MakeMove(models.Player2, botColumn)
		if err != nil {
			return err
		}

		if gs.Game.Status == models.StatusWon {
			gs.FinishedAt = time.Now()
			game_over_message := models.ServerMessage{
				Type:   "game_over",
				Winner: models.BotUsername,
				Reason: "connect_four",
				Board:  gs.Game.Board,
			}
			conn.SendMessage(gs.Player1Username, game_over_message)
			// Clean up connection (bot doesn't have connection)
			conn.RemoveConnection(gs.Player1Username)
			return nil
		}

		if gs.Game.Status == models.StatusDraw {
			gs.FinishedAt = time.Now()
			game_over_message := models.ServerMessage{
				Type:   "game_over",
				Winner: "draw",
				Reason: "draw",
				Board:  gs.Game.Board,
			}
			conn.SendMessage(gs.Player1Username, game_over_message)
			// Clean up connection (bot doesn't have connection)
			conn.RemoveConnection(gs.Player1Username)
			return nil
		}

		bot_move_made_message := models.ServerMessage{
			Type:     "move_made",
			Column:   botColumn,
			Row:      botRow,
			Player:   int(models.Player2),
			Board:    gs.Game.Board,
			NextTurn: int(gs.Game.CurrentPlayer),
		}
		conn.SendMessage(gs.Player1Username, bot_move_made_message)
	}

	return nil
}

func (gs *GameSession) HandleDisconnect(username string, conn ConnectionManagerInterface, sessionManager *SessionManager) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.Game.IsFinished() {
		return nil
	}

	gs.DisconnectedPlayers = append(gs.DisconnectedPlayers, username)

	// Check if both players are now disconnected
	if len(gs.DisconnectedPlayers) >= 2 {
		// Both players disconnected - end game as draw immediately
		gs.FinishedAt = time.Now()
		gs.Reason = "both_disconnected"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())
		err := db.SaveGame(
			gs.GameID,
			gs.Player1Username,
			gs.Player2Username,
			"draw",
			gs.Reason,
			gs.Game.MoveCount,
			duration,
			gs.CreatedAt,
			gs.FinishedAt,
		)
		if err != nil {
			return err
		}

		game_over_message := models.ServerMessage{
			Type:   "game_over",
			Winner: "draw",
			Reason: "both_disconnected",
			Board:  gs.Game.Board,
		}

		conn.SendMessage(gs.Player1Username, game_over_message)
		if !gs.IsBot() {
			conn.SendMessage(gs.Player2Username, game_over_message)
		}

		// Clean up connections
		conn.RemoveConnection(gs.Player1Username)
		if !gs.IsBot() {
			conn.RemoveConnection(gs.Player2Username)
		}

		return nil
	}

	// Only one player disconnected - start reconnect timer
	disconnect_message := models.ServerMessage{
		Type:    "opponent_disconnected",
		Message: "Your opponent has disconnected. Waiting for reconnection...",
	}

	gs.ReconnectTimer = time.AfterFunc(config.AppConfig.ReconnectTimeout, func() {
		gs.HandleReconnectTimeout(username, conn, sessionManager)
	})

	opponentUsername := gs.GetOpponentUsername(username)
	conn.SendMessage(opponentUsername, disconnect_message)
	return nil
}

func (gs *GameSession) HandleReconnectTimeout(username string, conn ConnectionManagerInterface, sessionManager *SessionManager) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Check if player is actually disconnected
	isDisconnected := false
	for _, disconnectedUser := range gs.DisconnectedPlayers {
		if disconnectedUser == username {
			isDisconnected = true
			break
		}
	}

	if !isDisconnected {
		return fmt.Errorf("player not disconnected")
	}

	// Check if game is already finished (race condition prevention)
	if gs.Game.IsFinished() {
		return nil
	}

	gs.FinishedAt = time.Now()
	gs.Reason = "opponent_timeout"

	if gs.ReconnectTimer != nil {
		gs.ReconnectTimer.Stop()
		gs.ReconnectTimer = nil
	}

	// Determine winner (opponent of disconnected player)
	opponentUsername := gs.GetOpponentUsername(username)
	opponentPlayerID, _ := gs.GetPlayerID(opponentUsername)

	// Mark game as finished
	gs.Game.Status = models.StatusWon
	gs.Game.Winner = opponentPlayerID

	// Save game to database
	duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())
	err := db.SaveGame(
		gs.GameID,
		gs.Player1Username,
		gs.Player2Username,
		opponentUsername,
		gs.Reason,
		gs.Game.MoveCount,
		duration,
		gs.CreatedAt,
		gs.FinishedAt,
	)
	if err != nil {
		return err
	}

	// Send game over message
	game_over_message := models.ServerMessage{
		Type:   "game_over",
		GameID: gs.GameID,
		Winner: opponentUsername,
		Board:  gs.Game.Board,
		Reason: gs.Reason,
	}

	conn.SendMessage(gs.Player1Username, game_over_message)
	if !gs.IsBot() {
		conn.SendMessage(gs.Player2Username, game_over_message)
	}

	// Clean up connections
	conn.RemoveConnection(gs.Player1Username)
	if !gs.IsBot() {
		conn.RemoveConnection(gs.Player2Username)
	}

	// Clean up session from SessionManager
	sessionManager.RemoveSession(
		gs.GameID,
		gs.Player1Username,
		gs.Player2Username,
	)

	return nil
}

func (gs *GameSession) HandleReconnect(username string, conn ConnectionManagerInterface) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	isDisconnected := false
	disconnectedIndex := -1
	for i, disconnectedUser := range gs.DisconnectedPlayers {
		if disconnectedUser == username {
			isDisconnected = true
			disconnectedIndex = i
			break
		}
	}

	if !isDisconnected {
		return fmt.Errorf("player was not disconnected")
	}

	gs.DisconnectedPlayers = append(gs.DisconnectedPlayers[:disconnectedIndex], gs.DisconnectedPlayers[disconnectedIndex+1:]...)

	playerID, _ := gs.GetPlayerID(username)
	opponentUsername := gs.GetOpponentUsername(username)

	reconnect_message := models.ServerMessage{
		Type:        "game_start",
		GameID:      gs.GameID,
		Opponent:    opponentUsername,
		YourPlayer:  int(playerID),
		CurrentTurn: int(gs.Game.CurrentPlayer),
		Board:       gs.Game.Board,
	}

	opponent_reconnect_message := models.ServerMessage{
		Type:    "opponent_reconnected",
		Message: "Your opponent has reconnected. Resuming the game.",
	}

	if gs.ReconnectTimer != nil {
		gs.ReconnectTimer.Stop()
		gs.ReconnectTimer = nil
	}

	conn.SendMessage(username, reconnect_message)
	conn.SendMessage(opponentUsername, opponent_reconnect_message)
	return nil
}

func (gs *GameSession) TerminateSession(forfeitingPlayer string, reason string, conn ConnectionManagerInterface) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.Game.IsFinished() {
		return nil // Already finished
	}

	gs.FinishedAt = time.Now()
	gs.Reason = reason

	// Stop any active reconnect timer
	if gs.ReconnectTimer != nil {
		gs.ReconnectTimer.Stop()
		gs.ReconnectTimer = nil
	}

	// Determine winner (opponent of forfeiting player)
	opponentUsername := gs.GetOpponentUsername(forfeitingPlayer)
	winner := opponentUsername

	// Special case: if opponent is BOT and player abandons, it's still a forfeit
	if opponentUsername == models.BotUsername {
		winner = models.BotUsername
	}

	duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())
	err := db.SaveGame(
		gs.GameID,
		gs.Player1Username,
		gs.Player2Username,
		winner,
		gs.Reason,
		gs.Game.MoveCount,
		duration,
		gs.CreatedAt,
		gs.FinishedAt,
	)
	if err != nil {
		return err
	}

	game_over_message := models.ServerMessage{
		Type:   "game_over",
		GameID: gs.GameID,
		Winner: winner,
		Board:  gs.Game.Board,
		Reason: reason,
	}

	conn.SendMessage(gs.Player1Username, game_over_message)
	if !gs.IsBot() {
		conn.SendMessage(gs.Player2Username, game_over_message)
	}

	// Clean up connections
	conn.RemoveConnection(gs.Player1Username)
	if !gs.IsBot() {
		conn.RemoveConnection(gs.Player2Username)
	}

	return nil
}
