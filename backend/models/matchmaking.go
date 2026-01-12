package models

import (
	"sync"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/config"
)

type Match struct {
	Player1ID       int64
	Player1Username string
	Player2ID       *int64 // nil for BOT
	Player2Username string
	BotDifficulty   string // "easy", "medium", "hard" - only used for bot games
}

type MatchmakingQueue struct {
	WaitingPlayers map[int64]string         // userID → username
	Difficulties   map[int64]string         // userID → bot difficulty
	Mux            *sync.Mutex
	MatchChannel   chan Match
	Timer          *map[int64]*time.Timer
}

func NewMatchmakingQueue() *MatchmakingQueue {
	timerMap := make(map[int64]*time.Timer)
	waitingPlayers := make(map[int64]string) // userID → username
	difficulties := make(map[int64]string)   // userID → difficulty
	queue := &MatchmakingQueue{
		WaitingPlayers: waitingPlayers,
		Difficulties:   difficulties,
		MatchChannel:   make(chan Match, 100),
		Mux:            &sync.Mutex{},
		Timer:          &timerMap,
	}
	return queue
}

func (m *MatchmakingQueue) AddPlayerToQueue(userID int64, username string, difficulty string) error {
	m.Mux.Lock()
	defer m.Mux.Unlock()

	if _, exists := m.WaitingPlayers[userID]; exists {
		return nil
	}

	// If difficulty is specified, immediately create a bot match
	// This means user explicitly chose to play against a bot
	if difficulty != "" {
		match := Match{
			Player1ID:       userID,
			Player1Username: username,
			Player2ID:       nil,
			Player2Username: BotUsername,
			BotDifficulty:   difficulty,
		}
		m.MatchChannel <- match
		return nil
	}

	// No difficulty = online matchmaking
	if len(m.WaitingPlayers) == 0 {
		m.WaitingPlayers[userID] = username
		m.Difficulties[userID] = difficulty
		// Use config timeout (5 minutes default)
		timer := time.AfterFunc(config.AppConfig.OnlineMatchmakingTimeout, func() {
			m.HandleTimeout(userID)
		})
		(*m.Timer)[userID] = timer
	} else {
		var opponentID int64
		var opponentUsername string
		for uid, name := range m.WaitingPlayers {
			opponentID = uid
			opponentUsername = name
			break
		}

		delete(m.WaitingPlayers, opponentID)
		delete(m.Difficulties, opponentID)
		m.stopAndDeleteTimer(opponentID)

		match := Match{
			Player1ID:       opponentID,
			Player1Username: opponentUsername,
			Player2ID:       &userID,
			Player2Username: username,
			BotDifficulty:   "", // PvP game, no difficulty needed
		}

		m.MatchChannel <- match
	}
	return nil
}

func (m *MatchmakingQueue) HandleTimeout(userID int64) {
	m.Mux.Lock()
	defer m.Mux.Unlock()

	_, exists := m.WaitingPlayers[userID]
	if !exists {
		return
	}

	delete(m.WaitingPlayers, userID)
	delete(m.Difficulties, userID)
	m.stopAndDeleteTimer(userID)
}

func (m *MatchmakingQueue) GetMatchChannel() chan Match {
	return m.MatchChannel
}

func (m *MatchmakingQueue) RemovePlayer(userID int64) {
	m.Mux.Lock()
	defer m.Mux.Unlock()

	delete(m.WaitingPlayers, userID)
	delete(m.Difficulties, userID)
	m.stopAndDeleteTimer(userID)
}

func (m *MatchmakingQueue) stopAndDeleteTimer(userID int64) {
	if timer := (*m.Timer)[userID]; timer != nil {
		timer.Stop()
	}
	delete(*m.Timer, userID)
}
