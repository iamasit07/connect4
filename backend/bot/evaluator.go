package bot

import (
	"math"

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

func findBestColumn(scores map[int]int) int {
	maxScore := -999
	bestColumn := 3

	for col := 0; col < models.Columns; col++ {
		score, exists := scores[col]
		if !exists {
			continue
		}

		if score > maxScore {
			maxScore = score
			bestColumn = col
		} else if score == maxScore {
			if math.Abs(float64(col - 3)) < math.Abs(float64(bestColumn - 3)) {
				bestColumn = col
			}
		}
	}

	return bestColumn
}