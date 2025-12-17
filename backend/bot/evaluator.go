package bot

import (
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

func evaluateStrategicValue(board [][]models.Player, row, column int, botPlayer models.Player) int {
	score := 0
	directions := [][2]int{
		{0, 1},  // horizontal
		{1, 0},  // vertical
		{-1, 1}, // diagonal with "/"
		{1, 1},  // diagonal with "\"
	}
	
	for _, dir := range directions {
		deltaRow, deltaCol := dir[0], dir[1]
		posCount := utils.CountDiskInDirection(board, row, column, deltaRow, deltaCol, botPlayer)
		negCount := utils.CountDiskInDirection(board, row, column, -deltaRow, -deltaCol, botPlayer)
		total := posCount + negCount + 1 

		if total >= 3 {
			score += 100
		} else if total == 2 {
			score += 50
		}
	}
	return score
}

// this simply checks the "threat" or the no of disks in a direction for the bot to block
func threatDetection(board [][]models.Player, row, column int, deltaRow, deltaCol int, opponent models.Player) int {
	positiveCount := utils.CountDiskInDirection(board, row, column, deltaRow, deltaCol, opponent)
	negativeCount := utils.CountDiskInDirection(board, row, column, -deltaRow, -deltaCol, opponent)
	return positiveCount + negativeCount + 1
	
}