package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/db"
	"github.com/iamasit07/4-in-a-row/backend/models"
)

const sessionKeyPrefix = "session:"
const sessionTTL = 30 * 24 * time.Hour // 30 days

// SetSession stores session in Redis (if available) and PostgreSQL
func SetSession(session *models.UserSession) error {
	// Always store in PostgreSQL
	err := db.CreateSession(
		session.UserID,
		session.SessionID,
		session.DeviceInfo,
		session.IPAddress,
		session.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store session in database: %v", err)
	}

	// Try to store in Redis if enabled
	if db.IsRedisEnabled() {
		err = setSessionInRedis(session)
		if err != nil {
			log.Printf("[SESSION] Warning: Failed to store session in Redis: %v", err)
			// Don't fail - PostgreSQL is our source of truth
		}
	}

	return nil
}

// setSessionInRedis stores session in Redis with TTL
func setSessionInRedis(session *models.UserSession) error {
	ctx := context.Background()
	key := sessionKeyPrefix + session.SessionID

	sessionData, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return db.RedisClient.Set(ctx, key, sessionData, sessionTTL).Err()
}

// GetSession retrieves session from Redis first, falls back to PostgreSQL
func GetSession(sessionID string) (*models.UserSession, error) {
	// Try Redis first if enabled
	if db.IsRedisEnabled() {
		session, err := getSessionFromRedis(sessionID)
		if err == nil && session != nil {
			return session, nil
		}
		// Redis miss or error - fall back to PostgreSQL
		if err != nil {
			log.Printf("[SESSION] Redis lookup failed, falling back to PostgreSQL: %v", err)
		}
	}

	// Fall back to PostgreSQL
	session, err := db.GetSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	// If found in PostgreSQL and Redis is enabled, cache it
	if session != nil && db.IsRedisEnabled() {
		err = setSessionInRedis(session)
		if err != nil {
			log.Printf("[SESSION] Warning: Failed to cache session in Redis: %v", err)
		}
	}

	return session, nil
}

// getSessionFromRedis retrieves session from Redis
func getSessionFromRedis(sessionID string) (*models.UserSession, error) {
	ctx := context.Background()
	key := sessionKeyPrefix + sessionID

	data, err := db.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var session models.UserSession
	err = json.Unmarshal([]byte(data), &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// InvalidateSession marks session as inactive in both Redis and PostgreSQL
func InvalidateSession(sessionID string) error {
	// Deactivate in PostgreSQL
	err := db.DeactivateSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to deactivate session in database: %v", err)
	}

	// Remove from Redis if enabled
	if db.IsRedisEnabled() {
		ctx := context.Background()
		key := sessionKeyPrefix + sessionID
		err = db.RedisClient.Del(ctx, key).Err()
		if err != nil {
			log.Printf("[SESSION] Warning: Failed to delete session from Redis: %v", err)
		}
	}

	return nil
}

// InvalidateAllUserSessions deactivates all sessions for a user
func InvalidateAllUserSessions(userID int64) error {
	// Get all active sessions for user from PostgreSQL
	activeSession, err := db.GetActiveSessionByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get active session: %v", err)
	}

	// Deactivate in PostgreSQL
	err = db.DeactivateAllUserSessions(userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user sessions in database: %v", err)
	}

	// Remove from Redis if found and Redis is enabled
	if activeSession != nil && db.IsRedisEnabled() {
		ctx := context.Background()
		key := sessionKeyPrefix + activeSession.SessionID
		err = db.RedisClient.Del(ctx, key).Err()
		if err != nil {
			log.Printf("[SESSION] Warning: Failed to delete session from Redis: %v", err)
		}
	}

	return nil
}

// UpdateSessionActivity updates the last activity timestamp
func UpdateSessionActivity(sessionID string) error {
	// Update in PostgreSQL
	err := db.UpdateSessionActivity(sessionID)
	if err != nil {
		return err
	}

	// Update in Redis if enabled (reset TTL)
	if db.IsRedisEnabled() {
		session, err := db.GetSessionByID(sessionID)
		if err == nil && session != nil {
			err = setSessionInRedis(session)
			if err != nil {
				log.Printf("[SESSION] Warning: Failed to update session in Redis: %v", err)
			}
		}
	}

	return nil
}
