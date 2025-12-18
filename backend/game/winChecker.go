package game

import "github.com/iamasit07/4-in-a-row/backend/models"

func CheckWin(board [][]models.PlayerID, row, column int, player models.PlayerID) bool{
	// to check the win we have to check in 4 diff directions

	// horizontal 
	count := 0
	for c:=0; c<models.Columns; c++ {
		if board[row][c] == player{
			count++
			if count == 4 {
				return true
			}
		} else {
			count = 0
		}
	}

	// vertical
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

	// upward diagonal (like this "\")
	count = 0
	startRow := row
	startCol := column
	for startRow > 0 && startCol > 0 {
		startRow--
		startCol--
	}
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

	// downward diagonal (like this "/")
	count = 0
	startRow = row
	startCol = column
	for startRow < models.Rows-1 && startCol > 0 {
		startRow++
		startCol--
	}
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