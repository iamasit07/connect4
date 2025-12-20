package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
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
	fmt.Println("Starting 4-in-a-row backend server...")

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
	matchMakingQueue := &models.MatchmakingQueue{
		WaitingPlayers: []string{},
		MatchChannel:   make(chan models.Match, 100),
		Mux:            &sync.Mutex{},
		Timer:          &timerMap,
	}

	sessionManager := &server.SessionManager{
		Session:    make(map[string]*server.GameSession),
		UserToGame: make(map[string]string),
		Mux:        &sync.Mutex{},
	}

	tokenManager := websocket.NewTokenManager()

	go MatchMakingListener(matchMakingQueue, connectionManager, sessionManager, tokenManager)

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		connectionManager.HandleWebSocket(w, r, sessionManager, matchMakingQueue, tokenManager)
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

	fmt.Printf("Server is listening on port %s\n", port)

	err = httpServer.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func MatchMakingListener(queue *models.MatchmakingQueue, cm *websocket.ConnectionManager, sm *server.SessionManager, tm *websocket.TokenManager) {
	for {
		match := <-queue.MatchChannel

		player1Username := match.Player1
		player2Username := match.Player2

		// Get userTokens for both players from TokenManager
		userTokens := make(map[string]string)
		if token1, exists := tm.GetTokenByUsername(player1Username); exists {
			userTokens[player1Username] = token1
		}
		if token2, exists := tm.GetTokenByUsername(player2Username); exists {
			userTokens[player2Username] = token2
		}

		// CreateSession will handle sending game_start messages
		session := sm.CreateSession(player1Username, player2Username, userTokens, cm)

		fmt.Printf("Match started between %s and %s with game ID %s\n",
			player1Username, player2Username, session.GameID)
	}
}
