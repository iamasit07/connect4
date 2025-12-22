package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ReconnectTimeout      time.Duration
	BotMatchmakingTimeout time.Duration
	BotUsername           string
	BotToken              string // Special token for bot games
	AllowedOrigins        []string
}

var AppConfig *Config

func LoadConfig() {
	reconnectTimeoutSec := GetEnvAsInt("RECONNECT_TIMEOUT_SECONDS", 30)
	botMatchmakingTimeoutSec := GetEnvAsInt("BOT_MATCHMAKING_TIMEOUT_SECONDS", 10)
	botUsername := GetEnv("BOT_USERNAME", "BOT")
	botToken := GetEnv("BOT_TOKEN", "tkn_bot_default") // Default bot token
	frontendURL := GetEnv("FRONTEND_URL", "https://4-in-a-row.iamasit07.me")

	// Build allowed origins list
	allowedOrigins := []string{
		frontendURL,
		"http://localhost:5173", // Local development
	}

	AppConfig = &Config{
		ReconnectTimeout:      time.Duration(reconnectTimeoutSec) * time.Second,
		BotMatchmakingTimeout: time.Duration(botMatchmakingTimeoutSec) * time.Second,
		BotUsername:           botUsername,
		BotToken:              botToken, // Assign bot token
		AllowedOrigins:        allowedOrigins,
	}

	log.Printf("Config loaded: ReconnectTimeout=%v, BotMatchmakingTimeout=%v, BotUsername=%s, AllowedOrigins=%v",
		AppConfig.ReconnectTimeout, AppConfig.BotMatchmakingTimeout, AppConfig.BotUsername, AppConfig.AllowedOrigins)
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
