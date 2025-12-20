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

func UpsertPlayer(username string, won bool) error {
	query := `
	INSERT INTO players (username, games_played, games_won)
	VALUES ($1, 1, CASE WHEN $2 THEN 1 ELSE 0 END)
	ON CONFLICT (username)
	DO UPDATE SET
		games_played = players.games_played + 1,
		games_won = players.games_won + CASE WHEN $2 THEN 1 ELSE 0 END;
	`
	winIncrement := 0
	if won {
		winIncrement = 1
	}

	_, err := DB.Exec(query, username, winIncrement)
	if err != nil {
		return fmt.Errorf("failed to upsert player: %v", err)
	}
	return nil
}

func SaveGame(gameID, player1, player2, winner, reason string, totalMoves, durationSeconds int, createdAt, finishedAt time.Time) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	defer tx.Rollback()

	playerWon := (winner == player1)
	if err := UpsertPlayerTx(tx, player1, playerWon); err != nil {
		return err
	}

	playerWon = (winner == player2 && player2 != "BOT")
	if err := UpsertPlayerTx(tx, player2, playerWon); err != nil {
		return err
	}

	query := `
	INSERT INTO game (id, game_id, player1_username, player2_username, winner, reason, total_moves, duration_seconds, created_at, finished_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);
	`

	_, err = tx.Exec(query, gameID, gameID, player1, player2, winner, reason, totalMoves, durationSeconds, createdAt, finishedAt)
	if err != nil {
		return fmt.Errorf("failed to insert game record: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	return nil
}

func UpsertPlayerTx(tx *sql.Tx, username string, won bool) error {
	query := `
	INSERT INTO players (username, games_played, games_won)
	VALUES ($1, 1, CASE WHEN $2 THEN 1 ELSE 0 END)
	ON CONFLICT (username)
	DO UPDATE SET
		games_played = players.games_played + 1,
		games_won = players.games_won + CASE WHEN $2 THEN 1 ELSE 0 END;
	`
	winIncrement := 0
	if won {
		winIncrement = 1
	}

	_, err := tx.Exec(query, username, winIncrement)
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
	ORDER BY win_rate DESC, games_played DESC
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
