CREATE TABLE IF NOT EXISTS players (
    id SERIAL PRIMARY KEY,                     -- Auto-incrementing user ID
    username VARCHAR(50) UNIQUE NOT NULL,      -- Username must be unique
    password_hash VARCHAR(255) NOT NULL,       -- Bcrypt hashed password
    games_played INT DEFAULT 0,
    games_won INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for looking up by username (for login)
CREATE INDEX IF NOT EXISTS idx_players_username ON players(username);

CREATE TABLE IF NOT EXISTS game (
    game_id VARCHAR(50) PRIMARY KEY,           -- Use game_id as primary key
    player1_id INT REFERENCES players(id),     -- Player 1 ID (foreign key)
    player1_username VARCHAR(50) NOT NULL,     -- Player 1 username (for display)
    player2_id INT,                            -- Player 2 ID (NULL for BOT)
    player2_username VARCHAR(50) NOT NULL,     -- Player 2 username (for display, "BOT" for bot)
    winner_id INT,                             -- Winner's ID (NULL for tie, can reference players or be NULL for BOT)
    winner_username VARCHAR(50),               -- Winner's username (for display)
    reason VARCHAR(100),
    total_moves INT,
    duration_seconds INT,
    created_at TIMESTAMP,
    finished_at TIMESTAMP
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_game_player1_id ON game(player1_id);
CREATE INDEX IF NOT EXISTS idx_game_player2_id ON game(player2_id);
CREATE INDEX IF NOT EXISTS idx_players_username ON players(username);
CREATE INDEX IF NOT EXISTS idx_players_games_won ON players(games_won DESC);
CREATE INDEX IF NOT EXISTS idx_game_game_id ON game(game_id);
CREATE INDEX IF NOT EXISTS idx_game_created_at ON game(created_at DESC);