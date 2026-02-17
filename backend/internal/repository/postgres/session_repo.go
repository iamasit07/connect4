package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/internal/domain"
)

type SessionRepo struct {
	DB *sql.DB
}

func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{DB: db}
}

// CreateSession creates a new session in the database
func (r *SessionRepo) CreateSession(userID int64, sessionID, deviceInfo, ipAddress string, expiresAt time.Time) error {
	query := `
	INSERT INTO user_sessions (user_id, session_id, device_info, ip_address, expires_at)
	VALUES ($1, CAST($2 as TEXT), $3, $4, $5);
	`
	_, err := r.DB.Exec(query, userID, sessionID, deviceInfo, ipAddress, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	return nil
}

// GetSessionByID retrieves a session by session_id
func (r *SessionRepo) GetSessionByID(sessionID string) (*domain.UserSession, error) {
	query := `
	SELECT id, user_id, session_id, device_info, ip_address, created_at, expires_at, last_activity, is_active
	FROM user_sessions
	WHERE session_id = CAST($1 as TEXT);
	`
	var session domain.UserSession
	err := r.DB.QueryRow(query, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.SessionID,
		&session.DeviceInfo,
		&session.IPAddress,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.LastActivity,
		&session.IsActive,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %v", err)
	}
	return &session, nil
}

// GetActiveSessionByUserID retrieves the active session for a user
func (r *SessionRepo) GetActiveSessionByUserID(userID int64) (*domain.UserSession, error) {
	query := `
	SELECT id, user_id, session_id, device_info, ip_address, created_at, expires_at, last_activity, is_active
	FROM user_sessions
	WHERE user_id = $1 AND is_active = TRUE
	LIMIT 1;
	`
	var session domain.UserSession
	err := r.DB.QueryRow(query, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.SessionID,
		&session.DeviceInfo,
		&session.IPAddress,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.LastActivity,
		&session.IsActive,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active session: %v", err)
	}
	return &session, nil
}

// DeactivateAllUserSessions marks all sessions for a user as inactive
func (r *SessionRepo) DeactivateAllUserSessions(userID int64) error {
	query := `
	UPDATE user_sessions
	SET is_active = FALSE
	WHERE user_id = $1 AND is_active = TRUE;
	`
	_, err := r.DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user sessions: %v", err)
	}
	return nil
}

// DeactivateSession marks a specific session as inactive
func (r *SessionRepo) DeactivateSession(sessionID string) error {
	query := `
	UPDATE user_sessions
	SET is_active = FALSE
	WHERE session_id = CAST($1 as TEXT);
	`
	_, err := r.DB.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to deactivate session: %v", err)
	}
	return nil
}

// UpdateSessionActivity updates the last_activity timestamp
func (r *SessionRepo) UpdateSessionActivity(sessionID string) error {
	query := `
	UPDATE user_sessions
	SET last_activity = CURRENT_TIMESTAMP
	WHERE session_id = CAST($1 as TEXT);
	`
	_, err := r.DB.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session activity: %v", err)
	}
	return nil
}

// CleanupOldSessions deletes inactive sessions older than specified days
func (r *SessionRepo) CleanupOldSessions(olderThanDays int) (int64, error) {
	query := `
	DELETE FROM user_sessions
	WHERE is_active = FALSE 
	AND created_at < NOW() - INTERVAL '1 day' * $1;
	`
	result, err := r.DB.Exec(query, olderThanDays)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old sessions: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %v", err)
	}

	return rowsAffected, nil
}

// GetUserSessionHistory retrieves recent login sessions for a user
func (r *SessionRepo) GetUserSessionHistory(userID int64, limit int) ([]domain.UserSession, error) {
	query := `
	SELECT id, user_id, session_id, device_info, ip_address, created_at, expires_at, last_activity, is_active
	FROM user_sessions
	WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT $2;
	`
	rows, err := r.DB.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query session history: %v", err)
	}
	defer rows.Close()

	var sessions []domain.UserSession
	for rows.Next() {
		var s domain.UserSession
		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.SessionID,
			&s.DeviceInfo,
			&s.IPAddress,
			&s.CreatedAt,
			&s.ExpiresAt,
			&s.LastActivity,
			&s.IsActive,
		); err != nil {
			return nil, fmt.Errorf("failed to scan session row: %v", err)
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate session rows: %v", err)
	}

	return sessions, nil
}
