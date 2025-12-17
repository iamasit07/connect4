package bot

import "github.com/iamasit07/4-in-a-row/backend/models"


func CalculateBestMove(board [][]models.Player, botPlayer models.Player) int {
	validMoves := GetValidMoves(board)

}

func SimulateMove(board [][]models.Player, column int, player models.Player) (int, [][]models.Player) {
	copyBoard := board
	DropDisc(copyBoard, column, player)
}