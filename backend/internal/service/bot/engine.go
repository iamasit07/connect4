package bot

import (
	"github.com/iamasit07/connect4/backend/internal/domain"
)

// CalculateBestMove selects the best move based on difficulty
func CalculateBestMove(board [][]domain.PlayerID, botPlayer domain.PlayerID, difficulty string) int {
	switch difficulty {
	case "easy":
		return CalculateBestMoveEasy(board, botPlayer)
	case "medium":
		return calculateMediumMove(board, botPlayer)
	case "hard":
		return CalculateBestMoveMinimax(board, botPlayer)
	default:
		return calculateMediumMove(board, botPlayer)
	}
}

func getOpponent(p domain.PlayerID) domain.PlayerID {
	if p == domain.Player1 {
		return domain.Player2
	}
	return domain.Player1
}