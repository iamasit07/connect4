package domain

var BotNames = map[string]string{
	"easy":   "Alice",
	"medium": "Bob",
	"hard":   "Charles",
}

func GetBotName(difficulty string) string {
	if name, ok := BotNames[difficulty]; ok {
		return name
	}
	return "BOT"
}

func IsBotName(username string) bool {
	if username == "BOT" {
		return true
	}
	for _, name := range BotNames {
		if username == name {
			return true
		}
	}
	return false
}

type PlayerID int

const (
	Empty PlayerID = 0
	Player1 PlayerID = 1
	Player2 PlayerID = 2
)

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