package models

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

type Match struct {
	Player1Token    string
	Player1Username string
	Player2Token    string
	Player2Username string
	GameID          string
}

type MatchmakingQueue struct {
	WaitingPlayers map[string]string  // token → username
	Mux            *sync.Mutex
	MatchChannel   chan Match
	Timer          *map[string]*time.Timer
}

func (m *MatchmakingQueue) NewMatchmakingQueue() *MatchmakingQueue {
	timerMap := make(map[string]*time.Timer)
	waitingPlayers := make(map[string]string)  // token → username
	queue := &MatchmakingQueue{
		WaitingPlayers: waitingPlayers,
		MatchChannel:   make(chan Match, 100),
		Mux:            &sync.Mutex{},
		Timer:          &timerMap,
	}
	return queue
}

func (m *MatchmakingQueue) AddPlayerToQueue(userToken, username string) error {
	m.Mux.Lock()
	defer m.Mux.Unlock()

	if _, exists := m.WaitingPlayers[userToken]; exists {
		return nil
	}

	if len(m.WaitingPlayers) == 0 {
		m.WaitingPlayers[userToken] = username
		timer := time.AfterFunc(10*time.Second, func() {
			m.HandleTimeout(userToken)
		})
		(*m.Timer)[userToken] = timer
	} else {
		var opponentToken, opponentUsername string
		for token, name := range m.WaitingPlayers {
			opponentToken = token
			opponentUsername = name
			break
		}

		delete(m.WaitingPlayers, opponentToken)
		m.stopAndDeleteTimer(opponentToken)

		match := Match{
			Player1Token:    opponentToken,
			Player1Username: opponentUsername,
			Player2Token:    userToken,
			Player2Username: username,
			GameID:          GenerateGameID(),
		}

		m.MatchChannel <- match
	}
	return nil
}

func (m *MatchmakingQueue) HandleTimeout(userToken string) {
	m.Mux.Lock()
	defer m.Mux.Unlock()

	username, exists := m.WaitingPlayers[userToken]
	if !exists {
		return
	}

	delete(m.WaitingPlayers, userToken)
	m.stopAndDeleteTimer(userToken)

	match := Match{
		Player1Token:    userToken,
		Player1Username: username,
		Player2Token:    "", // Will be set to BOT_TOKEN by caller
		Player2Username: BotUsername,
		GameID:          GenerateGameID(),
	}

	m.MatchChannel <- match
}

func (m *MatchmakingQueue) GetMatchChannel() chan Match {
	return m.MatchChannel
}

func (m *MatchmakingQueue) RemovePlayer(userToken string) {
	m.Mux.Lock()
	defer m.Mux.Unlock()

	delete(m.WaitingPlayers, userToken)
	m.stopAndDeleteTimer(userToken)
}

func (m *MatchmakingQueue) stopAndDeleteTimer(userToken string) {
	if timer := (*m.Timer)[userToken]; timer != nil {
		timer.Stop()
	}
	delete(*m.Timer, userToken)
}

func GenerateGameID() string {
	// Use crypto/rand for true randomness to prevent collisions
	// This generates a 16-character hexadecimal string
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp if rand fails (very unlikely)
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}
