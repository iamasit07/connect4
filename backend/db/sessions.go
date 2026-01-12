package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/models"
)

// CreateSession creates a new session in the database
func CreateSession(userID int64, sessionID, deviceInfo, ipAddress string, expiresAt time.Time) error {
	query := `
	INSERT INTO user_sessions (user_id, session_id, device_info, ip_address, expires_at)
	VALUES ($1, $2, $3, $4, $5);
	`
	_, err := DB.Exec(query, userID, sessionID, deviceInfo, ipAddress, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	return nil
}

// GetSessionByID retrieves a session by session_id
func GetSessionByID(sessionID string) (*models.UserSession, error) {
	query := `
	SELECT id, user_id, session_id, device_info, ip_address, created_at, expires_at, last_activity, is_active
	FROM user_sessions
	WHERE session_id = $1;
	`
	var session models.UserSession
	err := DB.QueryRow(query, sessionID).Scan(
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
func GetActiveSessionByUserID(userID int64) (*models.UserSession, error) {
	query := `
	SELECT id, user_id, session_id, device_info, ip_address, created_at, expires_at, last_activity, is_active
	FROM user_sessions
	WHERE user_id = $1 AND is_active = TRUE
	LIMIT 1;
	`
	var session models.UserSession
	err := DB.QueryRow(query, userID).Scan(
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
func DeactivateAllUserSessions(userID int64) error {
	query := `
	UPDATE user_sessions
	SET is_active = FALSE
	WHERE user_id = $1 AND is_active = TRUE;
	`
	_, err := DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user sessions: %v", err)
	}
	return nil
}

// DeactivateSession marks a specific session as inactive
func DeactivateSession(sessionID string) error {
	query := `
	UPDATE user_sessions
	SET is_active = FALSE
	WHERE session_id = $1;
	`
	_, err := DB.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to deactivate session: %v", err)
	}
	return nil
}

// UpdateSessionActivity updates the last_activity timestamp
func UpdateSessionActivity(sessionID string) error {
	query := `
	UPDATE user_sessions
	SET last_activity = CURRENT_TIMESTAMP
	WHERE session_id = $1;
	`
	_, err := DB.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session activity: %v", err)
	}
	return nil
}

// CleanupOldSessions deletes inactive sessions older than specified days
func CleanupOldSessions(olderThanDays int) (int64, error) {
	query := `
	DELETE FROM user_sessions
	WHERE is_active = FALSE 
	AND created_at < NOW() - INTERVAL '1 day' * $1;
	`
	result, err := DB.Exec(query, olderThanDays)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old sessions: %v", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %v", err)
	}
	
	return rowsAffected, nil
}

// GetUserSessionHistory retrieves session history for a user (limited)
func GetUserSessionHistory(userID int64, limit int) ([]models.UserSession, error) {
	query := `
	SELECT id, user_id, session_id, device_info, ip_address, created_at, expires_at, last_activity, is_active
	FROM user_sessions
	WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT $2;
	`
	
	rows, err := DB.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query session history: %v", err)
	}
	defer rows.Close()
	
	var sessions []models.UserSession
	for rows.Next() {
		var session models.UserSession
		err := rows.Scan(
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
		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %v", err)
		}
		sessions = append(sessions, session)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating session rows: %v", err)
	}
	
	return sessions, nil
}
