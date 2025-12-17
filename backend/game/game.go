package game

import "github.com/iamasit07/4-in-a-row/backend/models"

// the purpose of this file is to check the overall game logic

type Game struct {
	Board [][]models.Player
	CurrentPlayer models.Player
	Status models.GameStatus
	Winner models.Player
	MoveCount int
}

func NewGame() *Game {
	board := make([][]models.Player, models.Rows)
	for i := range board {
		board[i] = make([]models.Player, models.Columns)
	}

	return &Game{
		Board: board,
		CurrentPlayer: models.Player1,
		Status: models.Ongoing,
		Winner: models.Empty,
		MoveCount: 0,
	}
}

func MakeMove(game *Game, column int) error {
	if game.Status != models.Ongoing {
		return models.ErrGameOver
	}

	if !IsValidMove(game.Board, column) {
		return models.ErrInvalidMove
	}

	row, err := DropDisk(game.Board, column, game.CurrentPlayer)
	if err != nil {
		return err
	}
	
	game.MoveCount++

	if CheckWin(game.Board, row, column, game.CurrentPlayer) {
		game.Status = models.Finished
		game.Winner = game.CurrentPlayer
		return nil
	}

	if IsBoardFull(game.Board) {
		game.Status = models.Draw
		return nil
	}

	// switch player
	if game.CurrentPlayer == models.Player1 {
		game.CurrentPlayer = models.Player2
	} else {
		game.CurrentPlayer = models.Player1
	}
	
	return nil
}