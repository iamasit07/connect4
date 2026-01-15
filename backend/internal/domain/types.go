package domain

import "os"

type PlayerID int

const (
	Empty PlayerID = 0
	Player1 PlayerID = 1
	Player2 PlayerID = 2
)

// Bot username constant
// import bot username from .env
var BotUsername = os.Getenv("BOT_USERNAME")

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