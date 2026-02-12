package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/internal/repository/postgres"
	"github.com/iamasit07/4-in-a-row/backend/pkg/auth"
	"github.com/iamasit07/4-in-a-row/backend/pkg/httputil"
	"github.com/iamasit07/4-in-a-row/backend/pkg/useragent"
)

type Disconnector interface {
	DisconnectUser(userID int64, reason string)
}

type AuthHandler struct {
	UserRepo    *postgres.UserRepo
	SessionRepo *postgres.SessionRepo
	ConnManager Disconnector
}

func NewAuthHandler(userRepo *postgres.UserRepo, sessionRepo *postgres.SessionRepo, cm Disconnector) *AuthHandler {
	return &AuthHandler{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		ConnManager: cm,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if len(req.Username) < 3 || len(req.Username) > 50 {
		http.Error(w, "Username must be between 3 and 50 characters", http.StatusBadRequest)
		return
	}

	if strings.ToUpper(req.Username) == "BOT" {
		http.Error(w, "Username 'BOT' is reserved", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		http.Error(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	existing, _ := h.UserRepo.GetUserByIdentifier(req.Username)
	if existing != nil {
		http.Error(w, "Username or email already taken", http.StatusConflict)
		return
	}

	hashedPwd, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userID, err := h.UserRepo.CreateUser(req.Username, hashedPwd, req.Email, "")
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Create Session
	sessionID := auth.GenerateToken()
	deviceInfo := useragent.ExtractDeviceInfo(r)
	ipAddress := useragent.ExtractIPAddress(r)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	err = h.SessionRepo.CreateSession(userID, sessionID, deviceInfo, ipAddress, expiresAt)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateJWT(userID, req.Username, sessionID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	httputil.SetAuthCookie(w, token)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":    token,
		"username": req.Username,
		"user_id":  userID,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	user, err := h.UserRepo.GetUserByIdentifier(req.Username)
	if err != nil || user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 1. Deactivate old sessions
	err = h.SessionRepo.DeactivateAllUserSessions(user.ID)
	if err != nil {
		// Log error but proceed
	}

	if h.ConnManager != nil {
		h.ConnManager.DisconnectUser(user.ID, "Logged in from another device")
	}

	sessionID := auth.GenerateToken()
	deviceInfo := useragent.ExtractDeviceInfo(r)
	ipAddress := useragent.ExtractIPAddress(r)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	err = h.SessionRepo.CreateSession(user.ID, sessionID, deviceInfo, ipAddress, expiresAt)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateJWT(user.ID, user.Username, sessionID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	httputil.SetAuthCookie(w, token)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":    token,
		"username": user.Username,
		"user_id":  user.ID,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	httputil.ClearAuthCookie(w)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out"))
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get the token from the request to return it to the client
	token, err := httputil.GetTokenFromRequest(r)
	if err != nil {
		log.Printf("[AUTH] /me: Failed to get token for user %d: %v", userID, err)
		http.Error(w, "Token not found", http.StatusUnauthorized)
		return
	}

	log.Printf("[AUTH] /me: Returning token for user %d, token length: %d", userID, len(token))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":      user.ID,
		"username":     user.Username,
		"token":        token,
		"games_played": user.GamesPlayed,
		"games_won":    user.GamesWon,
	})
}

func (h *AuthHandler) Leaderboard(w http.ResponseWriter, r *http.Request) {
	rawStats, err := h.UserRepo.GetLeaderboard()
	if err != nil {
		http.Error(w, "Failed to fetch leaderboard", http.StatusInternalServerError)
		return
	}

	// Map to frontend expectation
	type LeaderboardEntry struct {
		Rank     int    `json:"rank"`
		Username string `json:"username"`
		Rating   int    `json:"rating"` // Simple rating based on wins for now
		Wins     int    `json:"wins"`
		Losses   int    `json:"losses"`
	}

	leaderboard := make([]LeaderboardEntry, 0, len(rawStats))
	for i, stat := range rawStats {
		losses := stat.GamesPlayed - stat.GamesWon
		if losses < 0 {
			losses = 0
		}
		
		entry := LeaderboardEntry{
			Rank:     i + 1,
			Username: stat.Username,
			Rating:   stat.GamesWon * 10, // Simple placeholder rating
			Wins:     stat.GamesWon,
			Losses:   losses,
		}
		leaderboard = append(leaderboard, entry)
	}

	json.NewEncoder(w).Encode(leaderboard)
}

func (h *AuthHandler) GetSessionHistory(w http.ResponseWriter, r *http.Request) {
	// 1. Get UserID from context (set by AuthMiddleware)
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// 2. Fetch Sessions
	sessions, err := h.SessionRepo.GetUserSessionHistory(userID, 10)
	if err != nil {
		http.Error(w, "Failed to fetch session history", http.StatusInternalServerError)
		return
	}

	// 3. Return JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sessions); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
