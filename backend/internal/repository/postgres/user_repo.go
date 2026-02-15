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
	Name         string
	AvatarURL    string
	Email        sql.NullString
	GoogleID     sql.NullString
	IsVerified   bool
	PasswordHash string
	GamesPlayed  int
	GamesWon     int
	GamesDrawn   int
	Rating       int
	CreatedAt    time.Time
}

type PlayerStats struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
}

// UserResponse returns a consistent JSON-friendly map of user data
func (u *User) UserResponse() map[string]interface{} {
	email := ""
	if u.Email.Valid {
		email = u.Email.String
	}
	return map[string]interface{}{
		"id":        u.ID,
		"username":  u.Username,
		"name":      u.Name,
		"avatar_url": u.AvatarURL,
		"email":     email,
		"rating":    u.Rating,
		"wins":      u.GamesWon,
		"losses":    u.GamesPlayed - u.GamesWon - u.GamesDrawn,
		"draws":     u.GamesDrawn,
	}
}

// CreateUser creates a new user with hashed password and optional email/google_id/avatar
func (r *UserRepo) CreateUser(username, name, passwordHash string, email, googleID, avatarURL string) (int64, error) {
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
	INSERT INTO players (username, name, password_hash, email, google_id, avatar_url, games_played, games_won, games_drawn, rating)
	VALUES ($1, $2, $3, $4, $5, $6, 0, 0, 0, 1000)
	RETURNING id;
	`
	var userID int64
	err := r.DB.QueryRow(query, username, name, passwordHash, emailParam, googleIDParam, avatarURL).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %v", err)
	}
	return userID, nil
}

// scanUser is a helper that scans a row into a User struct
func scanUser(row interface{ Scan(dest ...any) error }) (*User, error) {
	var user User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Name,
		&user.AvatarURL,
		&user.Email,
		&user.GoogleID,
		&user.IsVerified,
		&user.PasswordHash,
		&user.GamesPlayed,
		&user.GamesWon,
		&user.GamesDrawn,
		&user.Rating,
		&user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

const userSelectFields = `id, username, COALESCE(name, '') as name, COALESCE(avatar_url, '') as avatar_url, email, google_id, is_verified, password_hash, games_played, games_won, games_drawn, rating, created_at`

// GetUserByUsername retrieves a user by username
func (r *UserRepo) GetUserByUsername(username string) (*User, error) {
	query := `SELECT ` + userSelectFields + ` FROM players WHERE username = $1;`
	user, err := scanUser(r.DB.QueryRow(query, username))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepo) GetUserByEmail(email string) (*User, error) {
	query := `SELECT ` + userSelectFields + ` FROM players WHERE email = $1;`
	user, err := scanUser(r.DB.QueryRow(query, email))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return user, nil
}

// GetUserByIdentifier retrieves a user by username OR email
func (r *UserRepo) GetUserByIdentifier(identifier string) (*User, error) {
	query := `SELECT ` + userSelectFields + ` FROM players WHERE username = $1 OR email = $1;`
	user, err := scanUser(r.DB.QueryRow(query, identifier))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return user, nil
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
	query := `SELECT ` + userSelectFields + ` FROM players WHERE google_id = $1;`
	user, err := scanUser(r.DB.QueryRow(query, googleID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepo) GetUserByID(userID int64) (*User, error) {
	query := `SELECT ` + userSelectFields + ` FROM players WHERE id = $1;`
	user, err := scanUser(r.DB.QueryRow(query, userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	return user, nil
}

// UpdateProfile updates a user's display name
func (r *UserRepo) UpdateProfile(userID int64, name string) error {
	query := `UPDATE players SET name = $2 WHERE id = $1;`
	_, err := r.DB.Exec(query, userID, name)
	if err != nil {
		return fmt.Errorf("failed to update profile: %v", err)
	}
	return nil
}

// UpdateAvatar updates a user's avatar URL
func (r *UserRepo) UpdateAvatar(userID int64, avatarURL string) error {
	query := `UPDATE players SET avatar_url = $2 WHERE id = $1;`
	_, err := r.DB.Exec(query, userID, avatarURL)
	if err != nil {
		return fmt.Errorf("failed to update avatar: %v", err)
	}
	return nil
}

func (r *UserRepo) GetLeaderboard() ([]PlayerStats, error) {
	query := `
	SELECT 
		ROW_NUMBER() OVER (ORDER BY rating DESC, games_won DESC, username ASC) AS rank,
		username,
		rating,
		games_won,
		games_played - games_won - games_drawn AS losses
	FROM players
	ORDER BY rating DESC, games_won DESC, username ASC;
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %v", err)
	}
	defer rows.Close()

	leaderboard := make([]PlayerStats, 0)
	for rows.Next() {
		var stats PlayerStats
		if err := rows.Scan(&stats.Rank, &stats.Username, &stats.Rating, &stats.Wins, &stats.Losses); err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard row: %v", err)
		}
		leaderboard = append(leaderboard, stats)
	}

	return leaderboard, nil
}
