package db

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	ID           int64
	Username     string
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

// CreateUser creates a new user with hashed password
func CreateUser(username, passwordHash string) (int64, error) {
	query := `
	INSERT INTO players (username, password_hash, games_played, games_won)
	VALUES ($1, $2, 0, 0)
	RETURNING id;
	`
	var userID int64
	err := DB.QueryRow(query, username, passwordHash).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %v", err)
	}
	return userID, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (*User, error) {
	query := `
	SELECT id, username, password_hash, games_played, games_won, created_at
	FROM players
	WHERE username = $1;
	`
	var user User
	err := DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
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
func GetUserByID(userID int64) (*User, error) {
	query := `
	SELECT id, username, password_hash, games_played, games_won, created_at
	FROM players
	WHERE id = $1;
	`
	var user User
	err := DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
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

// UpdatePlayerStats updates player stats after a game
func UpdatePlayerStats(userID int64, won bool) error {
	query := `
	UPDATE players
	SET games_played = games_played + 1,
	    games_won = games_won + CASE WHEN $2 THEN 1 ELSE 0 END
	WHERE id = $1;
	`
	_, err := DB.Exec(query, userID, won)
	if err != nil {
		return fmt.Errorf("failed to update player stats: %v", err)
	}
	return nil
}

// SaveGame saves a finished game and updates player stats
func SaveGame(gameID string, player1ID int64, player1Username string, player2ID *int64, player2Username string, winnerID *int64, winnerUsername string, reason string, totalMoves, durationSeconds int, createdAt, finishedAt time.Time) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	defer tx.Rollback()

	// Update player1 stats
	player1Won := (winnerID != nil && *winnerID == player1ID)
	if err := UpdatePlayerStatsTx(tx, player1ID, player1Won); err != nil {
		return err
	}

	// Update player2 stats (if not bot)
	if player2Username != "BOT" && player2ID != nil {
		player2Won := (winnerID != nil && *winnerID == *player2ID)
		if err := UpdatePlayerStatsTx(tx, *player2ID, player2Won); err != nil {
			return err
		}
	}

	// Insert game record
	query := `
	INSERT INTO game (game_id, player1_id, player1_username, player2_id, player2_username, winner_id, winner_username, reason, total_moves, duration_seconds, created_at, finished_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
	`

	_, err = tx.Exec(query, gameID, player1ID, player1Username, player2ID, player2Username, winnerID, winnerUsername, reason, totalMoves, durationSeconds, createdAt, finishedAt)
	if err != nil {
		return fmt.Errorf("failed to insert game record: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	return nil
}

// UpdatePlayerStatsTx updates player stats within a transaction
func UpdatePlayerStatsTx(tx *sql.Tx, userID int64, won bool) error {
	query := `
	UPDATE players
	SET games_played = games_played + 1,
	    games_won = games_won + CASE WHEN $2 THEN 1 ELSE 0 END
	WHERE id = $1;
	`
	_, err := tx.Exec(query, userID, won)
	if err != nil {
		return fmt.Errorf("failed to update player stats in transaction: %v", err)
	}
	return nil
}

func GetLeaderboard() ([]PlayerStats, error) {
	query := `
	SELECT username, games_played, games_won,
		CASE 
			WHEN games_played > 0 THEN ROUND((games_won::decimal / games_played) * 100, 2)
			ELSE 0
		END AS win_rate
	FROM players
	ORDER BY games_won DESC, username ASC;
	`

	rows, err := DB.Query(query)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating leaderboard rows: %v", err)
	}

	return leaderboard, nil
}

// GameExists checks if a game with the given gameID exists in the database
func GameExists(gameID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM game WHERE game_id = $1);`
	var exists bool
	err := DB.QueryRow(query, gameID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if game exists: %v", err)
	}
	return exists, nil
}

// GameResult represents the result of a finished game
type GameResult struct {
	GameID          string
	Player1ID       int64
	Player1Username string
	Player2ID       *int64
	Player2Username string
	WinnerID        *int64
	WinnerUsername  string
	Reason          string
	TotalMoves      int
	DurationSeconds int
	CreatedAt       time.Time
	FinishedAt      time.Time
}

// GetGameByID retrieves game details from the database by gameID
func GetGameByID(gameID string) (*GameResult, error) {
	query := `
	SELECT game_id, player1_id, player1_username, player2_id, player2_username, 
	       winner_id, winner_username, reason, total_moves, duration_seconds, 
	       created_at, finished_at
	FROM game 
	WHERE game_id = $1;
	`
	
	var result GameResult
	var player2ID, winnerID sql.NullInt64
	var winnerUsername sql.NullString
	
	err := DB.QueryRow(query, gameID).Scan(
		&result.GameID,
		&result.Player1ID,
		&result.Player1Username,
		&player2ID,
		&result.Player2Username,
		&winnerID,
		&winnerUsername,
		&result.Reason,
		&result.TotalMoves,
		&result.DurationSeconds,
		&result.CreatedAt,
		&result.FinishedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get game by ID: %v", err)
	}
	
	if player2ID.Valid {
		id := player2ID.Int64
		result.Player2ID = &id
	}
	if winnerID.Valid {
		id := winnerID.Int64
		result.WinnerID = &id
	}
	if winnerUsername.Valid {
		result.WinnerUsername = winnerUsername.String
	}
	
	return &result, nil
}

