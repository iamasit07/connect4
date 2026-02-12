package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type GameRepo struct {
	DB *sql.DB
}

func NewGameRepo(db *sql.DB) *GameRepo {
	return &GameRepo{DB: db}
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

// SaveGame saves a finished game and updates player stats transactionally
func (r *GameRepo) SaveGame(gameID string, player1ID int64, player1Username string, player2ID *int64, player2Username string, winnerID *int64, winnerUsername string, reason string, totalMoves, durationSeconds int, createdAt, finishedAt time.Time, boardState [][]int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	defer tx.Rollback()

	// Update player1 stats
	player1Won := (winnerID != nil && *winnerID == player1ID)
	if err := r.updatePlayerStatsTx(tx, player1ID, player1Won); err != nil {
		return err
	}

	if player2ID != nil {
		player2Won := (winnerID != nil && *winnerID == *player2ID)
		if err := r.updatePlayerStatsTx(tx, *player2ID, player2Won); err != nil {
			return err
		}
	}

	// Insert or update game record (UPSERT to handle race conditions)
	boardJSON, err := json.Marshal(boardState)
	if err != nil {
		return fmt.Errorf("failed to marshal board state: %v", err)
	}

	query := `
	INSERT INTO game (game_id, player1_id, player1_username, player2_id, player2_username, winner_id, winner_username, reason, total_moves, duration_seconds, created_at, finished_at, board_state)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	ON CONFLICT (game_id) DO UPDATE SET
		winner_id = EXCLUDED.winner_id,
		winner_username = EXCLUDED.winner_username,
		reason = EXCLUDED.reason,
		total_moves = EXCLUDED.total_moves,
		duration_seconds = EXCLUDED.duration_seconds,
		finished_at = EXCLUDED.finished_at,
		board_state = EXCLUDED.board_state;
	`

	_, err = tx.Exec(query, gameID, player1ID, player1Username, player2ID, player2Username, winnerID, winnerUsername, reason, totalMoves, durationSeconds, createdAt, finishedAt, boardJSON)
	if err != nil {
		return fmt.Errorf("failed to upsert game record: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	return nil
}

// updatePlayerStatsTx updates player stats within a transaction
func (r *GameRepo) updatePlayerStatsTx(tx *sql.Tx, userID int64, won bool) error {
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

// GetGameByID retrieves game details from the database by gameID
func (r *GameRepo) GetGameByID(gameID string) (*GameResult, error) {
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
	
	err := r.DB.QueryRow(query, gameID).Scan(
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

// GetUserGameHistory retrieves all games for a user (both as player1 and player2)
func (r *GameRepo) GetUserGameHistory(userID int64) ([]GameResult, error) {
	query := `
	SELECT game_id, player1_id, player1_username, player2_id, player2_username, 
	       winner_id, winner_username, reason, total_moves, duration_seconds, 
	       created_at, finished_at
	FROM game 
	WHERE player1_id = $1 OR player2_id = $1
	ORDER BY finished_at DESC;
	`
	
	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query game history: %v", err)
	}
	defer rows.Close()
	
	var games []GameResult
	for rows.Next() {
		var result GameResult
		var player2ID, winnerID sql.NullInt64
		var winnerUsername sql.NullString
		
		err := rows.Scan(
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
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan game row: %v", err)
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
		
		games = append(games, result)
	}
	return games, nil
}

// GetGameBoard retrieves the board state for a game from the database
func (r *GameRepo) GetGameBoard(gameID string) ([][]int, error) {
	query := `SELECT board_state FROM game WHERE game_id = $1;`
	
	var boardJSON []byte
	err := r.DB.QueryRow(query, gameID).Scan(&boardJSON)
	if err == sql.ErrNoRows {
		// Return empty board if not found
		board := make([][]int, 6)
		for i := range board {
			board[i] = make([]int, 7)
		}
		return board, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get board state: %v", err)
	}
	
	if boardJSON == nil {
		board := make([][]int, 6)
		for i := range board {
			board[i] = make([]int, 7)
		}
		return board, nil
	}
	
	var board [][]int
	if err := json.Unmarshal(boardJSON, &board); err != nil {
		return nil, fmt.Errorf("failed to unmarshal board state: %v", err)
	}
	
	return board, nil
}
