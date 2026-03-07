package bot

import (
	"math"
	"math/rand"

	"github.com/iamasit07/connect4/backend/internal/domain"
)

const (
	MINIMAX_DEPTH        = 7
	MINIMAX_WIN          = 1000000
	MINIMAX_LOSS         = -1000000
	MINIMAX_DRAW         = 0
	POSITION_WEIGHT      = 10
	TWO_IN_ROW_WEIGHT    = 50
	THREE_IN_ROW_WEIGHT  = 500
	hardTopTierTolerance = 0.04
)

func CalculateBestMoveMinimax(board [][]domain.PlayerID, botPlayer domain.PlayerID) int {
	validColumns := domain.GetValidMoves(board)
	if len(validColumns) == 0 {
		return -1
	}

	opponent := getOpponent(botPlayer)

	// Immediate-win shortcut — always take it (no randomness needed here).
	for _, col := range validColumns {
		testBoard, row, _ := domain.SimulateMove(board, col, botPlayer)
		if _, won := domain.CheckWin(testBoard, row, col, botPlayer); won {
			return col
		}
	}

	alpha := math.MinInt32
	beta := math.MaxInt32

	type colScore struct {
		col   int
		score int
	}
	results := make([]colScore, 0, len(validColumns))

	for _, col := range validColumns {
		testBoard, _, _ := domain.SimulateMove(board, col, botPlayer)
		score := minimax(testBoard, MINIMAX_DEPTH-1, alpha, beta, false, botPlayer, opponent)
		results = append(results, colScore{col, score})
		if score > alpha {
			alpha = score
		}
	}

	// Find the best score across all moves.
	bestScore := math.MinInt32
	for _, r := range results {
		if r.score > bestScore {
			bestScore = r.score
		}
	}

	// Collect all moves within tolerance of the best score.
	tolerance := int(math.Abs(float64(bestScore)) * hardTopTierTolerance)
	if tolerance < 1 {
		tolerance = 1 // always allow at least a 1-point band
	}

	topMoves := make([]int, 0, len(results))
	for _, r := range results {
		if r.score >= bestScore-tolerance {
			topMoves = append(topMoves, r.col)
		}
	}

	// Pick randomly among the top-tier moves.
	return topMoves[rand.Intn(len(topMoves))]
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
			if _, won := domain.CheckWin(testBoard, row, col, botPlayer); won {
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
			if _, won := domain.CheckWin(testBoard, row, col, opponent); won {
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
