CREATE TABLE IF NOT EXISTS players (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    games_played INT DEFAULT 0,
    games_won INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
)

CREATE TABLE IF NOT EXISTS game (
    id VARCHAR(32) PRIMARY KEY,
    game_id VARCHAR(100) UNIQUE NOT NULL,
    player1_username VARCHAR(100) NOT NULL,
    player2_username VARCHAR(100) NOT NULL,
    winner VARCHAR(100),
    reason VARCHAR(255) NOT NULL,
    total_moves INT NOT NULL,
    duration_seconds INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMP NOT NULL
)

CREATE INDEX IF NOT EXISTS idx_players_username ON players(username);
CREATE INDEX IF NOT EXISTS idx_players_games_won ON players(games_won DESC);
CREATE INDEX IF NOT EXISTS idx_game_game_id ON game(game_id);
CREATE INDEX IF NOT EXISTS idx_game_created_at ON game(created_at DESC);