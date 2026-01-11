package bot

import (
	"github.com/iamasit07/4-in-a-row/backend/game"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

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

// Evaluate winning threats considering opponent's best response
// Returns score based on how many UNBLOCKABLE winning moves the player has
func evaluateWinningThreat(board [][]models.PlayerID, player models.PlayerID, opponent models.PlayerID) int {
	validMoves := utils.GetValidMoves(board)
	winningMoves := []int{}

	// Find all columns where player can win immediately
	for _, col := range validMoves {
		testBoard, row, _ := utils.SimulateMove(board, col, player)
		if game.CheckWin(testBoard, row, col, player) {
			winningMoves = append(winningMoves, col)
		}
	}

	// If player has 2+ winning moves, opponent can only block one = UNBLOCKABLE
	if len(winningMoves) >= 2 {
		return SCORE_CREATE_WIN_THREAT // This is a winning position!
	}

	// If player has 1 winning move, check if opponent can block it
	if len(winningMoves) == 1 {
		winCol := winningMoves[0]

		// Simulate opponent blocking
		blockBoard, _, _ := utils.SimulateMove(board, winCol, opponent)

		// After block, can player still create threats?
		stillHasThreat := false
		nextMoves := utils.GetValidMoves(blockBoard)
		for _, nextCol := range nextMoves {
			futureBoard, futureRow, _ := utils.SimulateMove(blockBoard, nextCol, player)
			if game.CheckWin(futureBoard, futureRow, nextCol, player) {
				stillHasThreat = true
				break
			}
		}

		if stillHasThreat {
			return SCORE_CREATE_WIN_THREAT / 2 // Blockable but still valuable
		}

		return SCORE_CREATE_WIN_THREAT / 4 // Easily blockable
	}

	return 0 // No immediate winning threat
}
