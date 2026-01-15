package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/internal/config"
	"github.com/iamasit07/4-in-a-row/backend/internal/repository/postgres"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/cleanup"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/game"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/matchmaking"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/session"
	transportHttp "github.com/iamasit07/4-in-a-row/backend/internal/transport/http"
	"github.com/iamasit07/4-in-a-row/backend/internal/transport/http/middleware"
	"github.com/iamasit07/4-in-a-row/backend/internal/transport/websocket"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	// 1. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 2. Load Config
	cfg := config.LoadConfig()

	// 2b. Database Connection with Pooling
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Apply Pool Settings
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetimeMin) * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatal("Database unreachable:", err)
	}

	// 2c. Run Migrations
	log.Println("Running database migrations...")
	if err := postgres.RunMigrations(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Database migration completed successfully")

	// 3. Initialize Repositories (Persistence Layer)
	gameRepo := postgres.NewGameRepo(db)
	userRepo := postgres.NewUserRepo(db)
	sessionRepo := postgres.NewSessionRepo(db)

	// 4. Initialize Services (Business Logic Layer)
	gameService := game.NewService(gameRepo)
	sessionManager := game.NewSessionManager(gameRepo)
	authService := session.NewAuthService(sessionRepo, nil) // No redis/cache for now
	connManager := websocket.NewConnectionManager()
	matchmakingQueue := matchmaking.NewMatchmakingQueue()

	// 5. Initialize Background Workers
	cleanupWorker := cleanup.NewWorker(sessionManager, sessionRepo) 
	go cleanupWorker.Start()

	go matchmaking.MatchMakingListener(matchmakingQueue, connManager, sessionManager)
	
	// 6. Initialize HTTP Handlers (API Layer)
	authHandler := transportHttp.NewAuthHandler(userRepo, sessionRepo, connManager)
	historyHandler := transportHttp.NewHistoryHandler(gameRepo)
	oauthHandler := transportHttp.NewOAuthHandler(userRepo, sessionRepo, &cfg.OAuthConfig, connManager)
	wsHandler := websocket.NewHandler(connManager, matchmakingQueue, sessionManager, gameService, authService)

	// 7. Setup Router
	mux := http.NewServeMux()

	protected := func(handler http.HandlerFunc) http.HandlerFunc {
		return middleware.AuthMiddleware(handler, sessionRepo)
	}

	// Auth Routes
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/logout", authHandler.Logout)
	mux.HandleFunc("/api/auth/me", protected(authHandler.Me))
	mux.HandleFunc("/api/leaderboard", authHandler.Leaderboard)

	// OAuth Routes
	mux.HandleFunc("/api/auth/google/login", oauthHandler.GoogleLogin)
	mux.HandleFunc("/api/auth/google/callback", oauthHandler.GoogleCallback)
	mux.HandleFunc("/api/auth/google/complete", oauthHandler.CompleteGoogleSignup)

	// Game History Routes
	mux.HandleFunc("/api/history", protected(historyHandler.GetHistory))
	mux.HandleFunc("/api/history/", protected(historyHandler.GetGameDetails))
	mux.HandleFunc("/api/sessions", protected(authHandler.GetSessionHistory))

	// WebSocket Route
	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// 8. Apply Global Middleware (CORS)
	handler := middleware.EnableCORS(mux)

	// 9. Start Server with Configured Port and Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port, // Use Port from config
		Handler: handler,
	}

	go func() {
		log.Printf("Server starting on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}