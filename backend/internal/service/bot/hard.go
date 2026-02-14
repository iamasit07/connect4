package bot

import (
	"math"

	"github.com/iamasit07/4-in-a-row/backend/internal/domain"
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

// CalculateBestMoveMinimax implements hard difficulty using Minimax with alpha-beta pruning
func CalculateBestMoveMinimax(board [][]domain.PlayerID, botPlayer domain.PlayerID) int {
	validColumns := domain.GetValidMoves(board)
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
		testBoard, row, _ := domain.SimulateMove(board, col, botPlayer)
		
		// If this move wins immediately, take it
		if domain.CheckWin(testBoard, row, col, botPlayer) {
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
func minimax(board [][]domain.PlayerID, depth int, alpha, beta int, isMaximizing bool, botPlayer, opponent domain.PlayerID) int {
	validColumns := domain.GetValidMoves(board)

	// Terminal conditions
	if depth == 0 || len(validColumns) == 0 {
		return evaluateBoard(board, botPlayer, opponent)
	}

	if isMaximizing {
		maxEval := math.MinInt32
		for _, col := range validColumns {
			testBoard, row, _ := domain.SimulateMove(board, col, botPlayer)
			
			// Check for win
			if domain.CheckWin(testBoard, row, col, botPlayer) {
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
			testBoard, row, _ := domain.SimulateMove(board, col, opponent)
			
			// Check for opponent win
			if domain.CheckWin(testBoard, row, col, opponent) {
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