package utils

import (
	"github.com/iamasit07/4-in-a-row/backend/game"
	"github.com/iamasit07/4-in-a-row/backend/models"
)

// this creates a deep copy of the board
func CopyBoard(board [][]models.Player) [][]models.Player {
	newBoard := make([][]models.Player, len(board))
	for i := range board {
		newBoard[i] = make([]models.Player, len(board[i]))
		copy(newBoard[i], board[i])
	}
	return newBoard
}

// this is a helper function that will later be used by the bot
func GetValidMoves(game *Game) []int {
	validMoves := []int{}
	for col := 0; col < models.Columns; col++ {
		if IsValidMove(game.Board, col) {
			validMoves = append(validMoves, col)
		}
	}
	return validMoves
}

func SimulateMove(board [][]models.Player, column int, player models.Player) ([][]models.Player, int, error) {
	newBoard := CopyBoard(board)
	row, err := game.DropDisk(newBoard, column, player)
	if err != nil {
		return nil, -1, err
	}
	return newBoard, row, nil
}