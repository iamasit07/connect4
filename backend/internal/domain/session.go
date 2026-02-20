package domain

import "time"

type UserSession struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	SessionID    string    `json:"session_id"`
	DeviceInfo   string    `json:"device_info"`
	IPAddress    string    `json:"ip_address"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	LastActivity time.Time `json:"last_activity"`
	IsActive     bool      `json:"is_active"`
}

// RefreshToken represents a stored refresh token for token rotation
type RefreshToken struct {
	ID        int64     `json:"id"`
	TokenID   string    `json:"token_id"`
	UserID    int64     `json:"user_id"`
	SessionID string    `json:"session_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Revoked   bool      `json:"revoked"`
}
