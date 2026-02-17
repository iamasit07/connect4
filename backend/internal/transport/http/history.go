package http

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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

func (h *HistoryHandler) GetHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		log.Printf("[HISTORY] Unauthorized: user_id is missing or zero")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	rawHistory, err := h.GameRepo.GetUserGameHistory(userID)
	if err != nil {
		log.Printf("[HISTORY] Error fetching history for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}

	response := make([]historyResponse, 0, len(rawHistory))
	for _, game := range rawHistory {
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

	c.JSON(http.StatusOK, response)
}

func (h *HistoryHandler) GetGameDetails(c *gin.Context) {
	gameID := c.Param("id")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	game, err := h.GameRepo.GetGameByID(gameID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	// Also fetch board state
	board, err := h.GameRepo.GetGameBoard(gameID)
	if err != nil {
		// If board fails, just return game info
		c.JSON(http.StatusOK, game)
		return
	}

	response := struct {
		*postgres.GameResult
		Board [][]int `json:"board_state"`
	}{
		GameResult: game,
		Board:      board,
	}

	c.JSON(http.StatusOK, response)
}
