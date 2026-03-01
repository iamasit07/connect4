package domain

type GameEventType string

const (
	EventGameStart       GameEventType = "game_start"
	EventMoveMade        GameEventType = "move_made"
	EventGameOver        GameEventType = "game_over"
	EventSpectatorJoined GameEventType = "spectator_joined"
	EventQueueJoined     GameEventType = "queue_joined"
	EventQueueLeft       GameEventType = "queue_left"
	EventError           GameEventType = "error"
	EventInfo            GameEventType = "info"
)

type GameEvent struct {
	Type       GameEventType
	Recipients []int64
	Payload    interface{}
}
