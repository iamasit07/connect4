package game

import "github.com/iamasit07/4-in-a-row/backend/models"

func NewBoard() [][]models.PlayerID {
	board := make([][]models.PlayerID, models.Rows)
	for i := range board {
		board[i] = make([]models.PlayerID, models.Columns)
	}
	return board
}

func IsValidMove(board [][]models.PlayerID, column int) bool {
	if column < 0 || column >= models.Columns {
		return false
	}

	// here board[0] represents the top row (0 -> top and 5 -> bottom)
	if board[0][column] != 0{
		return false
	}

	return true
}

func DropDisk(board [][]models.PlayerID, column int, player models.PlayerID) (int, error) {
	// shifting all the disk from top to bottom till it 
	// reaches the end or another disk
	for row := models.Rows - 1; row >= 0; row-- {
		if board[row][column] == models.Empty {
			board[row][column] = player
			return row, nil
		}
	}

	return -1, models.ErrColumnFull
}

func IsBoardFull(board [][]models.PlayerID) bool {
	for c := 0; c < models.Columns; c++ {
		if board[0][c] == models.Empty {
			return false
		}
	}

	return true
}