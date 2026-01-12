package models

type ClientMessage struct {
	Type            string `json:"type"`
	JWT             string `json:"jwt"`                     // JWT token for authentication
	GameID          string `json:"gameId,omitempty"`
	Column          int    `json:"column,omitempty"`
	Difficulty      string `json:"difficulty,omitempty"`    // Bot difficulty: "easy", "medium", "hard"
	RequestRematch  bool   `json:"requestRematch,omitempty"`
	RematchResponse string `json:"rematchResponse,omitempty"` // "accept" or "decline"
}

type ServerMessage struct {
	Type             string       `json:"type"`
	Message          string       `json:"message,omitempty"`
	GameID           string       `json:"gameId,omitempty"`
	Opponent         string       `json:"opponent,omitempty"`
	YourPlayer       int          `json:"yourPlayer,omitempty"`  // 1 or 2 for board position
	CurrentTurn      int          `json:"currentTurn,omitempty"` // 1 or 2
	Column           int          `json:"column,omitempty"`
	Row              int          `json:"row,omitempty"`
	Player           int          `json:"player,omitempty"`      // 1 or 2
	Board            [][]PlayerID `json:"board,omitempty"`
	NextTurn         int          `json:"nextTurn,omitempty"`    // 1 or 2
	Winner           string       `json:"winner,omitempty"`     // username or "draw"
	Reason           string       `json:"reason,omitempty"`
	TimeRemaining    int          `json:"timeRemaining,omitempty"`
	RematchRequester string       `json:"rematchRequester,omitempty"` // username who requested rematch
	RematchTimeout   int          `json:"rematchTimeout,omitempty"`   // seconds remaining to respond
	AllowRematch     *bool        `json:"allowRematch,omitempty"`     // Controls if rematch button shows (pointer for explicit false)
}

type ErrorMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
