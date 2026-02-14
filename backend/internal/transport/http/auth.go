package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/internal/repository/postgres"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/session"
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
	Cache       session.CacheRepository
}

func NewAuthHandler(userRepo *postgres.UserRepo, sessionRepo *postgres.SessionRepo, cm Disconnector, cache session.CacheRepository) *AuthHandler {
	return &AuthHandler{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		ConnManager: cm,
		Cache:       cache,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Name     string `json:"name"`
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

	req.Name = strings.TrimSpace(req.Name)
	userID, err := h.UserRepo.CreateUser(req.Username, req.Name, hashedPwd, req.Email, "")
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
		"token": token,
		"user": map[string]interface{}{
			"id":       userID,
			"username": req.Username,
			"name":     req.Name,
			"email":    req.Email,
			"rating":   1000,
			"wins":     0,
			"losses":   0,
			"draws":    0,
		},
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

	// Deactivate old sessions
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
		"token": token,
		"user":  user.UserResponse(),
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

	// 1. Get Token (needed for response)
	token, err := httputil.GetTokenFromRequest(r)
	if err != nil {
		log.Printf("[AUTH] /me: Failed to get token for user %d: %v", userID, err)
		http.Error(w, "Token not found", http.StatusUnauthorized)
		return
	}

	// 2. Try Cache
	if h.Cache != nil {
		cacheKey := fmt.Sprintf("user_profile:%d", userID)
		cachedData, err := h.Cache.Get(r.Context(), cacheKey)
		if err == nil && cachedData != "" {
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
				// Inject current token
				response["token"] = token
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Cache", "HIT")
				json.NewEncoder(w).Encode(response)
				return
			}
		}
	}

	// 3. Fallback to Database
	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil || user == nil {
		log.Printf("[AUTH] /me: GetUserByID failed for user %d: err=%v, user=%v", userID, err, user)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	log.Printf("[AUTH] /me: Returning token for user %d, token length: %d", userID, len(token))

	response := user.UserResponse()
	
	// 4. Update Cache
	if h.Cache != nil {
		cacheKey := fmt.Sprintf("user_profile:%d", userID)
		// Cache only user data, without token
		if data, err := json.Marshal(response); err == nil {
			// Cache for 1 hour
			h.Cache.Set(r.Context(), cacheKey, data, time.Hour)
		}
	}

	// 5. Return Response
	response["token"] = token
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) > 100 {
		http.Error(w, "Name must be at most 100 characters", http.StatusBadRequest)
		return
	}

	if err := h.UserRepo.UpdateProfile(userID, req.Name); err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Invalidate cache on update
	if h.Cache != nil {
        h.Cache.Del(r.Context(), fmt.Sprintf("user_profile:%d", userID))
    }

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": user.UserResponse(),
	})
}

func (h *AuthHandler) Leaderboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.UserRepo.GetLeaderboard()
	if err != nil {
		http.Error(w, "Failed to fetch leaderboard", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(stats)
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
