CREATE TABLE IF NOT EXISTS players (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) DEFAULT '',
    email VARCHAR(255) UNIQUE,
    google_id VARCHAR(255) UNIQUE,
    avatar_url TEXT DEFAULT '',
    is_verified BOOLEAN DEFAULT FALSE,
    password_hash VARCHAR(255) NOT NULL,
    rating INT DEFAULT 1000,
    games_played INT DEFAULT 0,
    games_won INT DEFAULT 0,
    games_drawn INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Player indexes
CREATE INDEX IF NOT EXISTS idx_players_username ON players(username);
CREATE INDEX IF NOT EXISTS idx_players_games_won ON players(games_won DESC);

CREATE TABLE IF NOT EXISTS game (
    game_id VARCHAR(50) PRIMARY KEY,
    player1_id INT REFERENCES players(id),
    player1_username VARCHAR(50) NOT NULL,
    player2_id INT,
    player2_username VARCHAR(50) NOT NULL,
    winner_id INT,
    winner_username VARCHAR(50),
    reason VARCHAR(100),
    total_moves INT,
    duration_seconds INT,
    created_at TIMESTAMP,
    finished_at TIMESTAMP,
    board_state JSONB
);

-- Game indexes
CREATE INDEX IF NOT EXISTS idx_game_player1_id ON game(player1_id);
CREATE INDEX IF NOT EXISTS idx_game_player2_id ON game(player2_id);
CREATE INDEX IF NOT EXISTS idx_game_game_id ON game(game_id);
CREATE INDEX IF NOT EXISTS idx_game_created_at ON game(created_at DESC);

-- User sessions table for single-device enforcement
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    session_id VARCHAR(64) UNIQUE NOT NULL,
    device_info VARCHAR(255),
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

-- Session indexes
CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON user_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_active ON user_sessions(user_id, is_active) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_sessions_cleanup ON user_sessions(created_at) WHERE is_active = FALSE;
CREATE UNIQUE INDEX IF NOT EXISTS idx_one_active_session ON user_sessions(user_id) WHERE is_active = TRUE;

-- Enable Row Level Security
ALTER TABLE players ENABLE ROW LEVEL SECURITY;
ALTER TABLE game ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_sessions ENABLE ROW LEVEL SECURITY;