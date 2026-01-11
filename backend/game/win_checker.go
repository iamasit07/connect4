package game

import "github.com/iamasit07/4-in-a-row/backend/models"

func CheckWin(board [][]models.PlayerID, row, column int, player models.PlayerID) bool{
	// Only check lines passing through the specific position (row, column)
	// This is more efficient than scanning the entire board

	// Check horizontal (through this row)
	count := 0
	for c := 0; c < models.Columns; c++ {
		if board[row][c] == player {
			count++
			if count == 4 {
				return true
			}
		} else {
			count = 0
		}
	}

	// Check vertical (through this column)
	count = 0
	for r := 0; r < models.Rows; r++ {
		if board[r][column] == player {
			count++
			if count == 4 {
				return true
			}
		} else {
			count = 0
		}
	}

	// Check diagonal \ (through this position)
	count = 0
	// Find starting position of this diagonal
	startRow, startCol := row, column
	for startRow > 0 && startCol > 0 {
		startRow--
		startCol--
	}
	// Scan the diagonal
	for startRow < models.Rows && startCol < models.Columns {
		if board[startRow][startCol] == player {
			count++
			if count == 4 {
				return true
			}
		} else {
			count = 0
		}
		startRow++
		startCol++
	}

	// Check diagonal / (through this position)
	count = 0
	// Find starting position of this diagonal
	startRow, startCol = row, column
	for startRow < models.Rows-1 && startCol > 0 {
		startRow++
		startCol--
	}
	// Scan the diagonal
	for startRow >= 0 && startCol < models.Columns {
		if board[startRow][startCol] == player {
			count++
			if count == 4 {
				return true
			}
		} else {
			count = 0
		}
		startRow--
		startCol++
	}

	return false
}