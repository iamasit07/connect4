package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/iamasit07/connect4/backend/internal/config"
	"github.com/iamasit07/connect4/backend/internal/domain"
	"github.com/iamasit07/connect4/backend/pkg/auth"
)

const sessionKeyPrefix = "session:"
const refreshTokenKeyPrefix = "refresh_token:"
const blockedSessionKeyPrefix = "blocked_session:"
const sessionTTL = 30 * 24 * time.Hour // 30 days

type SessionRepository interface {
	CreateSession(userID int64, sessionID, deviceInfo, ipAddress string, expiresAt time.Time) error
	GetSessionByID(sessionID string) (*domain.UserSession, error)
	GetActiveSessionByUserID(userID int64) (*domain.UserSession, error)
	DeactivateAllUserSessions(userID int64) error
	DeactivateSession(sessionID string) error
	UpdateSessionActivity(sessionID string) error
	GetUserSessionHistory(userID int64, limit int) ([]domain.UserSession, error)
	// Refresh token methods
	StoreRefreshToken(tokenID string, userID int64, sessionID string, expiresAt time.Time) error
	GetRefreshToken(tokenID string) (*domain.RefreshToken, error)
	RevokeRefreshToken(tokenID string) error
	RevokeAllUserRefreshTokens(userID int64) error
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

// BlocklistSession adds a session ID to the Redis blocklist with a TTL.
func (s *AuthService) BlocklistSession(sessionID string, ttl time.Duration) error {
	if s.cache == nil {
		return nil
	}
	ctx := context.Background()
	key := blockedSessionKeyPrefix + sessionID
	return s.cache.Set(ctx, key, "1", ttl)
}

// IsSessionBlocked checks if a session ID is in the blocklist.
func (s *AuthService) IsSessionBlocked(sessionID string) bool {
	if s.cache == nil {
		return false
	}
	ctx := context.Background()
	key := blockedSessionKeyPrefix + sessionID
	val, err := s.cache.Get(ctx, key)
	return err == nil && val != ""
}

// ValidateTokenOffline validates a JWT using only signature + blocklist.
func (s *AuthService) ValidateTokenOffline(tokenString string) (*auth.Claims, error) {
	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		return nil, err
	}
	if s.IsSessionBlocked(claims.SessionID) {
		return nil, errors.New("session is blocked/revoked")
	}
	return claims, nil
}

func (s *AuthService) GenerateTokenPair(userID int64, username, sessionID string) (accessToken, refreshToken string, err error) {
	accessToken, err = auth.GenerateAccessToken(userID, username, sessionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %v", err)
	}

	tokenID := auth.GenerateToken()

	refreshToken, err = auth.GenerateRefreshToken(userID, sessionID, tokenID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %v", err)
	}

	refreshTTL := time.Duration(config.AppConfig.RefreshTokenTTLDays) * 24 * time.Hour
	expiresAt := time.Now().Add(refreshTTL)

	if err := s.repo.StoreRefreshToken(tokenID, userID, sessionID, expiresAt); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %v", err)
	}

	if s.cache != nil {
		ctx := context.Background()
		key := refreshTokenKeyPrefix + tokenID
		rtData := &domain.RefreshToken{
			TokenID:   tokenID,
			UserID:    userID,
			SessionID: sessionID,
			ExpiresAt: expiresAt,
			CreatedAt: time.Now(),
			Revoked:   false,
		}
		data, err := json.Marshal(rtData)
		if err == nil {
			if cacheErr := s.cache.Set(ctx, key, data, refreshTTL); cacheErr != nil {
				log.Printf("[SESSION] Warning: Failed to cache refresh token: %v", cacheErr)
			}
		}
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) ValidateAndRefresh(refreshTokenString string) (newAccessToken, newRefreshToken string, err error) {
	claims, err := auth.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %v", err)
	}

	rt, err := s.getRefreshTokenMetadata(claims.TokenID)
	if err != nil {
		return "", "", fmt.Errorf("refresh token lookup failed: %v", err)
	}
	if rt == nil {
		return "", "", errors.New("refresh token not found")
	}
	if rt.Revoked {
		return "", "", errors.New("refresh token has been revoked")
	}
	if time.Now().After(rt.ExpiresAt) {
		return "", "", errors.New("refresh token has expired")
	}

	session, err := s.GetSession(claims.SessionID)
	if err != nil {
		return "", "", fmt.Errorf("session validation failed: %v", err)
	}
	if session == nil || !session.IsActive {
		return "", "", errors.New("session invalidated")
	}

	if err := s.RevokeRefreshToken(claims.TokenID); err != nil {
		log.Printf("[SESSION] Warning: Failed to revoke old refresh token: %v", err)
	}

	newAccessToken, newRefreshToken, err = s.GenerateTokenPair(rt.UserID, "", claims.SessionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new token pair: %v", err)
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *AuthService) ValidateAndRefreshWithUsername(refreshTokenString string, getUsername func(int64) (string, error)) (newAccessToken, newRefreshToken string, userID int64, err error) {
	claims, err := auth.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid refresh token: %v", err)
	}

	rt, err := s.getRefreshTokenMetadata(claims.TokenID)
	if err != nil {
		return "", "", 0, fmt.Errorf("refresh token lookup failed: %v", err)
	}
	if rt == nil {
		return "", "", 0, errors.New("refresh token not found")
	}
	if rt.Revoked {
		return "", "", 0, errors.New("refresh token has been revoked")
	}
	if time.Now().After(rt.ExpiresAt) {
		return "", "", 0, errors.New("refresh token has expired")
	}

	session, err := s.GetSession(claims.SessionID)
	if err != nil {
		return "", "", 0, fmt.Errorf("session validation failed: %v", err)
	}
	if session == nil || !session.IsActive {
		return "", "", 0, errors.New("session invalidated")
	}

	username, err := getUsername(rt.UserID)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to get username: %v", err)
	}

	if err := s.RevokeRefreshToken(claims.TokenID); err != nil {
		log.Printf("[SESSION] Warning: Failed to revoke old refresh token: %v", err)
	}

	newAccessToken, newRefreshToken, err = s.GenerateTokenPair(rt.UserID, username, claims.SessionID)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to generate new token pair: %v", err)
	}

	return newAccessToken, newRefreshToken, rt.UserID, nil
}

