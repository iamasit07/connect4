package models

type ClientMessage struct {
	Type   string `json:"type"`
	JWT    string `json:"jwt"`              // JWT token for authentication
	GameID string `json:"gameId,omitempty"`
	Column int    `json:"column,omitempty"`
}

type ServerMessage struct {
	Type          string       `json:"type"`
	Message       string       `json:"message,omitempty"`
	GameID        string       `json:"gameId,omitempty"`
	Opponent      string       `json:"opponent,omitempty"`
	YourPlayer    int          `json:"yourPlayer,omitempty"`  // 1 or 2 for board position
	CurrentTurn   int          `json:"currentTurn,omitempty"` // 1 or 2
	Column        int          `json:"column,omitempty"`
	Row           int          `json:"row,omitempty"`
	Player        int          `json:"player,omitempty"`      // 1 or 2
	Board         [][]PlayerID `json:"board,omitempty"`
	NextTurn      int          `json:"nextTurn,omitempty"`    // 1 or 2
	Winner        string       `json:"winner,omitempty"`     // username or "draw"
	Reason        string       `json:"reason,omitempty"`
	TimeRemaining int          `json:"timeRemaining,omitempty"`
}

type ErrorMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
