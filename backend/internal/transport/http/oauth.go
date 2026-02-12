package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/internal/config"
	"github.com/iamasit07/4-in-a-row/backend/internal/repository/postgres"
	"github.com/iamasit07/4-in-a-row/backend/pkg/auth"
	"github.com/iamasit07/4-in-a-row/backend/pkg/httputil"
	"github.com/iamasit07/4-in-a-row/backend/pkg/useragent"
)

type OAuthHandler struct {
	UserRepo    *postgres.UserRepo
	SessionRepo *postgres.SessionRepo
	Config      *config.OAuthConfig
	ConnManager Disconnector // Reusing the interface from auth.go
}

// NewOAuthHandler now requires SessionRepo and Disconnector (ConnManager)
func NewOAuthHandler(userRepo *postgres.UserRepo, sessionRepo *postgres.SessionRepo, cfg *config.OAuthConfig, cm Disconnector) *OAuthHandler {
	return &OAuthHandler{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		Config:      cfg,
		ConnManager: cm,
	}
}

// GoogleLogin redirects the user to Google
func (h *OAuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := h.Config.GoogleLoginConfig.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the response from Google
func (h *OAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	token, err := h.Config.GoogleLoginConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("[OAUTH] Failed to exchange token: %v", err)
		http.Redirect(w, r, config.AppConfig.FrontendURL+"/login?error=auth_failed", http.StatusTemporaryRedirect)
		return
	}

	userInfo, err := config.GetGoogleUserInfo(token.AccessToken)
	if err != nil {
		log.Printf("[OAUTH] Failed to get user info: %v", err)
		http.Redirect(w, r, config.AppConfig.FrontendURL+"/login?error=user_info_failed", http.StatusTemporaryRedirect)
		return
	}

	user, err := h.UserRepo.GetUserByEmail(userInfo.Email)

	if user != nil {
		// --- CASE A: EXISTING USER (LOGIN) ---

		// Auto-link Google ID if missing (The logic you liked)
		if !user.GoogleID.Valid {
			if err := h.UserRepo.UpdateUserGoogleID(userInfo.Email, userInfo.ID); err != nil {
				log.Printf("[OAUTH] Failed to link Google ID: %v", err)
			}
		}

		// Security: Invalidate old sessions
		log.Printf("[OAUTH] Deactivating all sessions for user %d (%s)", user.ID, user.Username)
		h.SessionRepo.DeactivateAllUserSessions(user.ID)
		if h.ConnManager != nil {
			h.ConnManager.DisconnectUser(user.ID, "Logged in from another device via Google")
		}

		// Create new session
		sessionID := auth.GenerateToken()
		deviceInfo := useragent.ExtractDeviceInfo(r)
		ipAddress := useragent.ExtractIPAddress(r)
		expiresAt := time.Now().Add(30 * 24 * time.Hour)

		log.Printf("[OAUTH] Creating new session for user %d: sessionID=%s", user.ID, sessionID)
		err = h.SessionRepo.CreateSession(user.ID, sessionID, deviceInfo, ipAddress, expiresAt)
		if err != nil {
			log.Printf("[OAUTH] Failed to create session: %v", err)
			http.Redirect(w, r, config.AppConfig.FrontendURL+"/login?error=server_error", http.StatusTemporaryRedirect)
			return
		}
		log.Printf("[OAUTH] Session created successfully for user %d", user.ID)

		// Generate Login JWT
		jwtToken, err := auth.GenerateJWT(user.ID, user.Username, sessionID)
		if err != nil {
			http.Redirect(w, r, config.AppConfig.FrontendURL+"/login?error=token_error", http.StatusTemporaryRedirect)
			return
		}

		// Set cookie and redirect to dashboard
		httputil.SetAuthCookie(w, jwtToken)
		http.Redirect(w, r, config.AppConfig.FrontendURL+"/dashboard", http.StatusTemporaryRedirect)

	} else {
		// --- CASE B: NEW USER (SETUP FLOW) ---

		// Do NOT create user yet. Generate a Setup Token instead.
		setupToken, err := auth.GenerateSetupToken(userInfo.Email, userInfo.ID)
		if err != nil {
			log.Printf("[OAUTH] Failed to generate setup token: %v", err)
			http.Redirect(w, r, config.AppConfig.FrontendURL+"/login?error=setup_failed", http.StatusTemporaryRedirect)
			return
		}

		redirectURL := fmt.Sprintf("%s/complete-signup?token=%s&email=%s&name=%s",
			config.AppConfig.FrontendURL,
			setupToken,
			userInfo.Email, userInfo.Name)

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	}
}

// CompleteGoogleSignup processes the final step of Google registration
func (h *OAuthHandler) CompleteGoogleSignup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SetupToken string `json:"token"`
		Username   string `json:"username"`
		Password   string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// 1. Validate Setup Token
	claims, err := auth.ValidateSetupToken(req.SetupToken)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or expired signup token"})
		return
	}

	// 2. Validate Input
	req.Username = strings.TrimSpace(req.Username)
	if len(req.Username) < 3 || len(req.Username) > 50 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Username must be between 3 and 50 characters"})
		return
	}
	if strings.ToUpper(req.Username) == "BOT" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Username 'BOT' is reserved"})
		return
	}
	if len(req.Password) < 6 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Password must be at least 6 characters"})
		return
	}

	// 3. Ensure User Doesn't Exist (Race condition check)
	existing, _ := h.UserRepo.GetUserByIdentifier(req.Username)
	if existing != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Username already taken"})
		return
	}
	emailUser, _ := h.UserRepo.GetUserByEmail(claims.Email)
	if emailUser != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email already registered. Please login."})
		return
	}

	// 4. Create User
	hashedPwd, _ := auth.HashPassword(req.Password)
	userID, err := h.UserRepo.CreateUser(req.Username, hashedPwd, claims.Email, claims.GoogleID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create account"})
		return
	}

	// 5. Create Session
	sessionID := auth.GenerateToken()
	deviceInfo := useragent.ExtractDeviceInfo(r)
	ipAddress := useragent.ExtractIPAddress(r)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	h.SessionRepo.CreateSession(userID, sessionID, deviceInfo, ipAddress, expiresAt)

	// 6. Generate Login Token
	token, _ := auth.GenerateJWT(userID, req.Username, sessionID)

	httputil.SetAuthCookie(w, token)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":    token,
		"username": req.Username,
		"user_id":  userID,
	})
}
