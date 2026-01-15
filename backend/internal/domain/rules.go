package domain

func CheckWin(board [][]PlayerID, row, column int, player PlayerID) bool{
	// Only check lines passing through the specific position (row, column)
	// This is more efficient than scanning the entire board

	// Check horizontal (through this row)
	count := 0
	for c := 0; c < Columns; c++ {
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
	for r := 0; r < Rows; r++ {
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
	for startRow < Rows && startCol < Columns {
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
	for startRow < Rows-1 && startCol > 0 {
		startRow++
		startCol--
	}
	// Scan the diagonal
	for startRow >= 0 && startCol < Columns {
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