package bot

import (
	"math/rand"

	"github.com/iamasit07/connect4/backend/internal/domain"
)
func CalculateBestMoveEasy(board [][]domain.PlayerID, botPlayer domain.PlayerID) int {
	validColumns := domain.GetValidMoves(board)
	if len(validColumns) == 0 {
		return -1
	}

	opponent := getOpponent(botPlayer)

	for _, col := range validColumns {
		testBoard, row, _ := domain.SimulateMove(board, col, botPlayer)
		_, won := domain.CheckWin(testBoard, row, col, botPlayer)
			if won {
			return col
		}
	}

	for _, col := range validColumns {
		testBoard, row, _ := domain.SimulateMove(board, col, opponent)
		if _, won := domain.CheckWin(testBoard, row, col, opponent); won {
			return col
		}
	}

	return validColumns[rand.Intn(len(validColumns))]
}
