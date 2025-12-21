package db

import (
	"database/sql"
	"fmt"
	"time"
)

type PlayerStats struct {
	Username    string  `json:"username"`
	GamesPlayed int     `json:"games_played"`
	GamesWon    int     `json:"games_won"`
	WinRate     float64 `json:"win_rate"`
}

func UpsertPlayer(userToken, username string, won bool) error {
	query := `
	INSERT INTO players (user_token, username, games_played, games_won)
	VALUES ($1, $2, 1, CASE WHEN $3 THEN 1 ELSE 0 END)
	ON CONFLICT (user_token)
	DO UPDATE SET
		username = $2,
		games_played = players.games_played + 1,
		games_won = players.games_won + CASE WHEN $3 THEN 1 ELSE 0 END;
	`
	winIncrement := 0
	if won {
		winIncrement = 1
	}

	_, err := DB.Exec(query, userToken, username, winIncrement)
	if err != nil {
		return fmt.Errorf("failed to upsert player: %v", err)
	}
	return nil
}

func SaveGame(gameID, player1Token, player1Username, player2Token, player2Username, winnerToken, winnerUsername, reason string, totalMoves, durationSeconds int, createdAt, finishedAt time.Time) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	defer tx.Rollback()

	// Update player1 stats
	player1Won := (winnerToken == player1Token)
	if err := UpsertPlayerTx(tx, player1Token, player1Username, player1Won); err != nil {
		return err
	}

	// Update player2 stats (if not bot)
	if player2Username != "BOT" {
		player2Won := (winnerToken == player2Token)
		if err := UpsertPlayerTx(tx, player2Token, player2Username, player2Won); err != nil {
			return err
		}
	}

	// Insert game record with tokens
	query := `
	INSERT INTO game (game_id, player1_token, player1_username, player2_token, player2_username, winner_token, winner_username, reason, total_moves, duration_seconds, created_at, finished_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
	`

	_, err = tx.Exec(query, gameID, player1Token, player1Username, player2Token, player2Username, winnerToken, winnerUsername, reason, totalMoves, durationSeconds, createdAt, finishedAt)
	if err != nil {
		return fmt.Errorf("failed to insert game record: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	return nil
}

func UpsertPlayerTx(tx *sql.Tx, userToken, username string, won bool) error {
	query := `
	INSERT INTO players (user_token, username, games_played, games_won)
	VALUES ($1, $2, 1, CASE WHEN $3 THEN 1 ELSE 0 END)
	ON CONFLICT (user_token)
	DO UPDATE SET
		username = $2,
		games_played = players.games_played + 1,
		games_won = players.games_won + CASE WHEN $3 THEN 1 ELSE 0 END;
	`
	winIncrement := 0
	if won {
		winIncrement = 1
	}

	_, err := tx.Exec(query, userToken, username, winIncrement)
	if err != nil {
		return fmt.Errorf("failed to upsert player in transaction: %v", err)
	}
	return nil
}

func GetLeaderboard(limit int) ([]PlayerStats, error) {
	query := `
	SELECT username, games_played, games_won,
		CASE 
			WHEN games_played > 0 THEN ROUND((games_won::decimal / games_played) * 100, 2)
			ELSE 0
		END AS win_rate
	FROM players
	ORDER BY win_rate DESC, games_played DESC, username ASC
	LIMIT $1;
	`

	rows, err := DB.Query(query, limit)
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
	GameID         string
	Player1Token   string
	Player1Username string
	Player2Token   string
	Player2Username string
	WinnerToken    string
	WinnerUsername string
	Reason         string
	TotalMoves     int
	DurationSeconds int
	CreatedAt      time.Time
	FinishedAt     time.Time
}

// GetGameByID retrieves game details from the database by gameID
func GetGameByID(gameID string) (*GameResult, error) {
	query := `
	SELECT game_id, player1_token, player1_username, player2_token, player2_username, 
	       winner_token, winner_username, reason, total_moves, duration_seconds, 
	       created_at, finished_at
	FROM game 
	WHERE game_id = $1;
	`
	
	var result GameResult
	var winnerToken, winnerUsername sql.NullString
	
	err := DB.QueryRow(query, gameID).Scan(
		&result.GameID,
		&result.Player1Token,
		&result.Player1Username,
		&result.Player2Token,
		&result.Player2Username,
		&winnerToken,
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
	
	if winnerToken.Valid {
		result.WinnerToken = winnerToken.String
	}
	if winnerUsername.Valid {
		result.WinnerUsername = winnerUsername.String
	}
	
	return &result, nil
}

