package bot

import (
	"math"
	"math/rand"

	"github.com/iamasit07/connect4/backend/internal/domain"
)

const scoreJitterPercent = 25

func calculateMediumMove(board [][]domain.PlayerID, botPlayer domain.PlayerID) int {
	validColumns := domain.GetValidMoves(board)
	if len(validColumns) == 0 {
		return -1
	}

	scores := make(map[int]int)
	opponent := getOpponent(botPlayer)

	for _, col := range validColumns {
		scores[col] = 0
	}

	// Pre-calculate simulated boards for each column to avoid redundant simulations
	botSimulations := make(map[int]struct {
		board [][]domain.PlayerID
		row   int
	})
	oppSimulations := make(map[int]struct {
		board [][]domain.PlayerID
		row   int
	})

	for _, col := range validColumns {
		botBoard, botRow, _ := domain.SimulateMove(board, col, botPlayer)
		botSimulations[col] = struct {
			board [][]domain.PlayerID
			row   int
		}{botBoard, botRow}

		oppBoard, oppRow, _ := domain.SimulateMove(board, col, opponent)
		oppSimulations[col] = struct {
			board [][]domain.PlayerID
			row   int
		}{oppBoard, oppRow}
	}

	// === PHASE 1: Check for immediate wins (highest priority) ===
	for _, col := range validColumns {
		sim := botSimulations[col]
		if _, won := domain.CheckWin(sim.board, sim.row, col, botPlayer); won {
			scores[col] += SCORE_WIN_NOW
		}
	}

	// === PHASE 2: Block opponent's immediate wins ===
	for _, col := range validColumns {
		sim := oppSimulations[col]
		if _, won := domain.CheckWin(sim.board, sim.row, col, opponent); won {
			scores[col] += SCORE_BLOCK_WIN
		}
	}

	// === PHASE 3: Look ahead - Create winning threats (3 steps with opponent response) ===
	for _, col := range validColumns {
		sim := botSimulations[col]
		threatScore := evaluateWinningThreat(sim.board, botPlayer, opponent)
		scores[col] += threatScore
	}

	// === PHASE 4: Block opponent's winning threats (3 steps) ===
	for _, col := range validColumns {
		sim := botSimulations[col]

		// After bot moves, what's opponent's best winning threat?
		opponentThreatScore := evaluateWinningThreat(sim.board, opponent, botPlayer)

		// If this move REDUCES opponent's threat, give it points
		currentOpponentThreat := evaluateWinningThreat(board, opponent, botPlayer)
		if opponentThreatScore < currentOpponentThreat {
			scores[col] += SCORE_BLOCK_WIN_THREAT
		}
	}

	// === PHASE 5: Evaluate current position strength ===
	for _, col := range validColumns {
		botSim := botSimulations[col]
		botThreats := evaluateThreats(botSim.board, botSim.row, col, botPlayer)
		scores[col] += botThreats

		// Count and block opponent's threats
		oppSim := oppSimulations[col]
		oppThreats := evaluateThreats(oppSim.board, oppSim.row, col, opponent)
		scores[col] += oppThreats / 2 // Half value for blocking vs creating
	}

	// === PHASE 6: Positional bonuses (center preference) ===
	center := domain.Columns / 2
	for _, col := range validColumns {
		distFromCenter := col - center
		if distFromCenter < 0 {
			distFromCenter = -distFromCenter
		}

		switch distFromCenter {
		case 0:
			scores[col] += SCORE_CENTER
		case 1:
			scores[col] += SCORE_NEAR_CENTER
		case 2:
			scores[col] += SCORE_EDGE
		}
	}

	// === PHASE 7: Add jitter so identical positions don't always play identically ===
	for _, col := range validColumns {
		base := scores[col]
		if base < 0 {
			base = -base
		}
		if base == 0 {
			base = SCORE_NEAR_CENTER // minimum reference so zero-score cols vary too
		}
		maxJitter := base * scoreJitterPercent / 100
		if maxJitter < 1 {
			maxJitter = 1
		}
		scores[col] += rand.Intn(maxJitter*2+1) - maxJitter
	}

	return findBestColumn(scores)
}

// Find the column with the highest score
func findBestColumn(scores map[int]int) int {
	maxScore := -999999
	bestColumn := 3 // Default to center

	for col := 0; col < domain.Columns; col++ {
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
