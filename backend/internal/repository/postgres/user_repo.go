package postgres

import (
	"database/sql"
	"fmt"
	"time"
)

type UserRepo struct {
	DB *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{DB: db}
}

type User struct {
	ID           int64
	Username     string
	Email        sql.NullString
	GoogleID     sql.NullString
	IsVerified   bool
	PasswordHash string
	GamesPlayed  int
	GamesWon     int
	CreatedAt    time.Time
}

type PlayerStats struct {
	Username    string  `json:"username"`
	GamesPlayed int     `json:"games_played"`
	GamesWon    int     `json:"games_won"`
	WinRate     float64 `json:"win_rate"`
}

// CreateUser creates a new user with hashed password and optional email/google_id
func (r *UserRepo) CreateUser(username, passwordHash string, email, googleID string) (int64, error) {
	// Handle empty strings as NULL
	var emailParam, googleIDParam interface{}
	emailParam = nil
	if email != "" {
		emailParam = email
	}
	googleIDParam = nil
	if googleID != "" {
		googleIDParam = googleID
	}

	query := `
	INSERT INTO players (username, password_hash, email, google_id, games_played, games_won)
	VALUES ($1, $2, $3, $4, 0, 0)
	RETURNING id;
	`
	var userID int64
	err := r.DB.QueryRow(query, username, passwordHash, emailParam, googleIDParam).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %v", err)
	}
	return userID, nil
}

// GetUserByUsername retrieves a user by username
func (r *UserRepo) GetUserByUsername(username string) (*User, error) {
	query := `
	SELECT id, username, email, google_id, is_verified, password_hash, games_played, games_won, created_at
	FROM players
	WHERE username = $1;
	`
	var user User
	err := r.DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.GoogleID,
		&user.IsVerified,
		&user.PasswordHash,
		&user.GamesPlayed,
		&user.GamesWon,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepo) GetUserByEmail(email string) (*User, error) {
	query := `
	SELECT id, username, email, google_id, is_verified, password_hash, games_played, games_won, created_at
	FROM players
	WHERE email = $1;
	`
	var user User
	err := r.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.GoogleID,
		&user.IsVerified,
		&user.PasswordHash,
		&user.GamesPlayed,
		&user.GamesWon,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return &user, nil
}

// GetUserByIdentifier retrieves a user by username OR email
func (r *UserRepo) GetUserByIdentifier(identifier string) (*User, error) {
	query := `
	SELECT id, username, email, google_id, is_verified, password_hash, games_played, games_won, created_at
	FROM players
	WHERE username = $1 OR email = $1;
	`
	var user User
	err := r.DB.QueryRow(query, identifier).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.GoogleID,
		&user.IsVerified,
		&user.PasswordHash,
		&user.GamesPlayed,
		&user.GamesWon,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return &user, nil
}

// UpdateUserGoogleID updates a user's Google ID based on their email
func (r *UserRepo) UpdateUserGoogleID(email, googleID string) error {
	query := `
	UPDATE players
	SET google_id = $2, is_verified = TRUE
	WHERE email = $1;
	`
	_, err := r.DB.Exec(query, email, googleID)
	if err != nil {
		return fmt.Errorf("failed to update google id: %v", err)
	}
	return nil
}

// GetUserByGoogleID retrieves a user by Google ID
func (r *UserRepo) GetUserByGoogleID(googleID string) (*User, error) {
	query := `
	SELECT id, username, email, google_id, is_verified, password_hash, games_played, games_won, created_at
	FROM players
	WHERE google_id = $1;
	`
	var user User
	err := r.DB.QueryRow(query, googleID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.GoogleID,
		&user.IsVerified,
		&user.PasswordHash,
		&user.GamesPlayed,
		&user.GamesWon,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepo) GetUserByID(userID int64) (*User, error) {
	query := `
	SELECT id, username, email, google_id, is_verified, password_hash, games_played, games_won, created_at
	FROM players
	WHERE id = $1;
	`
	var user User
	err := r.DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.GoogleID,
		&user.IsVerified,
		&user.PasswordHash,
		&user.GamesPlayed,
		&user.GamesWon,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return &user, nil
}

func (r *UserRepo) GetLeaderboard() ([]PlayerStats, error) {
	query := `
	SELECT username, games_played, games_won,
		CASE 
			WHEN games_played > 0 THEN ROUND((games_won::decimal / games_played) * 100, 2)
			ELSE 0
		END AS win_rate
	FROM players
	ORDER BY games_won DESC, username ASC;
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %v", err)
	}
	defer rows.Close()

	var leaderboard []PlayerStats
	for rows.Next() {
		var stats PlayerStats
		if err := rows.Scan(&stats.Username, &stats.GamesPlayed, &stats.GamesWon, &stats.WinRate); err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard row: %v", err)
		}
		leaderboard = append(leaderboard, stats)
	}

	return leaderboard, nil
}
