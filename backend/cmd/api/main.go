package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
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

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("No .env file found")
		}
	}

	cfg := config.LoadConfig()
	db, err := sql.Open("pgx", cfg.DatabaseURL)
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
	authHandler := transportHttp.NewAuthHandler(userRepo, sessionRepo, connManager, cache, authService, sessionManager)
	historyHandler := transportHttp.NewHistoryHandler(gameRepo)
	oauthHandler := transportHttp.NewOAuthHandler(userRepo, sessionRepo, &cfg.OAuthConfig, connManager)
	wsHandler := websocket.NewHandler(connManager, matchmakingQueue, sessionManager, gameService, authService)
	watchHandler := transportHttp.NewWatchHandler(sessionManager)

	// 7. Setup Gin Router
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Auth middleware for protected routes
	authMW := middleware.AuthMiddleware(authService)

	// Public Auth Routes
	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)
	router.GET("/api/leaderboard", authHandler.Leaderboard)

	// OAuth Routes (public)
	router.GET("/api/auth/google/login", oauthHandler.GoogleLogin)
	router.GET("/api/auth/google/callback", oauthHandler.GoogleCallback)
	router.POST("/api/auth/google/complete", oauthHandler.CompleteGoogleSignup)

	// Protected Routes
	protected := router.Group("/")
	protected.Use(authMW)
	{
		protected.POST("/api/auth/logout", authHandler.Logout)
		protected.GET("/api/auth/me", authHandler.Me)
		protected.PUT("/api/auth/profile", authHandler.UpdateProfile)
		protected.POST("/api/auth/avatar", authHandler.UploadAvatar)
		protected.DELETE("/api/auth/avatar/remove", authHandler.RemoveAvatar)

		// Game History Routes
		protected.GET("/api/history", historyHandler.GetHistory)
		protected.GET("/api/history/:id", historyHandler.GetGameDetails)
		protected.GET("/api/sessions", authHandler.GetSessionHistory)

		// Watch / Spectator Routes
		protected.GET("/api/watch", watchHandler.GetLiveGames)
	}

	// WebSocket Route (auth handled inside the WS handler itself)
	router.GET("/ws", wsHandler.HandleWebSocket)

	// Serve uploaded files (avatars)
	router.Static("/uploads", "./uploads")

	// Serve static frontend files (SPA fallback)
	if _, err := os.Stat("./static"); err == nil {
		router.Static("/assets", "./static/assets")

		// Serve known static file extensions directly
		router.GET("/", func(c *gin.Context) {
			c.File("./static/index.html")
		})

		// SPA fallback: serve index.html for all unmatched routes
		router.NoRoute(func(c *gin.Context) {
			path := "./static" + c.Request.URL.Path

			// Serve actual static files if they exist
			if info, err := os.Stat(path); err == nil && !info.IsDir() {
				c.File(path)
				return
			}

			// For asset requests that don't exist, return 404
			if strings.HasPrefix(c.Request.URL.Path, "/assets/") || strings.HasSuffix(c.Request.URL.Path, ".css") || strings.HasSuffix(c.Request.URL.Path, ".js") {
				c.Status(http.StatusNotFound)
				return
			}

			// SPA fallback
			c.File("./static/index.html")
		})
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
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
