package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/game"
)

type WatchHandler struct {
	SessionManager *game.SessionManager
}

func NewWatchHandler(sm *game.SessionManager) *WatchHandler {
	return &WatchHandler{SessionManager: sm}
}

type liveGameResponse struct {
	GameID         string         `json:"gameId"`
	Player1        playerResponse `json:"player1"`
	Player2        playerResponse `json:"player2"`
	SpectatorCount int            `json:"spectatorCount"`
	MoveCount      int            `json:"moveCount"`
	StartedAt      string         `json:"startedAt"`
}

type playerResponse struct {
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}

// GetLiveGames returns all active PvP games available for spectating
func (h *WatchHandler) GetLiveGames(c *gin.Context) {
	activeGames := h.SessionManager.GetActiveGames()

	response := make([]liveGameResponse, 0, len(activeGames))
	for _, g := range activeGames {
		response = append(response, liveGameResponse{
			GameID:         g.GameID,
			Player1:        playerResponse{Username: g.Player1, Rating: 0},
			Player2:        playerResponse{Username: g.Player2, Rating: 0},
			SpectatorCount: g.SpectatorCount,
			MoveCount:      g.MoveCount,
			StartedAt:      g.StartedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}