func (s *AuthService) getRefreshTokenMetadata(tokenID string) (*domain.RefreshToken, error) {
	if s.cache != nil {
		ctx := context.Background()
		key := refreshTokenKeyPrefix + tokenID
		data, err := s.cache.Get(ctx, key)
		if err == nil && data != "" {
			var rt domain.RefreshToken
			if err := json.Unmarshal([]byte(data), &rt); err == nil {
				return &rt, nil
			}
		}
	}
	return s.repo.GetRefreshToken(tokenID)
}

func (s *AuthService) RevokeRefreshToken(tokenID string) error {
	if err := s.repo.RevokeRefreshToken(tokenID); err != nil {
		return err
	}
	if s.cache != nil {
		ctx := context.Background()
		key := refreshTokenKeyPrefix + tokenID
		if err := s.cache.Del(ctx, key); err != nil {
			log.Printf("[SESSION] Warning: Failed to delete refresh token from cache: %v", err)
		}
	}
	return nil
}

func (s *AuthService) RevokeAllUserRefreshTokens(userID int64) error {
	return s.repo.RevokeAllUserRefreshTokens(userID)
}

func (s *AuthService) SetSession(session *domain.UserSession) error {
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
	if s.cache != nil {
		err = s.setSessionInCache(session)
		if err != nil {
			log.Printf("[SESSION] Warning: Failed to store session in cache: %v", err)
		}
	}
	return nil
}

func (s *AuthService) setSessionInCache(session *domain.UserSession) error {
	ctx := context.Background()
	key := sessionKeyPrefix + session.SessionID
	sessionData, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, key, sessionData, sessionTTL)
}

func (s *AuthService) GetSession(sessionID string) (*domain.UserSession, error) {
	if s.cache != nil {
		session, err := s.getSessionFromCache(sessionID)
		if err == nil && session != nil {
			return session, nil
		}
	}
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return nil, err
	}
	if session != nil && s.cache != nil {
		err = s.setSessionInCache(session)
		if err != nil {
			log.Printf("[SESSION] Warning: Failed to populate cache: %v", err)
		}
	}
	return session, nil
}

func (s *AuthService) GetActiveSession(userID int64) (*domain.UserSession, error) {
	return s.repo.GetActiveSessionByUserID(userID)
}

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

// InvalidateSession marks session as inactive AND adds to blocklist
func (s *AuthService) InvalidateSession(sessionID string) error {
	err := s.repo.DeactivateSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to deactivate session in database: %v", err)
	}
	if s.cache != nil {
		ctx := context.Background()
		key := sessionKeyPrefix + sessionID
		s.cache.Del(ctx, key)
	}
	return s.BlocklistSession(sessionID, 1*time.Hour)
}

// InvalidateAllUserSessions deactivates all sessions and blocklists the active one
func (s *AuthService) InvalidateAllUserSessions(userID int64) error {
	activeSession, _ := s.repo.GetActiveSessionByUserID(userID)
	
	err := s.repo.DeactivateAllUserSessions(userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user sessions in database: %v", err)
	}

	if activeSession != nil && activeSession.IsActive {
		s.BlocklistSession(activeSession.SessionID, 1*time.Hour)
		if s.cache != nil {
			ctx := context.Background()
			key := sessionKeyPrefix + activeSession.SessionID
			s.cache.Del(ctx, key)
		}
	}
	return nil
}

func (s *AuthService) UpdateSessionActivity(sessionID string) error {
	err := s.repo.UpdateSessionActivity(sessionID)
	if err != nil {
		return err
	}
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

func (s *AuthService) GetUserSessionHistory(userID int64, limit int) ([]domain.UserSession, error) {
	return s.repo.GetUserSessionHistory(userID, limit)
}

func (s *AuthService) ValidateToken(tokenString string) (*auth.Claims, error) {
	claims, err := auth.ValidateJWT(tokenString)
	if err != nil {
		return nil, err
	}
	session, err := s.GetSession(claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("session validation failed: %v", err)
	}
	if session == nil || !session.IsActive {
		return nil, errors.New("session invalidated")
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("session expired")
	}
	return claims, nil
}
