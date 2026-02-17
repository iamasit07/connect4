package domain

import "math"

const KFactor = 32.0

// CalculateElo returns the new rating for player A.
// score is 1.0 for a win, 0.5 for a draw, and 0.0 for a loss.
func CalculateElo(ratingA, ratingB int, score float64) int {
	expectedScoreA := 1.0 / (1.0 + math.Pow(10.0, float64(ratingB-ratingA)/400.0))
	newRating := float64(ratingA) + KFactor*(score-expectedScoreA)

	if newRating < 0 {
		return 0
	}
	return int(newRating)
}
