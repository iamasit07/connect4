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
	SendMessage(userToken string, message models.ServerMessage) error
	RemoveConnection(userToken string)
}

type GameSession struct {
	GameID              string
	Player1Token        string                      // Token-based identification
	Player1Username     string                      // For display
	Player2Token        string                      // Token-based identification
	Player2Username     string                      // For display
	Game                *game.Game
	PlayerMapping       map[string]models.PlayerID  // userToken → PlayerID
	Reason              string
	DisconnectedPlayers []string                    // userTokens
	ReconnectTimer      *time.Timer
	CreatedAt           time.Time
	FinishedAt          time.Time
	mu                  sync.Mutex
}

// Helper methods
func (gs *GameSession) GetPlayerID(userToken string) (models.PlayerID, bool) {
	playerID, exists := gs.PlayerMapping[userToken]
	return playerID, exists
}

func (gs *GameSession) GetUsername(playerID models.PlayerID) string {
	if playerID == models.Player1 {
		return gs.Player1Username
	}
	return gs.Player2Username
}

func (gs *GameSession) GetUsernameByToken(userToken string) string {
	if userToken == gs.Player1Token {
		return gs.Player1Username
	}
	return gs.Player2Username
}

func (gs *GameSession) GetOpponentUsername(userToken string) string {
	if userToken == gs.Player1Token {
		return gs.Player2Username
	}
	return gs.Player1Username
}

func (gs *GameSession) GetOpponentToken(userToken string) string {
	if userToken == gs.Player1Token {
		return gs.Player2Token
	}
	return gs.Player1Token
}

func (gs *GameSession) IsBot() bool {
	return gs.Player2Username == models.BotUsername
}

func NewGameSession(player1Token, player1Username, player2Token, player2Username string, conn ConnectionManagerInterface) *GameSession {
	gameID := utils.GenerateGameID()
	newGame := (&game.Game{}).NewGame()

	// Create token → PlayerID mapping
	mapping := make(map[string]models.PlayerID)
	mapping[player1Token] = models.Player1
	mapping[player2Token] = models.Player2

	gs := &GameSession{
		GameID:          gameID,
		Player1Token:    player1Token,
		Player1Username: player1Username,
		Player2Token:    player2Token,
		Player2Username: player2Username,
		Game:            newGame,
		PlayerMapping:   mapping,
		CreatedAt:       time.Now(),
		mu:              sync.Mutex{},
	}

	// Send game start to player1
	conn.SendMessage(player1Token, models.ServerMessage{
		Type:        "game_start",
		GameID:      gs.GameID,
		Opponent:    player2Username,
		YourPlayer:  int(models.Player1),
		YourToken:   player1Token,
		CurrentTurn: int(gs.Game.CurrentPlayer),
	})

	// Send game start to player2 (if not bot)
	if player2Username != models.BotUsername {
		conn.SendMessage(player2Token, models.ServerMessage{
			Type:        "game_start",
			GameID:      gs.GameID,
			Opponent:    player1Username,
			YourPlayer:  int(models.Player2),
			YourToken:   player2Token,
			CurrentTurn: int(gs.Game.CurrentPlayer),
		})
	}

	return gs
}

