package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/iamasit07/4-in-a-row/backend/config"
	"github.com/iamasit07/4-in-a-row/backend/db"
	"github.com/iamasit07/4-in-a-row/backend/middlewares"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/server"
	"github.com/iamasit07/4-in-a-row/backend/websocket"
)

func main() {
	log.Println("Starting 4-in-a-row backend server...")

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config.LoadConfig()

	dbUri := config.GetEnv("DB_URI", "")
	port := config.GetEnv("PORT", "8080")
	
	dbMaxOpenConns := config.GetEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	dbMaxIdleConns := config.GetEnvAsInt("DB_MAX_IDLE_CONNS", 25)
	dbConnMaxLifetimeMin := config.GetEnvAsInt("DB_CONN_MAX_LIFETIME_MINUTES", 5)

	err = db.InitDB(dbUri, dbMaxOpenConns, dbMaxIdleConns, dbConnMaxLifetimeMin)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.CloseDB()

	connectionManager := websocket.NewConnectionManager()

	timerMap := make(map[string]*time.Timer)
	waitingPlayersMap := make(map[string]string)  // token → username
	matchMakingQueue := &models.MatchmakingQueue{
		WaitingPlayers: waitingPlayersMap,
		MatchChannel:   make(chan models.Match, 100),
		Mux:            &sync.Mutex{},
		Timer:          &timerMap,
	}

	sessionManager := &server.SessionManager{
		Session:     make(map[string]*server.GameSession),
		TokenToGame: make(map[string]string),
		Mux:         &sync.Mutex{},
	}

	tokenManager := websocket.NewTokenManager()

	go MatchMakingListener(matchMakingQueue, connectionManager, sessionManager, tokenManager)

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.CreateUpgrader()
		
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error during connection upgrade:", err)
			return
		}

		websocket.HandleConnection(conn, connectionManager, matchMakingQueue, sessionManager, tokenManager)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/api/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		limitSetter := r.URL.Query().Get("limit")
		limit := 10
		if limitSetter != "" {
			fmt.Sscan(limitSetter, &limit)
		}

		leaderboard, err := db.GetLeaderboard(limit)
		if err != nil {
			http.Error(w, "Failed to fetch leaderboard", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(leaderboard)
	})

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: middlewares.EnableCORS(mux),
	}

	log.Printf("Server is listening on port %s\n", port)

	// Start server in goroutine
	go func() {
		err = httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Server is shutting down gracefully...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

func MatchMakingListener(queue *models.MatchmakingQueue, cm *websocket.ConnectionManager, sm *server.SessionManager, tm *websocket.TokenManager) {
	for {
		match := <-queue.MatchChannel

		player1Token := match.Player1Token
		player1Username := match.Player1Username
		player2Token := match.Player2Token
		player2Username := match.Player2Username

		// If player2 is bot, use BOT_TOKEN from config
		if player2Username == models.BotUsername {
			player2Token = config.AppConfig.BotToken
		}

		// CreateSession will handle sending game_start messages
		session := sm.CreateSession(player1Token, player1Username, player2Token, player2Username, cm)

		log.Printf("Match started between %s (%s) and %s (%s) with game ID %s\n",
			player1Username, player1Token, player2Username, player2Token, session.GameID)
	}
}
