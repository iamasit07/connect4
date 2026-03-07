package bot

import (
	"math/rand"

	"github.com/iamasit07/connect4/backend/internal/domain"
)

const easyMistakeChance = 40

func CalculateBestMoveEasy(board [][]domain.PlayerID, botPlayer domain.PlayerID) int {
	validColumns := domain.GetValidMoves(board)
	if len(validColumns) == 0 {
		return -1
	}

	if rand.Intn(100) >= easyMistakeChance {
		for _, col := range validColumns {
			testBoard, row, _ := domain.SimulateMove(board, col, botPlayer)
			if _, won := domain.CheckWin(testBoard, row, col, botPlayer); won {
				return col
			}
		}
	}

	return validColumns[rand.Intn(len(validColumns))]
}
