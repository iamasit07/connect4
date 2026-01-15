package matchmaking

import (
	"log"

	"github.com/iamasit07/4-in-a-row/backend/internal/domain"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/game"
)

func MatchMakingListener(queue *MatchmakingQueue, cm game.ConnectionManagerInterface, sm *game.SessionManager) {
	for {
		match := <-queue.MatchChannel

		player1ID := match.Player1ID
		player1Username := match.Player1Username
		player2ID := match.Player2ID
		player2Username := match.Player2Username

		log.Printf("[MATCHMAKING] Match found: %s (ID: %d) vs %s (ID: %v)",
			player1Username, player1ID, player2Username, player2ID)

		// Terminate any existing sessions for these users to prevent jagged state
		sm.TerminateSessionForUser(player1ID, cm)
		
		if player2Username != domain.BotUsername && player2ID != nil {
			sm.TerminateSessionForUser(*player2ID, cm)
		}

		// CreateSession will handle sending game_start messages
		session := sm.CreateSession(player1ID, player1Username, player2ID, player2Username, match.BotDifficulty, cm)

		log.Printf("Match started between %s (ID: %d) and %s (ID: %v) with game ID %s\n",
			player1Username, player1ID, player2Username, player2ID, session.GameID)
	}
}