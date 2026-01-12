package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/db"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

type GameHistoryItem struct {
	GameID          string    `json:"game_id"`
	Player1         PlayerInfo `json:"player1"`
	Player2         PlayerInfo `json:"player2"`
	Result          string     `json:"result"` // "won", "lost", "draw"
	Reason          string     `json:"reason"`
	TotalMoves      int        `json:"total_moves"`
	DurationSeconds int        `json:"duration_seconds"`
	CreatedAt       time.Time  `json:"created_at"`
	FinishedAt      time.Time  `json:"finished_at"`
}

type PlayerInfo struct {
	ID       *int64 `json:"id,omitempty"`
	Username string `json:"username"`
}

type GameHistoryResponse struct {
	Games []GameHistoryItem `json:"games"`
}

// HandleGameHistory returns the game history for the authenticated user
func HandleGameHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get token from cookie
	token, err := utils.GetTokenFromCookie(r)
	if err != nil {
		log.Printf("[GAME HISTORY] Error getting token from cookie: %v", err)
		log.Printf("[GAME HISTORY] Cookies: %v", r.Cookies())
		writeJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Validate JWT
	claims, err := utils.ValidateJWT(token)
	if err != nil {
		log.Printf("[GAME HISTORY] Error validating JWT: %v", err)
		writeJSONError(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	log.Printf("[GAME HISTORY] Fetching history for user %d (%s)", claims.UserID, claims.Username)

	// Fetch user's game history from database
	games, err := db.GetUserGameHistory(claims.UserID)
	if err != nil {
		writeJSONError(w, "Failed to fetch game history", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	historyItems := make([]GameHistoryItem, 0, len(games))
	for _, game := range games {
		// Determine result from the perspective of the requesting user
		var result string
		if game.WinnerUsername == "draw" {
			result = "draw"
		} else if game.WinnerID != nil && *game.WinnerID == claims.UserID {
			result = "won"
		} else {
			result = "lost"
		}

		item := GameHistoryItem{
			GameID: game.GameID,
			Player1: PlayerInfo{
				ID:       &game.Player1ID,
				Username: game.Player1Username,
			},
			Player2: PlayerInfo{
				ID:       game.Player2ID,
				Username: game.Player2Username,
			},
			Result:          result,
			Reason:          game.Reason,
			TotalMoves:      game.TotalMoves,
			DurationSeconds: game.DurationSeconds,
			CreatedAt:       game.CreatedAt,
			FinishedAt:      game.FinishedAt,
		}
		historyItems = append(historyItems, item)
	}

	writeJSON(w, GameHistoryResponse{Games: historyItems}, http.StatusOK)
}

type BoardViewResponse struct {
	GameID          string     `json:"game_id"`
	Board           [][]int    `json:"board"`
	Player1Username string     `json:"player1_username"`
	Player2Username string     `json:"player2_username"`
	Winner          string     `json:"winner"`
	Result          string     `json:"result"`
	TotalMoves      int        `json:"total_moves"`
	IsFinished      bool       `json:"is_finished"`
	CreatedAt       time.Time  `json:"created_at"`
	FinishedAt      time.Time  `json:"finished_at"`
}

// HandleGetBoard returns the board state for a finished game (read-only)
func HandleGetBoard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract gameID from path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		writeJSONError(w, "Invalid request path", http.StatusBadRequest)
		return
	}
	gameID := parts[len(parts)-2] // /api/game/:gameID/board

	// Fetch game from database
	game, err := db.GetGameByID(gameID)
	if err != nil {
		writeJSONError(w, "Failed to fetch game", http.StatusInternalServerError)
		return
	}
	if game == nil {
		writeJSONError(w, "Game not found", http.StatusNotFound)
		return
	}

	// Fetch board state
	board, err := db.GetGameBoard(gameID)
	if err != nil {
		writeJSONError(w, "Failed to fetch board state", http.StatusInternalServerError)
		return
	}

	response := BoardViewResponse{
		GameID:          game.GameID,
		Board:           board,
		Player1Username: game.Player1Username,
		Player2Username: game.Player2Username,
		Winner:          game.WinnerUsername,
		Result:          game.Reason,
		TotalMoves:      game.TotalMoves,
		IsFinished:      true,
		CreatedAt:       game.CreatedAt,
		FinishedAt:      game.FinishedAt,
	}

	writeJSON(w, response, http.StatusOK)
}
