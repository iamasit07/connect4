package bot

import (
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

func evaluateStrategicValue(board [][]models.PlayerID, row, column int, botPlayer models.PlayerID) int {
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

// Evaluate threats (3-in-a-row, 2-in-a-row) for a given position
func evaluateThreats(board [][]models.PlayerID, row, col int, player models.PlayerID) int {
	score := 0
	directions := [][2]int{
		{0, 1},  // horizontal
		{1, 0},  // vertical
		{1, 1},  // diagonal \
		{1, -1}, // diagonal /
	}

	for _, dir := range directions {
		dRow, dCol := dir[0], dir[1]

		// Count in both directions
		posCount := utils.CountDiskInDirection(board, row, col, dRow, dCol, player)
		negCount := utils.CountDiskInDirection(board, row, col, -dRow, -dCol, player)
		total := posCount + negCount

		// Check if there's space to extend (important!)
		hasSpace := checkSpaceForExtension(board, row, col, dRow, dCol, posCount, negCount)

		if !hasSpace {
			continue // No point in counting if we can't extend
		}

		// Score based on how many connected pieces
		if total >= 3 {
			score += SCORE_THREE_IN_ROW
		} else if total == 2 {
			score += SCORE_TWO_IN_ROW
		} else if total == 1 {
			score += 25 // Single connection
		}
	}

	return score
}
