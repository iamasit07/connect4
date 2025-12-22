package models

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

type Match struct {
	Player1ID       int64
	Player1Username string
	Player2ID       *int64 // nil for BOT
	Player2Username string
}

type MatchmakingQueue struct {
	WaitingPlayers map[int64]string         // userID → username
	Mux            *sync.Mutex
	MatchChannel   chan Match
	Timer          *map[int64]*time.Timer
}

func NewMatchmakingQueue() *MatchmakingQueue {
	timerMap := make(map[int64]*time.Timer)
	waitingPlayers := make(map[int64]string) // userID → username
	queue := &MatchmakingQueue{
		WaitingPlayers: waitingPlayers,
		MatchChannel:   make(chan Match, 100),
		Mux:            &sync.Mutex{},
		Timer:          &timerMap,
	}
	return queue
}

func (m *MatchmakingQueue) AddPlayerToQueue(userID int64, username string) error {
	m.Mux.Lock()
	defer m.Mux.Unlock()

	if _, exists := m.WaitingPlayers[userID]; exists {
		return nil
	}

	if len(m.WaitingPlayers) == 0 {
		// First player in queue, start bot timer
		m.WaitingPlayers[userID] = username
		timer := time.AfterFunc(10*time.Second, func() {
			m.HandleTimeout(userID)
		})
		(*m.Timer)[userID] = timer
	} else {
		// Match with waiting player
		var opponentID int64
		var opponentUsername string
		for uid, name := range m.WaitingPlayers {
			opponentID = uid
			opponentUsername = name
			break
		}

		delete(m.WaitingPlayers, opponentID)
		m.stopAndDeleteTimer(opponentID)

		match := Match{
			Player1ID:       opponentID,
			Player1Username: opponentUsername,
			Player2ID:       &userID,
			Player2Username: username,
		}

		m.MatchChannel <- match
	}
	return nil
}

func (m *MatchmakingQueue) HandleTimeout(userID int64) {
	m.Mux.Lock()
	defer m.Mux.Unlock()

	username, exists := m.WaitingPlayers[userID]
	if !exists {
		return
	}

	delete(m.WaitingPlayers, userID)
	m.stopAndDeleteTimer(userID)

	match := Match{
		Player1ID:       userID,
		Player1Username: username,
		Player2ID:       nil, // BOT
		Player2Username: BotUsername,
	}

	m.MatchChannel <- match
}

func (m *MatchmakingQueue) GetMatchChannel() chan Match {
	return m.MatchChannel
}

func (m *MatchmakingQueue) RemovePlayer(userID int64) {
	m.Mux.Lock()
	defer m.Mux.Unlock()

	delete(m.WaitingPlayers, userID)
	m.stopAndDeleteTimer(userID)
}

func (m *MatchmakingQueue) stopAndDeleteTimer(userID int64) {
	if timer := (*m.Timer)[userID]; timer != nil {
		timer.Stop()
	}
	delete(*m.Timer, userID)
}

func GenerateGameID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}
