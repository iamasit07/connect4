package config

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port                 string
	MatchmakingTimeout   time.Duration
	BotToken             string
	AllowedOrigins       []string
	OAuthConfig          OAuthConfig
	DatabaseURL          string
	DBMaxOpenConns       int
	DBMaxIdleConns       int
	DBConnMaxLifetimeMin int
	FrontendURL          string
	JWTSecret            string
}

var AppConfig *Config

func LoadConfig() *Config {
	port := GetEnv("PORT", "8080")
	matchmakingTimeoutSec := GetEnvAsInt("MATCHMAKING_TIMEOUT_SECONDS", 300)
	botToken := GetEnv("BOT_TOKEN", "tkn_bot_default")

	// Frontend & CORS
	frontendURL := GetEnv("FRONTEND_URL", "https://4-in-a-row.iamasit07.me")
	allowedOriginsStr := GetEnv("ALLOWED_ORIGINS", "")

	// Build allowed origins list (Frontend URL + Localhost + CSV values)
	allowedOrigins := []string{
		frontendURL,
		"http://localhost:5173", // Local development
	}
	if allowedOriginsStr != "" {
		extras := strings.Split(allowedOriginsStr, ",")
		for _, origin := range extras {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	// Database Config
	// Append simple_protocol for PgBouncer compatibility (pgx driver)
	dbURL := GetEnv("DATABASE_URL", GetEnv("DATABASE_URI", ""))
	if dbURL != "" {
		if u, err := url.Parse(dbURL); err == nil {
			q := u.Query()
			if q.Get("default_query_exec_mode") == "" {
				q.Set("default_query_exec_mode", "simple_protocol")
				u.RawQuery = q.Encode()
				dbURL = u.String()
			}
		}
	}
	dbMaxOpenConns := GetEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	dbMaxIdleConns := GetEnvAsInt("DB_MAX_IDLE_CONNS", 25)
	dbConnMaxLifetimeMin := GetEnvAsInt("DB_CONN_MAX_LIFETIME_MINUTES", 5)

	// Security
	jwtSecret := GetEnv("JWT_SECRET", "your-secret-key-change-this-in-production")

	oauthConfig := LoadOAuthConfig(frontendURL)

	AppConfig = &Config{
		Port:                 port,
		MatchmakingTimeout:   time.Duration(matchmakingTimeoutSec) * time.Second,
		BotToken:             botToken,
		AllowedOrigins:       allowedOrigins,
		OAuthConfig:          *oauthConfig,
		DatabaseURL:          dbURL,
		DBMaxOpenConns:       dbMaxOpenConns,
		DBMaxIdleConns:       dbMaxIdleConns,
		DBConnMaxLifetimeMin: dbConnMaxLifetimeMin,
		FrontendURL:          frontendURL,
		JWTSecret:            jwtSecret,
	}

	return AppConfig
}

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Invalid integer value for %s: %s, using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}