func (gs *GameSession) HandleMove(userToken string, column int, conn ConnectionManagerInterface) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	playerID, exists := gs.GetPlayerID(userToken)
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
		winnerToken := userToken
		gs.Reason = "connect_four"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		// Save game
		err := db.SaveGame(
			gs.GameID,
			gs.Player1Token, gs.Player1Username,
			gs.Player2Token, gs.Player2Username,
			winnerToken, winnerUsername,
			gs.Reason,
			gs.Game.MoveCount,
			duration,
			gs.CreatedAt,
			gs.FinishedAt,
		)
		if err != nil {
			log.Printf("[GAME] Error saving game: %v", err)
		}

		// Send game_over to both players
		gameOverMsg := models.ServerMessage{
			Type:   "game_over",
			Winner: winnerUsername,
			Reason: gs.Reason,
			Board:  gs.Game.Board,
		}

		conn.SendMessage(gs.Player1Token, gameOverMsg)
		if !gs.IsBot() {
			conn.SendMessage(gs.Player2Token, gameOverMsg)
		}

		// Clean up connections
		conn.RemoveConnection(gs.Player1Token)
		if !gs.IsBot() {
			conn.RemoveConnection(gs.Player2Token)
		}

		return nil
	}

	// Check for draw
	if gs.Game.Status == models.StatusDraw {
		gs.FinishedAt = time.Now()
		gs.Reason = "draw"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		// Save game
		err := db.SaveGame(
			gs.GameID,
			gs.Player1Token, gs.Player1Username,
			gs.Player2Token, gs.Player2Username,
			"", "draw",
			gs.Reason,
			gs.Game.MoveCount,
			duration,
			gs.CreatedAt,
			gs.FinishedAt,
		)
		if err != nil {
			log.Printf("[GAME] Error saving game: %v", err)
		}

		// Send game_over to both players
		gameOverMsg := models.ServerMessage{
			Type:   "game_over",
			Winner: "draw",
			Reason: "draw",
			Board:  gs.Game.Board,
		}

		conn.SendMessage(gs.Player1Token, gameOverMsg)
		if !gs.IsBot() {
			conn.SendMessage(gs.Player2Token, gameOverMsg)
		}

		// Clean up connections
		conn.RemoveConnection(gs.Player1Token)
		if !gs.IsBot() {
			conn.RemoveConnection(gs.Player2Token)
		}

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

	conn.SendMessage(gs.Player1Token, moveMadeMsg)
	if !gs.IsBot() {
		conn.SendMessage(gs.Player2Token, moveMadeMsg)
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

		// Save game
		err := db.SaveGame(
			gs.GameID,
			gs.Player1Token, gs.Player1Username,
			gs.Player2Token, gs.Player2Username,
			gs.Player2Token, models.BotUsername,
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
			Winner: models.BotUsername,
			Reason: gs.Reason,
			Board:  gs.Game.Board,
		}
		conn.SendMessage(gs.Player1Token, gameOverMsg)
		conn.RemoveConnection(gs.Player1Token)
		return nil
	}

	// Check for draw
	if gs.Game.Status == models.StatusDraw {
		gs.FinishedAt = time.Now()
		gs.Reason = "draw"

		duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

		// Save game
		err := db.SaveGame(
			gs.GameID,
			gs.Player1Token, gs.Player1Username,
			gs.Player2Token, gs.Player2Username,
			"", "draw",
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
			Winner: "draw",
			Reason: "draw",
			Board:  gs.Game.Board,
		}
		conn.SendMessage(gs.Player1Token, gameOverMsg)
		conn.RemoveConnection(gs.Player1Token)
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
	conn.SendMessage(gs.Player1Token, botMoveMsg)

	return nil
}

func (gs *GameSession) HandleDisconnect(userToken string, conn ConnectionManagerInterface, sessionManager *SessionManager) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if gs.Game.IsFinished() {
		return nil
	}

	username := gs.GetUsernameByToken(userToken)
	log.Printf("[DISCONNECT] Player %s (%s) disconnected from game %s", username, userToken, gs.GameID)

	// Add to disconnected list
	gs.DisconnectedPlayers = append(gs.DisconnectedPlayers, userToken)

	// Notify opponent
	opponentToken := gs.GetOpponentToken(userToken)
	if !gs.IsBot() {
		conn.SendMessage(opponentToken, models.ServerMessage{
			Type:    "player_disconnected",
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
			gs.Player1Token, gs.Player1Username,
			gs.Player2Token, gs.Player2Username,
			"", "draw",
			gs.Reason,
			gs.Game.MoveCount,
			duration,
			gs.CreatedAt,
			gs.FinishedAt,
		)
		if err != nil {
			log.Printf("[GAME] Error saving game: %v", err)
		}

		// Remove connections
		conn.RemoveConnection(gs.Player1Token)
		conn.RemoveConnection(gs.Player2Token)

		// Remove session
		sessionManager.RemoveSession(gs.GameID, gs.Player1Token, gs.Player2Token)

		return nil
	}

	// Start reconnect timer
	gs.ReconnectTimer = time.AfterFunc(config.AppConfig.ReconnectTimeout, func() {
		gs.HandleReconnectTimeout(userToken, conn, sessionManager)
	})

	return nil
}

func (gs *GameSession) HandleReconnect(userToken string, conn ConnectionManagerInterface) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	username := gs.GetUsernameByToken(userToken)
	log.Printf("[RECONNECT] Player %s (%s) reconnecting to game %s", username, userToken, gs.GameID)

	// Remove from disconnected list
	newDisconnected := []string{}
	for _, token := range gs.DisconnectedPlayers {
		if token != userToken {
			newDisconnected = append(newDisconnected, token)
		}
	}
	gs.DisconnectedPlayers = newDisconnected

	// Stop timer
	if gs.ReconnectTimer != nil {
		gs.ReconnectTimer.Stop()
		gs.ReconnectTimer = nil
	}

	// Send reconnect message
	playerID, _ := gs.GetPlayerID(userToken)
	conn.SendMessage(userToken, models.ServerMessage{
		Type:        "reconnect_success",
		GameID:      gs.GameID,
		Opponent:    gs.GetOpponentUsername(userToken),
		YourPlayer:  int(playerID),
		YourToken:   userToken,
		CurrentTurn: int(gs.Game.CurrentPlayer),
		Board:       gs.Game.Board,
	})

	// Notify opponent
	opponentToken := gs.GetOpponentToken(userToken)
	if !gs.IsBot() {
		conn.SendMessage(opponentToken, models.ServerMessage{
			Type:    "player_reconnected",
			Message: fmt.Sprintf("%s reconnected", username),
		})
	}

	return nil
}

