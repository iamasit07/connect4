package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/internal/repository/postgres"
)

type HistoryHandler struct {
	GameRepo *postgres.GameRepo
}

func NewHistoryHandler(gameRepo *postgres.GameRepo) *HistoryHandler {
	return &HistoryHandler{GameRepo: gameRepo}
}

func (h *HistoryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rawHistory, err := h.GameRepo.GetUserGameHistory(userID)
	if err != nil {
		http.Error(w, "Failed to fetch history", http.StatusInternalServerError)
		return
	}

	// Map to frontend expectation
	type GameHistoryItem struct {
		ID               string    `json:"id"`
		OpponentUsername string    `json:"opponentUsername"`
		Result           string    `json:"result"` // "win", "loss", "draw"
		EndReason        string    `json:"endReason"`
		CreatedAt        time.Time `json:"createdAt"`
		MovesCount       int       `json:"movesCount"`
	}

	history := make([]GameHistoryItem, 0, len(rawHistory))
	for _, game := range rawHistory {
		item := GameHistoryItem{
			ID:         game.GameID,
			EndReason:  game.Reason,
			CreatedAt:  game.CreatedAt,
			MovesCount: game.TotalMoves,
		}

		// Determine opponent
		if game.Player1ID == userID {
			if game.Player2Username != "" {
				item.OpponentUsername = game.Player2Username
			} else {
				item.OpponentUsername = "Waiting..." // Should shouldn't happen for finished games
			}
		} else {
			item.OpponentUsername = game.Player1Username
		}

		// Determine result
		if game.WinnerID != nil {
			if *game.WinnerID == userID {
				item.Result = "win"
			} else {
				item.Result = "loss"
			}
		} else {
			item.Result = "draw"
		}

		history = append(history, item)
	}

	json.NewEncoder(w).Encode(history)
}

func (h *HistoryHandler) GetGameDetails(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path (assuming /api/history/{id})
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}
	// Depending on route definition, ID might be last or 2nd to last
	// If path is /api/history/{id}, it's last.
	// We'll trust the user's logic or adjust.
	gameID := pathParts[len(pathParts)-1]
    if gameID == "" {
        gameID = pathParts[len(pathParts)-2]
    }

	game, err := h.GameRepo.GetGameByID(gameID)
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}
	
	// Also fetch board state
	board, err := h.GameRepo.GetGameBoard(gameID)
	if err != nil {
		// If board fails, just return game info
		json.NewEncoder(w).Encode(game)
		return
	}

	response := struct {
		*postgres.GameResult
		Board [][]int `json:"board_state"`
	}{
		GameResult: game,
		Board:      board,
	}

	json.NewEncoder(w).Encode(response)
}
