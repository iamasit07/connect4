package jobs

import (
	"log"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/db"
)

// StartSessionCleanupCron starts a daily cron job to cleanup old sessions
func StartSessionCleanupCron() {
	// Run cleanup immediately on startup
	go runCleanup()

	// Then run every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for range ticker.C {
			runCleanup()
		}
	}()

	log.Println("[CLEANUP] Session cleanup cron job started (runs daily)")
}

func runCleanup() {
	log.Println("[CLEANUP] Running session cleanup...")
	count, err := db.CleanupOldSessions(90) // Delete sessions older than 90 days
	if err != nil {
		log.Printf("[CLEANUP] Error cleaning up sessions: %v", err)
	} else {
		log.Printf("[CLEANUP] Successfully deleted %d old sessions", count)
	}
}