func (gs *GameSession) HandleReconnectTimeout(userToken string, conn ConnectionManagerInterface, sessionManager *SessionManager) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	username := gs.GetUsernameByToken(userToken)
	opponentToken := gs.GetOpponentToken(userToken)
	opponentUsername := gs.GetOpponentUsername(userToken)

	log.Printf("[TIMEOUT] Player %s (%s) reconnect timeout in game %s", username, userToken, gs.GameID)

	gs.FinishedAt = time.Now()
	gs.Reason = "timeout"

	duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

	// Save game - opponent wins
	err := db.SaveGame(
		gs.GameID,
		gs.Player1Token, gs.Player1Username,
		gs.Player2Token, gs.Player2Username,
		opponentToken, opponentUsername,
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

	conn.SendMessage(userToken, gameOverMsg)
	if !gs.IsBot() {
		conn.SendMessage(opponentToken, gameOverMsg)
	}

	// Clean up
	conn.RemoveConnection(userToken)
	if !gs.IsBot() {
		conn.RemoveConnection(opponentToken)
	}

	sessionManager.RemoveSession(gs.GameID, gs.Player1Token, gs.Player2Token)
}

func (gs *GameSession) TerminateSession(userToken, reason string, conn ConnectionManagerInterface) error {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	username := gs.GetUsernameByToken(userToken)
	opponentToken := gs.GetOpponentToken(userToken)
	opponentUsername := gs.GetOpponentUsername(userToken)

	log.Printf("[TERMINATE] Game %s terminated by %s (%s), reason: %s", gs.GameID, username, userToken, reason)

	gs.FinishedAt = time.Now()
	gs.Reason = reason

	duration := int(gs.FinishedAt.Sub(gs.CreatedAt).Seconds())

	// Determine winner based on reason
	var winnerToken, winnerUsername string
	if reason == "abandoned" {
		// Player who abandoned loses, opponent wins
		winnerToken = opponentToken
		winnerUsername = opponentUsername
	} else {
		// Other reasons treated as draw
		winnerToken = ""
		winnerUsername = "draw"
	}

	// Save game
	err := db.SaveGame(
		gs.GameID,
		gs.Player1Token, gs.Player1Username,
		gs.Player2Token, gs.Player2Username,
		winnerToken, winnerUsername,
		gs.Reason,
		gs.Game.MoveCount,
		duration,
		gs.CreatedAt,
		gs.FinishedAt,
	)
	if err != nil {
		log.Printf("[GAME] Error saving game: %v", err)
	}

	// Send game_over messages
	gameOverMsg := models.ServerMessage{
		Type:   "game_over",
		Winner: winnerUsername,
		Reason: reason,
	}

	conn.SendMessage(userToken, gameOverMsg)
	if !gs.IsBot() {
		conn.SendMessage(opponentToken, gameOverMsg)
	}

	// Clean up connections
	conn.RemoveConnection(userToken)
	if !gs.IsBot() {
		conn.RemoveConnection(opponentToken)
	}

	return nil
}
