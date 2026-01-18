package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/internal/domain"
	"github.com/iamasit07/4-in-a-row/backend/pkg/auth"
)

const sessionKeyPrefix = "session:"
const sessionTTL = 30 * 24 * time.Hour // 30 days

type SessionRepository interface {
	CreateSession(userID int64, sessionID, deviceInfo, ipAddress string, expiresAt time.Time) error
	GetSessionByID(sessionID string) (*domain.UserSession, error)
	GetActiveSessionByUserID(userID int64) (*domain.UserSession, error)
	DeactivateAllUserSessions(userID int64) error
	DeactivateSession(sessionID string) error
	UpdateSessionActivity(sessionID string) error
	GetUserSessionHistory(userID int64, limit int) ([]domain.UserSession, error)
}

type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
}

// AuthService handles authentication session logic
type AuthService struct {
	repo  SessionRepository
	cache CacheRepository // Optional, can be nil
}

func NewAuthService(repo SessionRepository, cache CacheRepository) *AuthService {
	return &AuthService{
		repo:  repo,
		cache: cache,
	}
}

// SetSession stores session in database and cache
func (s *AuthService) SetSession(session *domain.UserSession) error {
	// Always store in database
	err := s.repo.CreateSession(
		session.UserID,
		session.SessionID,
		session.DeviceInfo,
		session.IPAddress,
		session.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to store session in database: %v", err)
	}

	// Try to store in cache if available
	if s.cache != nil {
		err = s.setSessionInCache(session)
		if err != nil {
			log.Printf("[SESSION] Warning: Failed to store session in cache: %v", err)
		}
	}

	return nil
}

// setSessionInCache stores session in cache with TTL
func (s *AuthService) setSessionInCache(session *domain.UserSession) error {
	ctx := context.Background()
	key := sessionKeyPrefix + session.SessionID

	sessionData, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return s.cache.Set(ctx, key, sessionData, sessionTTL)
}

// GetSession retrieves session from cache first, falls back to database
func (s *AuthService) GetSession(sessionID string) (*domain.UserSession, error) {
	// Try cache first
	if s.cache != nil {
		session, err := s.getSessionFromCache(sessionID)
		if err == nil && session != nil {
			return session, nil
		}
		if err != nil {
			// Just log, don't fail, fallback to DB
			// log.Printf("[SESSION] Cache lookup failed: %v", err)
		}
	}

	// Fall back to database
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	// If found in DB and cache is enabled, populate cache
	if session != nil && s.cache != nil {
		err = s.setSessionInCache(session)
		if err != nil {
			log.Printf("[SESSION] Warning: Failed to populate cache: %v", err)
		}
	}

	return session, nil
}

// GetActiveSession retrieves the currently active session for a user
func (s *AuthService) GetActiveSession(userID int64) (*domain.UserSession, error) {
	return s.repo.GetActiveSessionByUserID(userID)
}

// getSessionFromCache retrieves session from cache
func (s *AuthService) getSessionFromCache(sessionID string) (*domain.UserSession, error) {
	ctx := context.Background()
	key := sessionKeyPrefix + sessionID

	data, err := s.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var session domain.UserSession
	err = json.Unmarshal([]byte(data), &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// InvalidateSession marks session as inactive
func (s *AuthService) InvalidateSession(sessionID string) error {
	// Deactivate in database
	err := s.repo.DeactivateSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to deactivate session in database: %v", err)
	}

	// Remove from cache
	if s.cache != nil {
		ctx := context.Background()
		key := sessionKeyPrefix + sessionID
		err = s.cache.Del(ctx, key)
		if err != nil {
			log.Printf("[SESSION] Warning: Failed to delete session from cache: %v", err)
		}
	}

	return nil
}

// InvalidateAllUserSessions deactivates all sessions for a user
func (s *AuthService) InvalidateAllUserSessions(userID int64) error {
	// Get active session for user to invalidate cache
	activeSession, err := s.repo.GetActiveSessionByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get active session: %v", err)
	}

	// Deactivate all in database
	err = s.repo.DeactivateAllUserSessions(userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user sessions in database: %v", err)
	}

	// Remove active session from cache
	if activeSession != nil && s.cache != nil {
		ctx := context.Background()
		key := sessionKeyPrefix + activeSession.SessionID
		err = s.cache.Del(ctx, key)
		if err != nil {
			log.Printf("[SESSION] Warning: Failed to delete session from cache: %v", err)
		}
	}

	return nil
}

// UpdateSessionActivity updates the last activity timestamp
func (s *AuthService) UpdateSessionActivity(sessionID string) error {
	// Update in database
	err := s.repo.UpdateSessionActivity(sessionID)
	if err != nil {
		return err
	}

	// Update in cache (reset TTL) if enabled
	if s.cache != nil {
		session, err := s.repo.GetSessionByID(sessionID)
		if err == nil && session != nil {
			err = s.setSessionInCache(session)
			if err != nil {
				log.Printf("[SESSION] Warning: Failed to update session in cache: %v", err)
			}
		}
	}

	return nil
}

// GetUserSessionHistory retrieves session history for a user
func (s *AuthService) GetUserSessionHistory(userID int64, limit int) ([]domain.UserSession, error) {
	return s.repo.GetUserSessionHistory(userID, limit)
}

// ValidateToken validates a JWT token and checks if the session is active
func (s *AuthService) ValidateToken(tokenString string) (*auth.Claims, error) {
	// First validate the JWT itself
	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		return nil, err
	}

	// Check if session is still active
	session, err := s.GetSession(claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session validation failed: %v", err)
	}

	if session == nil || !session.IsActive {
		return nil, errors.New("session invalidated")
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session expired")
	}

	return claims, nil
}
