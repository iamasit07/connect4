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

func CalculateBestMove(board [][]models.PlayerID, botPlayer models.PlayerID, difficulty string) int {
	parsedDifficulty := ParseDifficulty(difficulty)

	switch parsedDifficulty {
	case DifficultyEasy:
		return calculateEasyMove(board, botPlayer)
	case DifficultyMedium:
		return calculateMediumMove(board, botPlayer)
	case DifficultyHard:
		return calculateHardMove(board, botPlayer)
	default:
		return calculateMediumMove(board, botPlayer)
	}
}

func calculateMediumMove(board [][]models.PlayerID, botPlayer models.PlayerID) int {
	validColumns := utils.GetValidMoves(board)
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
		board [][]models.PlayerID
		row   int
	})
	oppSimulations := make(map[int]struct {
		board [][]models.PlayerID
		row   int
	})

	for _, col := range validColumns {
		botBoard, botRow, _ := utils.SimulateMove(board, col, botPlayer)
		botSimulations[col] = struct {
			board [][]models.PlayerID
			row   int
		}{botBoard, botRow}

		oppBoard, oppRow, _ := utils.SimulateMove(board, col, opponent)
		oppSimulations[col] = struct {
			board [][]models.PlayerID
			row   int
		}{oppBoard, oppRow}
	}

	// === PHASE 1: Check for immediate wins (highest priority) ===
	for _, col := range validColumns {
		sim := botSimulations[col]
		if game.CheckWin(sim.board, sim.row, col, botPlayer) {
			scores[col] += SCORE_WIN_NOW
		}
	}

	// === PHASE 2: Block opponent's immediate wins ===
	for _, col := range validColumns {
		sim := oppSimulations[col]
		if game.CheckWin(sim.board, sim.row, col, opponent) {
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
	center := models.Columns / 2
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

	return findBestColumn(scores)
}

// Check if there's room to extend a line (critical for real threats)
func checkSpaceForExtension(board [][]models.PlayerID, row, col, dRow, dCol, posCount, negCount int) bool {
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
