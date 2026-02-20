package domain

func CheckWin(board [][]PlayerID, row, column int, player PlayerID) ([]int, bool) {
	toIndex := func(r, c int) int {
		return r*Columns + c
	}

	// Check horizontal (through this row)
	count := 0
	for c := 0; c < Columns; c++ {
		if board[row][c] == player {
			count++
			if count == 4 {
				return []int{
					toIndex(row, c),
					toIndex(row, c-1),
					toIndex(row, c-2),
					toIndex(row, c-3),
				}, true
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
				return []int{
					toIndex(r, column),
					toIndex(r-1, column),
					toIndex(r-2, column),
					toIndex(r-3, column),
				}, true
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
				return []int{
					toIndex(startRow, startCol),
					toIndex(startRow-1, startCol-1),
					toIndex(startRow-2, startCol-2),
					toIndex(startRow-3, startCol-3),
				}, true
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
				return []int{
					toIndex(startRow, startCol),
					toIndex(startRow+1, startCol-1),
					toIndex(startRow+2, startCol-2),
					toIndex(startRow+3, startCol-3),
				}, true
			}
		} else {
			count = 0
		}
		startRow--
		startCol++
	}

	return nil, false
}