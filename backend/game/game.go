package game

import "github.com/iamasit07/4-in-a-row/backend/models"

// the purpose of this file is to check the overall game logic

type Game struct {
	Board [][]models.PlayerID
	CurrentPlayer models.PlayerID
	Status models.GameStatus
	Winner models.PlayerID
	MoveCount int
}

func(g *Game) NewGame() *Game {
	board := make([][]models.PlayerID, models.Rows)
	for i := range board {
		board[i] = make([]models.PlayerID, models.Columns)
	}

	return &Game{
		Board: board,
		CurrentPlayer: models.Player1,
		Status: models.StatusActive,
		Winner: models.Empty,
		MoveCount: 0,
	}
}

func (g *Game) MakeMove(player models.PlayerID, column int) (int, error) {
	if g.Status != models.StatusActive {
		return -1, models.ErrInvalidMove
	}

	if !IsValidMove(g.Board, column) {
		return -1, models.ErrInvalidMove
	}

	row, err := DropDisk(g.Board, column, g.CurrentPlayer)
	if err != nil {
		return -1, err
	}
	
	g.MoveCount++

	if CheckWin(g.Board, row, column, g.CurrentPlayer) {
		g.Status = models.StatusWon
		g.Winner = g.CurrentPlayer
		return -1, nil
	}

	if IsBoardFull(g.Board) {
		g.Status = models.StatusDraw
		return -1, nil
	}

	// switch player
	if g.CurrentPlayer == models.Player1 {
		g.CurrentPlayer = models.Player2
	} else {
		g.CurrentPlayer = models.Player1
	}
	
	return row, nil
}

func (g *Game) IsFinished() bool {
	return g.Status == models.StatusWon || g.Status == models.StatusDraw
}