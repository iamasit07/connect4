package matchmaking

import (
	"log"

	"github.com/iamasit07/connect4/backend/internal/service/game"
)

func MatchMakingListener(queue *MatchmakingQueue, sm *game.SessionManager) {
	for {
		match := <-queue.MatchChannel

		player1ID := match.Player1ID
		player1Username := match.Player1Username
		player2ID := match.Player2ID
		player2Username := match.Player2Username

		session := sm.CreateSession(player1ID, player1Username, player2ID, player2Username, match.BotDifficulty)

		log.Printf("[MATCHMAKING] Match started: %s vs %s (game: %s)",
			player1Username, player2Username, session.GameID)
	}
}