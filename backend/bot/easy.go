package bot

import (
	"math/rand"

	"github.com/iamasit07/4-in-a-row/backend/game"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

// calculateEasyMove implements easy difficulty bot
// Strategy: One-step lookahead only
// 1. Check if bot can win immediately
// 2. Check if opponent can win immediately and block
// 3. Otherwise, make a random valid move
func calculateEasyMove(board [][]models.PlayerID, botPlayer models.PlayerID) int {
	validColumns := utils.GetValidMoves(board)
	if len(validColumns) == 0 {
		return -1
	}

	opponent := getOpponent(botPlayer)

	// PRIORITY 1: Check if bot can win immediately
	for _, col := range validColumns {
		testBoard, row, _ := utils.SimulateMove(board, col, botPlayer)
		if game.CheckWin(testBoard, row, col, botPlayer) {
			return col
		}
	}

	// PRIORITY 2: Block opponent's immediate win
	for _, col := range validColumns {
		testBoard, row, _ := utils.SimulateMove(board, col, opponent)
		if game.CheckWin(testBoard, row, col, opponent) {
			return col // Block this winning move
		}
	}

	// PRIORITY 3: Make a random valid move
	return validColumns[rand.Intn(len(validColumns))]
}
