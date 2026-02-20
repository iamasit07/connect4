package cleanup

import (
	"log"
	"time"

	"github.com/iamasit07/connect4/backend/internal/repository/postgres"
	"github.com/iamasit07/connect4/backend/internal/service/game"
)

type Worker struct {
	SessionManager *game.SessionManager
	SessionRepository *postgres.SessionRepo
}

func NewWorker(sm *game.SessionManager, sr *postgres.SessionRepo) *Worker {
	return &Worker{SessionManager: sm, SessionRepository: sr}
}

// Start initiates the background ticker
func (w *Worker) Start() {
	go w.runCleanup()

	// Then run periodically (every 1 hour)
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			w.runCleanup()
		}
	}()
	log.Println("[CLEANUP] Background worker started")
}

// runCleanup executes the actual cleanup logic
func (w *Worker) runCleanup() {
	log.Println("[CLEANUP] Starting scheduled cleanup task...")

	w.SessionManager.CleanupOldSessions()

	daysToKeep := 30 // Delete sessions older than 30 days
	deletedCount, err := w.SessionRepository.CleanupOldSessions(daysToKeep)
	if err != nil {
		log.Printf("[CLEANUP] Error cleaning up DB sessions: %v", err)
	} else {
		if deletedCount > 0 {
			log.Printf("[CLEANUP] Removed %d expired sessions from database", deletedCount)
		}
	}
}