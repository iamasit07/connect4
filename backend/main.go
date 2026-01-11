package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/iamasit07/4-in-a-row/backend/config"
	"github.com/iamasit07/4-in-a-row/backend/db"
	"github.com/iamasit07/4-in-a-row/backend/handlers"
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
	matchMakingQueue := models.NewMatchmakingQueue()
	sessionManager := server.NewSessionManager()

	go MatchMakingListener(matchMakingQueue, connectionManager, sessionManager)

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.CreateUpgrader()

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error during connection upgrade:", err)
			return
		}

		websocket.HandleConnection(conn, connectionManager, matchMakingQueue, sessionManager)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Auth routes
	mux.HandleFunc("/api/auth/signup", handlers.HandleSignup)
	mux.HandleFunc("/api/auth/login", handlers.MakeHandleLogin(connectionManager))
	mux.HandleFunc("/api/auth/logout", handlers.HandleLogout)
	mux.HandleFunc("/api/auth/me", handlers.HandleMe)

	mux.HandleFunc("/api/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		leaderboard, err := db.GetLeaderboard()
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

	// Start server in a separate goroutine
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

func MatchMakingListener(queue *models.MatchmakingQueue, cm *websocket.ConnectionManager, sm *server.SessionManager) {
	for {
		match := <-queue.MatchChannel

		player1ID := match.Player1ID
		player1Username := match.Player1Username
		player2ID := match.Player2ID
		player2Username := match.Player2Username

		log.Printf("[MATCHMAKING] Match found: %s (ID: %d) vs %s (ID: %v)",
			player1Username, player1ID, player2Username, player2ID)

		// Check if player1 has any active sessions
		if sm.HasActiveGame(player1ID) {
			player1Session, _ := sm.GetSessionByUserID(player1ID)
			if player1Session != nil && !player1Session.Game.IsFinished() {
				log.Printf("[MATCHMAKING] Player1 %s has active game %s - terminating it before new match",
					player1Username, player1Session.GameID)

				// Terminate the session
				err := player1Session.TerminateSessionByAbandonment(player1ID, cm)
				if err != nil {
					log.Printf("[MATCHMAKING] Failed to terminate player1's session: %v", err)
				}

				sm.RemoveSession(player1Session.GameID)
			}
		}

		// Check if player2 has any active sessions (if not bot)
		if player2Username != models.BotUsername && player2ID != nil {
			if sm.HasActiveGame(*player2ID) {
				player2Session, _ := sm.GetSessionByUserID(*player2ID)
				if player2Session != nil && !player2Session.Game.IsFinished() {
					log.Printf("[MATCHMAKING] Player2 %s has active game %s - terminating it before new match",
						player2Username, player2Session.GameID)

					err := player2Session.TerminateSessionByAbandonment(*player2ID, cm)
					if err != nil {
						log.Printf("[MATCHMAKING] Failed to terminate player2's session: %v", err)
					}

					sm.RemoveSession(player2Session.GameID)
				}
			}
		}

		// CreateSession will handle sending game_start messages
		session := sm.CreateSession(player1ID, player1Username, player2ID, player2Username, match.BotDifficulty, cm)

		log.Printf("Match started between %s (ID: %d) and %s (ID: %v) with game ID %s\n",
			player1Username, player1ID, player2Username, player2ID, session.GameID)
	}
}
