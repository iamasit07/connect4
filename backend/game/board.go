package game

import "github.com/iamasit07/4-in-a-row/backend/models"

func NewGame() [models.Rows][models.Columns]int {
	return [models.Rows][models.Columns]int{}
}

func IsValidMove(board [models.Rows][models.Columns] int, column int) bool {
	if column < 0 || column >= models.Columns {
		return false
	}

	// here board[0] represents the top row (0 -> top and 5 -> bottom)
	if board[0][column] != 0{
		return false
	}

	return true
}

func DropDisk(board [][]models.Player, column int, player models.Player) (int, error) {
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

func IsBoardFull(board [][]models.Player) bool {
	if board[0] != 0 {
		return true
	}

	return false
}