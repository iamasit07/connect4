CREATE TABLE IF NOT EXISTS players (
    user_token VARCHAR(100) PRIMARY KEY,  -- Token is now primary key
    username VARCHAR(50) NOT NULL,        -- Username can be duplicate
    games_played INT DEFAULT 0,
    games_won INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for looking up by username (for display)
CREATE INDEX IF NOT EXISTS idx_players_username ON players(username);

CREATE TABLE IF NOT EXISTS game (
    game_id VARCHAR(50) PRIMARY KEY,           -- Use game_id as primary key
    player1_token VARCHAR(100) NOT NULL,       -- Player 1 token
    player1_username VARCHAR(50) NOT NULL,     -- Player 1 username (for display)
    player2_token VARCHAR(100) NOT NULL,       -- Player 2 token
    player2_username VARCHAR(50) NOT NULL,     -- Player 2 username (for display)
    winner_token VARCHAR(100),                 -- Winner's token
    winner_username VARCHAR(50),               -- Winner's username (for display)
    reason VARCHAR(100),
    total_moves INT,
    duration_seconds INT,
    created_at TIMESTAMP,
    finished_at TIMESTAMP
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_game_player1_token ON game(player1_token);
CREATE INDEX IF NOT EXISTS idx_game_player2_token ON game(player2_token);
CREATE INDEX IF NOT EXISTS idx_players_username ON players(username);
CREATE INDEX IF NOT EXISTS idx_players_games_won ON players(games_won DESC);
CREATE INDEX IF NOT EXISTS idx_game_game_id ON game(game_id);
CREATE INDEX IF NOT EXISTS idx_game_created_at ON game(created_at DESC);