package domain

type Game struct {
	Board [][]PlayerID
	CurrentPlayer PlayerID
	Status GameStatus
	Winner PlayerID
	MoveCount int
}

func(g *Game) NewGame() *Game {
	return &Game{
		Board: NewBoard(),
		CurrentPlayer: Player1,
		Status: StatusActive,
		Winner: Empty,
		MoveCount: 0,
	}
}

func (g *Game) MakeMove(player PlayerID, column int) (int, error) {
	if g.Status != StatusActive {
		return -1, ErrInvalidMove
	}

	if !IsValidMove(g.Board, column) {
		return -1, ErrInvalidMove
	}

	row, err := DropDisk(g.Board, column, g.CurrentPlayer)
	if err != nil {
		return -1, err
	}
	
	g.MoveCount++

	if CheckWin(g.Board, row, column, g.CurrentPlayer) {
		g.Status = StatusWon
		g.Winner = g.CurrentPlayer
		return row, nil
	}

	if IsBoardFull(g.Board) {
		g.Status = StatusDraw
		return row, nil
	}

	if g.CurrentPlayer == Player1 {
		g.CurrentPlayer = Player2
	} else {
		g.CurrentPlayer = Player1
	}
	
	return row, nil
}

func (g *Game) IsFinished() bool {
	return g.Status == StatusWon || g.Status == StatusDraw
}