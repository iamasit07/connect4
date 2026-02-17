package matchmaking

import (
	"log"

	"github.com/iamasit07/4-in-a-row/backend/internal/service/game"
)

func MatchMakingListener(queue *MatchmakingQueue, cm game.ConnectionManagerInterface, sm *game.SessionManager) {
	for {
		match := <-queue.MatchChannel

		player1ID := match.Player1ID
		player1Username := match.Player1Username
		player2ID := match.Player2ID
		player2Username := match.Player2Username

		// CreateSession will handle sending game_start messages
		session := sm.CreateSession(player1ID, player1Username, player2ID, player2Username, match.BotDifficulty, cm)

		log.Printf("[MATCHMAKING] Match started: %s vs %s (game: %s)",
			player1Username, player2Username, session.GameID)
	}
}