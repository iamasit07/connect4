package models

import (
	"sync"
	"time"
)

type Match struct {
	Player1   string
	Player2   string
	GameID    string
}

type MatchmakingQueue struct {
	waitingPlayers []string
	mu             sync.Mutex
	matchChannel   chan Match
	timer 		*map[string]*time.Timer
}

func NewMatchmakingQueue() *MatchmakingQueue {
	queue := &MatchmakingQueue{
		waitingPlayers: []string{},
		matchChannel:   make(chan Match, 100),
		timer:          &map[string]*time.Timer{},
	}
	return queue
}

func AddPlayer(username string, queue *MatchmakingQueue) error {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	
	for _, player := range queue.waitingPlayers {
		if player == username {
			return nil
		}
	}

	if len(queue.waitingPlayers) == 0 {
		queue.waitingPlayers = append(queue.waitingPlayers, username)
		time := time.AfterFunc(10 * time.Second, func(){
			HandleTimeout(username, queue)
		})
		(*queue.timer)[username] = time
	} else {
		opponent := queue.waitingPlayers[0]
		queue.waitingPlayers = queue.waitingPlayers[1:]

		timer1 := (*queue.timer)[opponent]
		timer2 := (*queue.timer)[username]
		if timer1 != nil {
			timer1.Stop()
		}
		if timer2 != nil {
			timer2.Stop()
		}

		delete(*queue.timer, opponent)
		delete(*queue.timer, username)

		match := Match{
			Player1: opponent,
			Player2: username,
			GameID:  GenerateGameID(),
		}

		queue.matchChannel <- match
	}
}

func HandleTimeout(username string, queue *MatchmakingQueue) {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	for i, player := range queue.waitingPlayers {
		if player == username {
			queue.waitingPlayers = append(queue.waitingPlayers[:i], queue.waitingPlayers[i+1:]...)
			break
		}
	}

	timer := (*queue.timer)[username]
	if timer != nil {
		timer.Stop()
	}
	delete(*queue.timer, username)

	match := Match{
		Player1: username,
		Player2: "BOT",
		GameID:  GenerateGameID(),
	}

	queue.matchChannel <- match
}

func GetMatchChannel(queue *MatchmakingQueue) chan Match {
	return queue.matchChannel
}

func RemovePlayer(username string, queue *MatchmakingQueue) {
	queue.mu.Lock()
	defer queue.mu.Unlock()

	for i, player := range queue.waitingPlayers {
		if player == username {
			queue.waitingPlayers = append(queue.waitingPlayers[:i], queue.waitingPlayers[i+1:]...)
			break
		}

		timer := (*queue.timer)[username]
		if timer != nil {
			timer.Stop()
		}
		delete(*queue.timer, username)
	}
}