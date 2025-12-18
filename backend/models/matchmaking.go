package models

import (
	"fmt"
	"sync"
	"time"
)

type Match struct {
	Player1 string
	Player2 string
	GameID  string
}

type MatchmakingQueue struct {
	WaitingPlayers []string
	Mux            *sync.Mutex
	MatchChannel   chan Match
	Timer          *map[string]*time.Timer
}

func (m *MatchmakingQueue) NewMatchmakingQueue() *MatchmakingQueue {
	timerMap := make(map[string]*time.Timer)
	queue := &MatchmakingQueue{
		WaitingPlayers: []string{},
		MatchChannel:   make(chan Match, 100),
		Mux:            &sync.Mutex{},
		Timer:          &timerMap,
	}
	return queue
}

func (m *MatchmakingQueue) AddPlayerToQueue(username string) error {
	m.Mux.Lock()
	defer m.Mux.Unlock()
	
	for _, player := range m.WaitingPlayers {
		if player == username {
			return nil
		}
	}

	if len(m.WaitingPlayers) == 0 {
		m.WaitingPlayers = append(m.WaitingPlayers, username)
		timer := time.AfterFunc(10*time.Second, func() {
			m.HandleTimeout(username)
		})
		(*m.Timer)[username] = timer
	} else {
		opponent := m.WaitingPlayers[0]
		m.WaitingPlayers = m.WaitingPlayers[1:]

		timer1 := (*m.Timer)[opponent]
		timer2 := (*m.Timer)[username]
		if timer1 != nil {
			timer1.Stop()
		}
		if timer2 != nil {
			timer2.Stop()
		}

		delete(*m.Timer, opponent)
		delete(*m.Timer, username)

		match := Match{
			Player1: opponent,
			Player2: username,
			GameID:  GenerateGameID(),
		}

		m.MatchChannel <- match
	}
	return nil
}

func (m *MatchmakingQueue) HandleTimeout(username string) {
	m.Mux.Lock()
	defer m.Mux.Unlock()

	for i, player := range m.WaitingPlayers {
		if player == username {
			m.WaitingPlayers = append(m.WaitingPlayers[:i], m.WaitingPlayers[i+1:]...)
			break
		}
	}

	timer := (*m.Timer)[username]
	if timer != nil {
		timer.Stop()
	}
	delete(*m.Timer, username)

	match := Match{
		Player1: username,
		Player2: BotUsername,
		GameID:  GenerateGameID(),
	}

	m.MatchChannel <- match
}

func (m *MatchmakingQueue) GetMatchChannel() chan Match {
	return m.MatchChannel
}

func (m *MatchmakingQueue) RemovePlayer(username string) {
	m.Mux.Lock()
	defer m.Mux.Unlock()

	for i, player := range m.WaitingPlayers {
		if player == username {
			m.WaitingPlayers = append(m.WaitingPlayers[:i], m.WaitingPlayers[i+1:]...)
			break
		}

		timer := (*m.Timer)[username]
		if timer != nil {
			timer.Stop()
		}
		delete(*m.Timer, username)
	}
}

func GenerateGameID() string {
	id := time.Now().UnixNano()
	return fmt.Sprintf("%d", id)
}