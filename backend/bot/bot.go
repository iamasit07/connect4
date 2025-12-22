package bot

import (
	"math"

	"github.com/iamasit07/4-in-a-row/backend/game"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

// Score priorities (from highest to lowest)
const (
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

// Main bot function with improved scoring
func CalculateBestMove(board [][]models.PlayerID, botPlayer models.PlayerID) int {
	validColumns := utils.GetValidMoves(board)
	if len(validColumns) == 0 {
		return -1
	}

	scores := make(map[int]int)
	opponent := getOpponent(botPlayer)

	// Initialize scores
	for _, col := range validColumns {
		scores[col] = 0
	}

	// === PHASE 1: Check for immediate wins (highest priority) ===
	for _, col := range validColumns {
		newBoard, row, _ := utils.SimulateMove(board, col, botPlayer)
		if game.CheckWin(newBoard, row, col, botPlayer) {
			scores[col] += SCORE_WIN_NOW
		}
	}

	// === PHASE 2: Block opponent's immediate wins ===
	for _, col := range validColumns {
		newBoard, row, _ := utils.SimulateMove(board, col, opponent)
		if game.CheckWin(newBoard, row, col, opponent) {
			scores[col] += SCORE_BLOCK_WIN
		}
	}

	// === PHASE 3: Look ahead - Create winning threats (3 steps with opponent response) ===
	for _, col := range validColumns {
		boardAfterBotMove, _, _ := utils.SimulateMove(board, col, botPlayer)

		// Check if bot can create an UNBLOCKABLE winning threat
		threatScore := evaluateWinningThreat(boardAfterBotMove, botPlayer, opponent)
		scores[col] += threatScore
	}

	// === PHASE 4: Block opponent's winning threats (3 steps) ===
	for _, col := range validColumns {
		boardAfterBotMove, _, _ := utils.SimulateMove(board, col, botPlayer)

		// After bot moves, what's opponent's best winning threat?
		opponentThreatScore := evaluateWinningThreat(boardAfterBotMove, opponent, botPlayer)

		// If this move REDUCES opponent's threat, give it points
		currentOpponentThreat := evaluateWinningThreat(board, opponent, botPlayer)
		if opponentThreatScore < currentOpponentThreat {
			scores[col] += SCORE_BLOCK_WIN_THREAT
		}
	}

	// === PHASE 5: Evaluate current position strength ===
	for _, col := range validColumns {
		newBoard, row, _ := utils.SimulateMove(board, col, botPlayer)

		// Count bot's threats
		botThreats := evaluateThreats(newBoard, row, col, botPlayer)
		scores[col] += botThreats

		// Count and block opponent's threats
		oppBoard, oppRow, _ := utils.SimulateMove(board, col, opponent)
		oppThreats := evaluateThreats(oppBoard, oppRow, col, opponent)
		scores[col] += oppThreats / 2 // Half value for blocking vs creating
	}

	// === PHASE 6: Positional bonuses (center preference) ===
	for _, col := range validColumns {
		switch col {
		case 3:
			scores[col] += SCORE_CENTER
		case 2, 4:
			scores[col] += SCORE_NEAR_CENTER
		case 1, 5:
			scores[col] += SCORE_EDGE
		}
	}

	return findBestColumn(scores)
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

// Check if there's room to extend a line (critical for real threats)
// IMPORTANT: Accounts for gravity - empty spaces must be playable (bottom row or has piece below)
func checkSpaceForExtension(board [][]models.PlayerID, row, col, dRow, dCol, posCount, negCount int, player models.PlayerID) bool {
	// Check positive direction
	posRow := row + dRow*(posCount+1)
	posCol := col + dCol*(posCount+1)
	if isInBounds(posRow, posCol) && board[posRow][posCol] == models.Empty && isPlayableSpace(board, posRow, posCol) {
		return true
	}

	// Check negative direction
	negRow := row - dRow*(negCount+1)
	negCol := col - dCol*(negCount+1)
	if isInBounds(negRow, negCol) && board[negRow][negCol] == models.Empty && isPlayableSpace(board, negRow, negCol) {
		return true
	}

	return false
}

// Check if a space is actually playable (respects gravity)
// A space is playable if it's on the bottom row OR has a piece directly below it
func isPlayableSpace(board [][]models.PlayerID, row, col int) bool {
	// Bottom row is always playable
	if row == models.Rows-1 {
		return true
	}

	// Otherwise, must have a piece (any player) directly below
	return board[row+1][col] != models.Empty
}

// Helper: check if position is within board bounds
func isInBounds(row, col int) bool {
	return row >= 0 && row < models.Rows && col >= 0 && col < models.Columns
}

// Helper: get opponent player
func getOpponent(player models.PlayerID) models.PlayerID {
	if player == models.Player1 {
		return models.Player2
	}
	return models.Player1
}

// Find the column with the highest score
func findBestColumn(scores map[int]int) int {
	maxScore := -999999
	bestColumn := 3 // Default to center

	for col := 0; col < models.Columns; col++ {
		score, exists := scores[col]
		if !exists {
			continue
		}

		if score > maxScore {
			maxScore = score
			bestColumn = col
		} else if score == maxScore {
			// Tie-breaker: prefer columns closer to center
			if math.Abs(float64(col-3)) < math.Abs(float64(bestColumn-3)) {
				bestColumn = col
			}
		}
	}

	return bestColumn
}
