package bot

import (
	"github.com/iamasit07/4-in-a-row/backend/internal/domain"
)

const (
	// Score priorities (from highest to lowest)
	SCORE_WIN_NOW           = 100000 // Bot can win immediately
	SCORE_BLOCK_WIN         = 10000  // Block opponent's immediate win
	SCORE_CREATE_WIN_THREAT = 8000   // Create a position where bot can win next move
	SCORE_BLOCK_WIN_THREAT  = 5000   // Block opponent's potential win setup
	SCORE_THREE_IN_ROW      = 400    // Bot has 3 in a row (good threat)
	SCORE_BLOCK_THREE       = 300    // Block opponent's 3 in a row
	SCORE_TWO_IN_ROW        = 100    // Bot has 2 in a row
	SCORE_BLOCK_TWO         = 50     // Block opponent's 2 in a row
	SCORE_CENTER            = 30     // Center column bonus
	SCORE_NEAR_CENTER       = 20     // Near center bonus
	SCORE_EDGE              = 5      // Edge columns
)

// evaluateBoard calculates a heuristic score for the current board position
func evaluateBoard(board [][]domain.PlayerID, botPlayer, opponent domain.PlayerID) int {
	score := 0

	// Evaluate all positions on the board
	for row := 0; row < domain.Rows; row++ {
		for col := 0; col < domain.Columns; col++ {
			switch board[row][col] {
			case botPlayer:
				score += evaluatePosition(board, row, col, botPlayer)
			case opponent:
				score -= evaluatePosition(board, row, col, opponent)
			}
		}
	}

	// Center column preference
	centerCol := domain.Columns / 2
	for row := 0; row < domain.Rows; row++ {
		switch board[row][centerCol] {
		case botPlayer:
			score += POSITION_WEIGHT * 2
		case opponent:
			score -= POSITION_WEIGHT * 2
		}
	}

	return score
}

// evaluatePosition evaluates a single position's contribution to the score
func evaluatePosition(board [][]domain.PlayerID, row, col int, player domain.PlayerID) int {
	score := POSITION_WEIGHT

	directions := [][2]int{
		{0, 1},  // horizontal
		{1, 0},  // vertical
		{1, 1},  // diagonal \
		{1, -1}, // diagonal /
	}

	for _, dir := range directions {
		dRow, dCol := dir[0], dir[1]
		
		// Count consecutive pieces in both directions
		posCount := domain.CountDiskInDirection(board, row, col, dRow, dCol, player)
		negCount := domain.CountDiskInDirection(board, row, col, -dRow, -dCol, player)
		total := posCount + negCount

		// Check if there's room to extend
		hasSpace := checkSpaceForExtension(board, row, col, dRow, dCol, posCount, negCount)

		if hasSpace {
			if total >= 3 {
				score += THREE_IN_ROW_WEIGHT
			} else if total == 2 {
				score += TWO_IN_ROW_WEIGHT
			}
		}
	}

	return score
}

// Evaluate threats (3-in-a-row, 2-in-a-row) for a given position
func evaluateThreats(board [][]domain.PlayerID, row, col int, player domain.PlayerID) int {
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
		posCount := domain.CountDiskInDirection(board, row, col, dRow, dCol, player)
		negCount := domain.CountDiskInDirection(board, row, col, -dRow, -dCol, player)
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
func evaluateWinningThreat(board [][]domain.PlayerID, player domain.PlayerID, opponent domain.PlayerID) int {
	validMoves := domain.GetValidMoves(board)
	winningMoves := []int{}

	// Find all columns where player can win immediately
	for _, col := range validMoves {
		testBoard, row, _ := domain.SimulateMove(board, col, player)
		if domain.CheckWin(testBoard, row, col, player) {
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
		blockBoard, _, _ := domain.SimulateMove(board, winCol, opponent)

		// After block, can player still create threats?
		stillHasThreat := false
		nextMoves := domain.GetValidMoves(blockBoard)
		for _, nextCol := range nextMoves {
			futureBoard, futureRow, _ := domain.SimulateMove(blockBoard, nextCol, player)
			if domain.CheckWin(futureBoard, futureRow, nextCol, player) {
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

// Helper: check if there's room to extend a line
func checkSpaceForExtension(board [][]domain.PlayerID, row, col, dRow, dCol, posCount, negCount int) bool {
	// Check positive direction
	posRow := row + dRow*(posCount+1)
	posCol := col + dCol*(posCount+1)
	if isInBounds(posRow, posCol) && board[posRow][posCol] == domain.Empty && isPlayableSpace(board, posRow, posCol) {
		return true
	}

	// Check negative direction
	negRow := row - dRow*(negCount+1)
	negCol := col - dCol*(negCount+1)
	if isInBounds(negRow, negCol) && board[negRow][negCol] == domain.Empty && isPlayableSpace(board, negRow, negCol) {
		return true
	}

	return false
}

// Check if a space is actually playable (respects gravity)
func isPlayableSpace(board [][]domain.PlayerID, row, col int) bool {
	// Bottom row is always playable
	if row == domain.Rows-1 {
		return true
	}
	// Otherwise, must have a piece (any player) directly below
	return board[row+1][col] != domain.Empty
}

// Helper: check if position is within board bounds
func isInBounds(row, col int) bool {
	return row >= 0 && row < domain.Rows && col >= 0 && col < domain.Columns
}
