package models

// to represent the players in the game
type Player int

const (
	Empty Player = 0
	Player1 Player = 1
	Player2 Player = 2
)

//for board representation 
const (
	Rows = 6
	Columns = 7
	ToWin = 4
)

// to represent the game status
type GameStatus string

const (
	StatusActive   GameStatus = "active"
	StatusWon GameStatus = "won"
	StatusDraw GameStatus = "draw"
)

// basic error that can occur
type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrInvalidMove Error = "invalid move"
	ErrColumnFull  Error = "column is full"
)