CREATE TABLE IF NOT EXISTS players (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    name TEXT DEFAULT '',
    email TEXT UNIQUE,
    google_id TEXT UNIQUE,
    avatar_url TEXT DEFAULT '',
    is_verified BOOLEAN DEFAULT FALSE,
    password_hash TEXT NOT NULL,
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
    game_id TEXT PRIMARY KEY,
    player1_id INT REFERENCES players(id),
    player1_username TEXT NOT NULL,
    player2_id INT,
    player2_username TEXT NOT NULL,
    winner_id INT,
    winner_username TEXT,
    reason TEXT,
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
    session_id TEXT UNIQUE NOT NULL,
    device_info TEXT,
    ip_address TEXT,
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

-- Refresh tokens table for access/refresh token rotation
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    token_id TEXT UNIQUE NOT NULL,
    user_id INT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    session_id TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN DEFAULT FALSE
);

-- Refresh token indexes
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_id ON refresh_tokens(token_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_session_id ON refresh_tokens(session_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_cleanup ON refresh_tokens(created_at) WHERE revoked = TRUE;

-- Enable Row Level Security
ALTER TABLE players ENABLE ROW LEVEL SECURITY;
ALTER TABLE game ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE refresh_tokens ENABLE ROW LEVEL SECURITY;