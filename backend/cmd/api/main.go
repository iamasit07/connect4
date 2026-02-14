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
	"github.com/iamasit07/4-in-a-row/backend/internal/domain"
	"github.com/iamasit07/4-in-a-row/backend/internal/repository/postgres"
	"github.com/iamasit07/4-in-a-row/backend/internal/repository/redis"
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
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("No .env file found")
		}
	}

	cfg := config.LoadConfig()
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

	// DIAGNOSTIC: Check for any triggers on the players table
	rows, trigErr := db.Query(`SELECT trigger_name, event_manipulation, action_statement FROM information_schema.triggers WHERE event_object_table = 'players'`)
	if trigErr == nil {
		defer rows.Close()
		for rows.Next() {
			var name, event, action string
			rows.Scan(&name, &event, &action)
			log.Printf("[DIAG] TRIGGER on players: name=%s, event=%s, action=%s", name, event, action)
		}
	}
	// Also check how many players exist
	var playerCount int
	db.QueryRow(`SELECT COUNT(*) FROM players`).Scan(&playerCount)
	log.Printf("[DIAG] Players table has %d rows", playerCount)

	// 3. Initialize Repositories (Persistence Layer)
	gameRepo := postgres.NewGameRepo(db)
	userRepo := postgres.NewUserRepo(db)
	sessionRepo := postgres.NewSessionRepo(db)

	// 3b. Initialize Redis
	if err := redis.InitRedis(); err != nil {
		log.Printf("Failed to initialize Redis: %v", err)
	}
	defer redis.CloseRedis()

	// 4. Initialize Services (Business Logic Layer)
	gameService := game.NewService(gameRepo)
	sessionManager := game.NewSessionManager(gameRepo)
	
	// Setup Redis Cache wrapper if Redis is enabled
	var cache session.CacheRepository
	if redis.IsRedisEnabled() && redis.RedisClient != nil {
		cache = redis.NewRedisCache(redis.RedisClient)
	}
	
	authService := session.NewAuthService(sessionRepo, cache)
	connManager := websocket.NewConnectionManager()
	
	// Define timeout callback for matchmaking
	onMatchmakingTimeout := func(userID int64) {
		connManager.SendMessage(userID, domain.ServerMessage{Type: "queue_timeout"})
	}
	matchmakingQueue := matchmaking.NewMatchmakingQueue(onMatchmakingTimeout)

	// 5. Initialize Background Workers
	cleanupWorker := cleanup.NewWorker(sessionManager, sessionRepo) 
	go cleanupWorker.Start()

	go matchmaking.MatchMakingListener(matchmakingQueue, connManager, sessionManager)
	
	// 6. Initialize HTTP Handlers (API Layer)
	authHandler := transportHttp.NewAuthHandler(userRepo, sessionRepo, connManager, cache)
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
	mux.HandleFunc("/api/auth/profile", protected(authHandler.UpdateProfile))
	mux.HandleFunc("/api/leaderboard", authHandler.Leaderboard)

	// OAuth Routes
	mux.HandleFunc("/api/auth/google/login", oauthHandler.GoogleLogin)
	mux.HandleFunc("/api/auth/google/callback", oauthHandler.GoogleCallback)
	mux.HandleFunc("/api/auth/google/complete", oauthHandler.CompleteGoogleSignup)

	// Game History Routes
	mux.HandleFunc("/api/history", protected(historyHandler.GetHistory))
	mux.HandleFunc("/api/history/", protected(historyHandler.GetGameDetails))
	mux.HandleFunc("/api/sessions", protected(authHandler.GetSessionHistory))

	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)

	if _, err := os.Stat("./static"); err == nil {
		fileServer := http.FileServer(http.Dir("./static"))
		
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			path := "./static" + r.URL.Path
			
			if r.URL.Path == "/" {
				http.ServeFile(w, r, "./static/index.html")
				return
			}
			
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				fileServer.ServeHTTP(w, r)
				return
			}
			
			if len(r.URL.Path) > 4 && (r.URL.Path[0:8] == "/assets/" || r.URL.Path[len(r.URL.Path)-4:] == ".css" || r.URL.Path[len(r.URL.Path)-3:] == ".js") {
				http.NotFound(w, r)
				return
			}
			
			// 4. Otherwise, serve index.html (client-side routing)
			http.ServeFile(w, r, "./static/index.html")
		})
	}

	handler := middleware.EnableCORS(mux)
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
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