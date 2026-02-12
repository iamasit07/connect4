package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/iamasit07/4-in-a-row/backend/internal/repository/postgres"
)

type HistoryHandler struct {
	GameRepo *postgres.GameRepo
}

func NewHistoryHandler(gameRepo *postgres.GameRepo) *HistoryHandler {
	return &HistoryHandler{GameRepo: gameRepo}
}

type historyResponse struct {
	ID               string `json:"id"`
	OpponentUsername string `json:"opponentUsername"`
	Result           string `json:"result"`
	EndReason        string `json:"endReason"`
	CreatedAt        string `json:"createdAt"`
	MovesCount       int    `json:"movesCount"`
}

func (h *HistoryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	history, err := h.GameRepo.GetUserGameHistory(userID)
	if err != nil {
		log.Printf("[HISTORY] Error fetching history for user %d: %v", userID, err)
		http.Error(w, "Failed to fetch history", http.StatusInternalServerError)
		return
	}

	log.Printf("[HISTORY] Found %d games for user %d", len(history), userID)

	response := make([]historyResponse, 0, len(history))
	for _, game := range history {
		opponent := game.Player2Username
		if game.Player1ID != userID {
			opponent = game.Player1Username
		}
		if opponent == "" {
			opponent = "BOT"
		}

		var result string
		if game.Reason == "draw" {
			result = "draw"
		} else if game.WinnerID != nil && *game.WinnerID == userID {
			result = "win"
		} else {
			result = "loss"
		}

		response = append(response, historyResponse{
			ID:               game.GameID,
			OpponentUsername: opponent,
			Result:           result,
			EndReason:        game.Reason,
			CreatedAt:        game.CreatedAt.Format("2006-01-02T15:04:05Z"),
			MovesCount:       game.TotalMoves,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *HistoryHandler) GetGameDetails(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path (assuming /api/history/{id})
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}
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
