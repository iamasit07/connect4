package models

type PlayerID int

const (
	Empty PlayerID = 0
	Player1 PlayerID = 1
	Player2 PlayerID = 2
)

// Bot username constant
const BotUsername = "BOT"

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