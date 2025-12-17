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

func CalculateBestMove(g *game.Game, board [][]models.Player, botPlayer models.Player) int {
	validColumns := utils.GetValidMoves(g)
	bestColumn := -1
	maxScore := -10000
	
	// win in the next move
	for row , col := range validColumns {
		simulatedBoard, _, _ := utils.SimulateMove(board, col, botPlayer)
		if game.CheckWin(simulatedBoard, row, col, botPlayer) {
			return col
		}
	}

	opponent := models.Player1
	if botPlayer == models.Player1 {
		opponent = models.Player2
	}

	// block oppenent from winning
	for _, col := range validColumns {
		simulatedBoard, _, _ := utils.SimulateMove(board, col, opponent)
		if game.CheckWin(simulatedBoard, 0, col, opponent) {
			return col
		}
	}

	for _, col := range validColumns {
		score := evaluateStrategicValue(board,0, col, botPlayer)
		if score > maxScore {
			maxScore = score
			bestColumn = col
		}
	}

	centerPreference := map[int]int{
		3: 20,
		2: 15,
		4: 15,
		1: 5,
		5: 5,
	}
	if pref, exists := centerPreference[bestColumn]; exists {
		maxScore += pref
	}
	return bestColumn
}
