package bot

import (
	"math/rand"

	"github.com/iamasit07/4-in-a-row/backend/internal/domain"
)
func CalculateBestMoveEasy(board [][]domain.PlayerID, botPlayer domain.PlayerID) int {
	validColumns := domain.GetValidMoves(board)
	if len(validColumns) == 0 {
		return -1
	}

	opponent := getOpponent(botPlayer)

	for _, col := range validColumns {
		testBoard, row, _ := domain.SimulateMove(board, col, botPlayer)
		if domain.CheckWin(testBoard, row, col, botPlayer) {
			return col
		}
	}

	for _, col := range validColumns {
		testBoard, row, _ := domain.SimulateMove(board, col, opponent)
		if domain.CheckWin(testBoard, row, col, opponent) {
			return col
		}
	}

	return validColumns[rand.Intn(len(validColumns))]
}
