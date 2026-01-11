package bot

import (
	"math"

	"github.com/iamasit07/4-in-a-row/backend/game"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

const (
	MINIMAX_DEPTH     = 7
	MINIMAX_WIN       = 1000000
	MINIMAX_LOSS      = -1000000
	MINIMAX_DRAW      = 0
	POSITION_WEIGHT   = 10
	TWO_IN_ROW_WEIGHT = 50
	THREE_IN_ROW_WEIGHT = 500
)

// calculateHardMove implements hard difficulty using Minimax with alpha-beta pruning
func calculateHardMove(board [][]models.PlayerID, botPlayer models.PlayerID) int {
	validColumns := utils.GetValidMoves(board)
	if len(validColumns) == 0 {
		return -1
	}

	bestCol := validColumns[0]
	bestScore := math.MinInt32
	alpha := math.MinInt32
	beta := math.MaxInt32

	opponent := getOpponent(botPlayer)

	// Try each valid column
	for _, col := range validColumns {
		testBoard, row, _ := utils.SimulateMove(board, col, botPlayer)
		
		// If this move wins immediately, take it
		if game.CheckWin(testBoard, row, col, botPlayer) {
			return col
		}

		// Calculate score using minimax
		score := minimax(testBoard, MINIMAX_DEPTH-1, alpha, beta, false, botPlayer, opponent)

		if score > bestScore {
			bestScore = score
			bestCol = col
		}

		alpha = max(alpha, bestScore)
	}

	return bestCol
}

// minimax implements the minimax algorithm with alpha-beta pruning
func minimax(board [][]models.PlayerID, depth int, alpha, beta int, isMaximizing bool, botPlayer, opponent models.PlayerID) int {
	validColumns := utils.GetValidMoves(board)

	// Terminal conditions
	if depth == 0 || len(validColumns) == 0 {
		return evaluateBoard(board, botPlayer, opponent)
	}

	if isMaximizing {
		maxEval := math.MinInt32
		for _, col := range validColumns {
			testBoard, row, _ := utils.SimulateMove(board, col, botPlayer)
			
			// Check for win
			if game.CheckWin(testBoard, row, col, botPlayer) {
				return MINIMAX_WIN - (MINIMAX_DEPTH - depth) // Prefer quicker wins
			}

			eval := minimax(testBoard, depth-1, alpha, beta, false, botPlayer, opponent)
			maxEval = max(maxEval, eval)
			alpha = max(alpha, eval)
			
			if beta <= alpha {
				break // Beta cutoff
			}
		}
		return maxEval
	} else {
		minEval := math.MaxInt32
		for _, col := range validColumns {
			testBoard, row, _ := utils.SimulateMove(board, col, opponent)
			
			// Check for opponent win
			if game.CheckWin(testBoard, row, col, opponent) {
				return MINIMAX_LOSS + (MINIMAX_DEPTH - depth) // Prefer delaying losses
			}

			eval := minimax(testBoard, depth-1, alpha, beta, true, botPlayer, opponent)
			minEval = min(minEval, eval)
			beta = min(beta, eval)
			
			if beta <= alpha {
				break // Alpha cutoff
			}
		}
		return minEval
	}
}

// evaluateBoard calculates a heuristic score for the current board position
func evaluateBoard(board [][]models.PlayerID, botPlayer, opponent models.PlayerID) int {
	score := 0

	// Evaluate all positions on the board
	for row := 0; row < models.Rows; row++ {
		for col := 0; col < models.Columns; col++ {
			if board[row][col] == botPlayer {
				score += evaluatePosition(board, row, col, botPlayer, opponent)
			} else if board[row][col] == opponent {
				score -= evaluatePosition(board, row, col, opponent, botPlayer)
			}
		}
	}

	// Center column preference
	centerCol := models.Columns / 2
	for row := 0; row < models.Rows; row++ {
		if board[row][centerCol] == botPlayer {
			score += POSITION_WEIGHT * 2
		} else if board[row][centerCol] == opponent {
			score -= POSITION_WEIGHT * 2
		}
	}

	return score
}

// evaluatePosition evaluates a single position's contribution to the score
func evaluatePosition(board [][]models.PlayerID, row, col int, player, opponent models.PlayerID) int {
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
		posCount := utils.CountDiskInDirection(board, row, col, dRow, dCol, player)
		negCount := utils.CountDiskInDirection(board, row, col, -dRow, -dCol, player)
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
