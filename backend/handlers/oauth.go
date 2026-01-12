package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"

	"github.com/iamasit07/4-in-a-row/backend/config"
	"github.com/iamasit07/4-in-a-row/backend/db"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

// HandleGoogleLogin redirects the user to the Google OAuth consent page
func HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if config.GoogleOAuthConfig == nil {
        http.Error(w, "OAuth config not loaded", http.StatusInternalServerError)
        return
    }
    
    // In production, use a random state string and store it in a cookie to prevent CSRF
	url := config.GoogleOAuthConfig.AuthCodeURL("random-state-string", oauth2.AccessTypeOffline)
    log.Printf("[OAUTH] Redirecting to Google Login: %s", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleGoogleCallback handles the callback from Google
func HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
    log.Println("[OAUTH] Callback received")

	state := r.URL.Query().Get("state")
	if state != "random-state-string" {
        log.Printf("[OAUTH] Invalid state: %s", state)
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
        log.Println("[OAUTH] Code not found in callback URL")
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

    log.Println("[OAUTH] Exchanging code for token...")
	token, err := config.GoogleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
        log.Printf("[OAUTH] Failed to exchange token: %v", err)
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

    log.Println("[OAUTH] Fetching user info from Google...")
	client := config.GoogleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
        log.Printf("[OAUTH] Failed to get user info: %v", err)
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

    log.Println("[OAUTH] Reading user info body...")
	userData, err := io.ReadAll(resp.Body)
	if err != nil {
        log.Printf("[OAUTH] Failed to read user data: %v", err)
		http.Error(w, "Failed to read user data", http.StatusInternalServerError)
		return
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.Unmarshal(userData, &googleUser); err != nil {
        log.Printf("[OAUTH] Failed to parse user data: %v", err)
		http.Error(w, "Failed to parse user data", http.StatusInternalServerError)
		return
	}
    log.Printf("[OAUTH] Fetched Google User: %s (Email: %s)", googleUser.ID, googleUser.Email)

	// Check if user exists
    log.Println("[OAUTH] Checking if user exists by Google ID...")
	user, err := db.GetUserByGoogleID(googleUser.ID)
	if err != nil {
        log.Printf("[OAUTH] Database error checking Google ID: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if user != nil {
		// User exists, log them in
        log.Printf("[OAUTH] User found by Google ID: %s. Logging in...", user.Username)

		// Invalidate old sessions
		err = utils.InvalidateAllUserSessions(user.ID)
		if err != nil {
			log.Printf("[OAUTH] Warning: Failed to invalidate old sessions: %v", err)
		}

		// Generate session
		sessionID, err := utils.GenerateSessionID()
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		deviceInfo := utils.ExtractDeviceInfo(r)
		ipAddress := utils.ExtractIPAddress(r)
		expiration := time.Now().Add(720 * time.Hour)

		session := &models.UserSession{
			UserID:     user.ID,
			SessionID:  sessionID,
			DeviceInfo: deviceInfo,
			IPAddress:  ipAddress,
			ExpiresAt:  expiration,
			IsActive:   true,
		}
		err = utils.SetSession(session)
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		jwtToken, err := utils.GenerateJWT(user.ID, user.Username, sessionID)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		utils.SetAuthCookie(w, jwtToken)
        
        // Redirect to frontend (dashboard)
        frontendURL := config.GetEnv("FRONTEND_URL", "http://localhost:5173")
        log.Printf("[OAUTH] Redirecting to: %s", frontendURL)
		http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect) 
        return
	}

	// User does not exist by GoogleID, check by Email for account linking
    log.Println("[OAUTH] User not found by Google ID. Checking by Email...")
    existingUserByEmail, err := db.GetUserByEmail(googleUser.Email)
    if err != nil {
        log.Printf("[OAUTH] Error checking user by email: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

    if existingUserByEmail != nil {
        // Link accounts
        log.Printf("[OAUTH] User found by Email: %s. Linking Google Account...", existingUserByEmail.Username)
        err = db.UpdateUserGoogleID(googleUser.Email, googleUser.ID)
        if err != nil {
            log.Printf("[OAUTH] Failed to link accounts: %v", err)
            http.Error(w, "Failed to link accounts", http.StatusInternalServerError)
            return
        }
        
        // Log them in
        log.Println("[OAUTH] Account linked. Logging in...")

		// Invalidate old sessions
		err = utils.InvalidateAllUserSessions(existingUserByEmail.ID)
		if err != nil {
			log.Printf("[OAUTH] Warning: Failed to invalidate old sessions: %v", err)
		}

		// Generate session
		sessionID, err := utils.GenerateSessionID()
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		deviceInfo := utils.ExtractDeviceInfo(r)
		ipAddress := utils.ExtractIPAddress(r)
		expiration := time.Now().Add(720 * time.Hour)

		session := &models.UserSession{
			UserID:     existingUserByEmail.ID,
			SessionID:  sessionID,
			DeviceInfo: deviceInfo,
			IPAddress:  ipAddress,
			ExpiresAt:  expiration,
			IsActive:   true,
		}
		err = utils.SetSession(session)
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

        jwtToken, err := utils.GenerateJWT(existingUserByEmail.ID, existingUserByEmail.Username, sessionID)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		utils.SetAuthCookie(w, jwtToken)
        frontendURL := config.GetEnv("FRONTEND_URL", "http://localhost:5173")
        log.Printf("[OAUTH] Redirecting to: %s", frontendURL)
		http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect) 
        return
    }

	// User does not exist, redirect to complete signup
    log.Println("[OAUTH] New user. Redirecting to complete signup...")
	setupToken, err := utils.GenerateSetupToken(googleUser.Email, googleUser.ID)
	if err != nil {
		http.Error(w, "Failed to generate setup token", http.StatusInternalServerError)
		return
	}
    
	// Redirect to frontend complete signup page with setup token
    frontendURL := config.GetEnv("FRONTEND_URL", "http://localhost:5173")
	http.Redirect(w, r, fmt.Sprintf("%s/complete-signup?token=%s", frontendURL, setupToken), http.StatusTemporaryRedirect)
}

type CompleteSignupRequest struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// HandleCompleteGoogleSignup completes the signup process
func HandleCompleteGoogleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CompleteSignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate setup token
	claims, err := utils.ValidateSetupToken(req.Token)
	if err != nil {
		writeJSONError(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Validate inputs
	req.Username = strings.TrimSpace(req.Username)
	if len(req.Username) < 3 || len(req.Username) > 50 {
		writeJSONError(w, "Username must be between 3 and 50 characters", http.StatusBadRequest)
		return
	}
    if strings.ToUpper(req.Username) == "BOT" {
		writeJSONError(w, "Username 'BOT' is reserved", http.StatusBadRequest)
		return
	}
	if len(req.Password) < 6 {
		writeJSONError(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

    // Check if username taken
    existingUser, err := db.GetUserByUsername(req.Username)
    if err != nil {
        writeJSONError(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    if existingUser != nil {
        writeJSONError(w, "Username already exists", http.StatusConflict)
        return
    }

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

    // Create user with Google ID and Email
    // Using CreateUser with all fields
	userID, err := db.CreateUser(req.Username, string(passwordHash), claims.Email, claims.GoogleID)
	if err != nil {
		log.Printf("[AUTH] Error creating user: %v", err)
		writeJSONError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate session
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	deviceInfo := utils.ExtractDeviceInfo(r)
	ipAddress := utils.ExtractIPAddress(r)
	expiration := time.Now().Add(720 * time.Hour)

	session := &models.UserSession{
		UserID:     userID,
		SessionID:  sessionID,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		ExpiresAt:  expiration,
		IsActive:   true,
	}
	err = utils.SetSession(session)
	if err != nil {
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate JWT
	jwtToken, err := utils.GenerateJWT(userID, req.Username, sessionID)
	if err != nil {
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	utils.SetAuthCookie(w, jwtToken)

	writeJSON(w, AuthResponse{
		Token:    jwtToken,
		UserID:   userID,
		Username: req.Username,
	}, http.StatusCreated)
}
