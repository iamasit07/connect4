package bot

import (
	"github.com/iamasit07/4-in-a-row/backend/game"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

// this is the most important part of the bot
// it is consisted of three main parts
// 1. check if it can win in the next move
// 2. check if the opponent can win in the next move and block them
// 3. evaluate the strategic value of each valid move and choose the best one

// **MORE EXPLANATION IN README**

func CalculateBestMove(board [][]models.Player, botPlayer models.Player) int {
	validColumns := utils.GetValidMoves(board)
	scores := make(map[int]int)

	for _, col := range validColumns {
		scores[col] = 0
	}
	
	// win in the next move
	for _ , col := range validColumns {
		newBoard, row, _ := utils.SimulateMove(board, col, botPlayer)
		if game.CheckWin(newBoard, row, col, botPlayer) {
			return col
		}
	}

	opponent := models.Player1
	if botPlayer == models.Player1 {
		opponent = models.Player2
	}

	// block oppenent from winning
	for _, col := range validColumns {
		simulatedBoard, row, _ := utils.SimulateMove(board, col, opponent)
		if game.CheckWin(simulatedBoard, row, col, opponent) {
			scores[col] += 500
		}
	}

	for _, col := range validColumns {
		newBoard, row, _ := utils.SimulateMove(board, col, botPlayer)
		strategicValue := evaluateStrategicValue(newBoard, row, col, botPlayer)
		scores[col] += strategicValue
	}

	// creating a bias towards the center columns
	scores[3] += 20
	scores[2] += 15
	scores[4] += 15
	scores[1] += 5
	scores[5] += 5

	return findBestColumn(scores)
}
