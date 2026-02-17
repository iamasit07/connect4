package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iamasit07/4-in-a-row/backend/internal/repository/postgres"
	"github.com/iamasit07/4-in-a-row/backend/internal/service/session"
	"github.com/iamasit07/4-in-a-row/backend/pkg/auth"
	"github.com/iamasit07/4-in-a-row/backend/pkg/httputil"
	"github.com/iamasit07/4-in-a-row/backend/pkg/useragent"
)

type Disconnector interface {
	DisconnectUser(userID int64, reason string)
}

type SessionInvalidator interface {
	InvalidateSession(sessionID string) error
}

type AuthHandler struct {
	UserRepo    *postgres.UserRepo
	SessionRepo *postgres.SessionRepo
	ConnManager Disconnector
	Cache       session.CacheRepository
	AuthService SessionInvalidator
}

func NewAuthHandler(userRepo *postgres.UserRepo, sessionRepo *postgres.SessionRepo, cm Disconnector, cache session.CacheRepository, authSvc SessionInvalidator) *AuthHandler {
	return &AuthHandler{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		ConnManager: cm,
		Cache:       cache,
		AuthService: authSvc,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	if len(req.Username) < 3 || len(req.Username) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be between 3 and 50 characters"})
		return
	}

	if strings.ToUpper(req.Username) == "BOT" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username 'BOT' is reserved"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 6 characters"})
		return
	}

	existing, _ := h.UserRepo.GetUserByIdentifier(req.Username)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username or email already taken"})
		return
	}

	hashedPwd, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	userID, err := h.UserRepo.CreateUser(req.Username, req.Name, hashedPwd, req.Email, "", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Create Session
	sessionID := auth.GenerateToken()
	deviceInfo := useragent.ExtractDeviceInfo(c.Request)
	ipAddress := useragent.ExtractIPAddress(c.Request)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	err = h.SessionRepo.CreateSession(userID, sessionID, deviceInfo, ipAddress, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	token, err := auth.GenerateJWT(userID, req.Username, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	httputil.SetAuthCookie(c.Writer, token)
	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user": gin.H{
			"id":         userID,
			"username":   req.Username,
			"name":       req.Name,
			"avatar_url": "",
			"email":      req.Email,
			"rating":     1000,
			"wins":       0,
			"losses":     0,
			"draws":      0,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	user, err := h.UserRepo.GetUserByIdentifier(req.Username)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
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
	deviceInfo := useragent.ExtractDeviceInfo(c.Request)
	ipAddress := useragent.ExtractIPAddress(c.Request)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	err = h.SessionRepo.CreateSession(user.ID, sessionID, deviceInfo, ipAddress, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	token, err := auth.GenerateJWT(user.ID, user.Username, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	httputil.SetAuthCookie(c.Writer, token)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user.UserResponse(),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Invalidate session server-side (DB + Redis cache)
	sessionID, exists := c.Get("session_id")
	if exists {
		if sid, ok := sessionID.(string); ok && sid != "" {
			if err := h.AuthService.InvalidateSession(sid); err != nil {
				log.Printf("[AUTH] Failed to invalidate session %s on logout: %v", sid, err)
			}
		}
	}

	httputil.ClearAuthCookie(c.Writer)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 1. Get Token (needed for response)
	token, err := httputil.GetTokenFromRequest(c.Request)
	if err != nil {
		log.Printf("[AUTH] /me: Failed to get token for user %d: %v", userID, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not found"})
		return
	}

	// 2. Try Cache
	if h.Cache != nil {
		cacheKey := fmt.Sprintf("user_profile:%d", userID)
		cachedData, err := h.Cache.Get(c.Request.Context(), cacheKey)
		if err == nil && cachedData != "" {
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(cachedData), &response); err == nil {
				// Inject current token
				response["token"] = token
				c.Header("X-Cache", "HIT")
				c.JSON(http.StatusOK, response)
				return
			}
		}
	}

	// 3. Fallback to Database
	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil || user == nil {
		log.Printf("[AUTH] /me: GetUserByID failed for user %d: err=%v, user=%v", userID, err, user)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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
			h.Cache.Set(c.Request.Context(), cacheKey, data, time.Hour)
		}
	}

	// 5. Return Response
	response["token"] = token
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name must be at most 100 characters"})
		return
	}

	if err := h.UserRepo.UpdateProfile(userID, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	user, err := h.UserRepo.GetUserByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Invalidate cache on update
	if h.Cache != nil {
		h.Cache.Del(c.Request.Context(), fmt.Sprintf("user_profile:%d", userID))
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user.UserResponse(),
	})
}

const maxAvatarSize = 2 * 1024 * 1024 // 2MB

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Limit request body size
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAvatarSize)

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided or file too large (max 2MB)"})
		return
	}
	defer file.Close()

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: JPEG, PNG, WebP"})
		return
	}

	// Create uploads directory
	uploadDir := "./uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("[AVATAR] Failed to create upload dir: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, filename)

	// Save file
	dst, err := os.Create(savePath)
	if err != nil {
		log.Printf("[AVATAR] Failed to create file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("[AVATAR] Failed to save file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Delete old avatar file if it's a local upload
	oldUser, _ := h.UserRepo.GetUserByID(userID)
	if oldUser != nil && oldUser.AvatarURL != "" && strings.HasPrefix(oldUser.AvatarURL, "/uploads/") {
		oldPath := "." + oldUser.AvatarURL
		os.Remove(oldPath)
	}

	// Save URL to database
	avatarURL := "/uploads/avatars/" + filename
	if err := h.UserRepo.UpdateAvatar(userID, avatarURL); err != nil {
		log.Printf("[AVATAR] Failed to update avatar in DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar"})
		return
	}

	// Invalidate profile cache
	if h.Cache != nil {
		h.Cache.Del(c.Request.Context(), fmt.Sprintf("user_profile:%d", userID))
	}

	c.JSON(http.StatusOK, gin.H{
		"avatar_url": avatarURL,
	})
}

func (h *AuthHandler) RemoveAvatar(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Delete old avatar file if local
	user, _ := h.UserRepo.GetUserByID(userID)
	if user != nil && user.AvatarURL != "" && strings.HasPrefix(user.AvatarURL, "/uploads/") {
		os.Remove("." + user.AvatarURL)
	}

	if err := h.UserRepo.UpdateAvatar(userID, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove avatar"})
		return
	}

	// Invalidate profile cache
	if h.Cache != nil {
		h.Cache.Del(c.Request.Context(), fmt.Sprintf("user_profile:%d", userID))
	}

	c.JSON(http.StatusOK, gin.H{
		"avatar_url": "",
	})
}

func (h *AuthHandler) Leaderboard(c *gin.Context) {
	stats, err := h.UserRepo.GetLeaderboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *AuthHandler) GetSessionHistory(c *gin.Context) {
	// 1. Get UserID from context (set by AuthMiddleware)
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	// 2. Fetch Sessions
	sessions, err := h.SessionRepo.GetUserSessionHistory(userID, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch session history"})
		return
	}

	// 3. Return JSON
	c.JSON(http.StatusOK, sessions)
}
