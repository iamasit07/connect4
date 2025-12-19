package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"

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
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbUri := os.Getenv("DB_URI")

	// db connection
	err = db.InitDB(dbUri)
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

	go MatchMakingListener(matchMakingQueue, connectionManager, sessionManager)

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		connectionManager.HandleWebSocket(w, r, sessionManager, matchMakingQueue)
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
		Addr:    ":8080",
		Handler: middlewares.EnableCORS(mux),
	}

	fmt.Println("Server is listening on port 8080")

	err = httpServer.ListenAndServe()
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func MatchMakingListener(queue *models.MatchmakingQueue, cm *websocket.ConnectionManager, sm *server.SessionManager) {
	for {
		match := <-queue.MatchChannel

		player1Username := match.Player1
		player2Username := match.Player2

		// CreateSession will handle sending game_start messages
		session := sm.CreateSession(player1Username, player2Username, cm)

		fmt.Printf("Match started between %s and %s with game ID %s\n",
			player1Username, player2Username, session.GameID)
	}
}
