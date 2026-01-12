package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/iamasit07/4-in-a-row/backend/db"
	"github.com/iamasit07/4-in-a-row/backend/models"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Identifier string `json:"username"` // Can be username or email. Mapped from "username" json field to avoid frontend changes if possible, or we change frontend too.
    // Let's decide: User said "login with username or email". Frontend login.tsx uses "username" state. api.ts sends "username".
    // So if I keep "username" json tag, I don't break frontend API contract immediately. I will assume "username" field can contain email.
	Password   string `json:"password"`
}

type AuthResponse struct {
	Token    string `json:"token"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// HandleSignup handles user registration
func HandleSignup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate username
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		writeJSONError(w, "Username is required", http.StatusBadRequest)
		return
	}
	if len(req.Username) < 3 || len(req.Username) > 50 {
		writeJSONError(w, "Username must be between 3 and 50 characters", http.StatusBadRequest)
		return
	}
	if strings.ToUpper(req.Username) == "BOT" {
		writeJSONError(w, "Username 'BOT' is reserved", http.StatusBadRequest)
		return
	}

    // Validate email
    req.Email = strings.TrimSpace(req.Email)
    if req.Email == "" {
        writeJSONError(w, "Email is required", http.StatusBadRequest)
        return
    }
    // Simple email regex or check
    if !strings.Contains(req.Email, "@") {
         writeJSONError(w, "Invalid email format", http.StatusBadRequest)
         return
    }

	// Validate password
	if len(req.Password) < 6 {
		writeJSONError(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	// Check if username already exists
	existingUser, err := db.GetUserByUsername(req.Username)
	if err != nil {
		log.Printf("[AUTH] Error checking existing user: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if existingUser != nil {
		writeJSONError(w, "Username already exists", http.StatusConflict)
		return
	}
    
    // Check if email already exists
    existingEmail, err := db.GetUserByEmail(req.Email)
    if err != nil {
		log.Printf("[AUTH] Error checking existing email: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
    if existingEmail != nil {
        writeJSONError(w, "Email already exists", http.StatusConflict)
		return
    }

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[AUTH] Error hashing password: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	userID, err := db.CreateUser(req.Username, string(passwordHash), req.Email, "")
	if err != nil {
		log.Printf("[AUTH] Error creating user: %v", err)
		writeJSONError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate session ID
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		log.Printf("[AUTH] Error generating session ID: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Extract device info and IP
	deviceInfo := utils.ExtractDeviceInfo(r)
	ipAddress := utils.ExtractIPAddress(r)

	// Create session (stores in Redis + PostgreSQL)
	expiration := time.Now().Add(720 * time.Hour) // 30 days
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
		log.Printf("[AUTH] Error creating session: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate JWT with session ID
	token, err := utils.GenerateJWT(userID, req.Username, sessionID)
	if err != nil {
		log.Printf("[AUTH] Error generating JWT: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("[AUTH] User registered successfully: %s (ID: %d)", req.Username, userID)

	// Set JWT in HTTP-only cookie
	utils.SetAuthCookie(w, token)

	// Send response (token included for hybrid approach - frontend can read from response)
	writeJSON(w, AuthResponse{
		Token:    token,
		UserID:   userID,
		Username: req.Username,
	}, http.StatusCreated)
}

// HandleLogin handles user login
func MakeHandleLogin(connManager interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HandleLoginWithConnManager(w, r, connManager)
	}
}

func HandleLoginWithConnManager(w http.ResponseWriter, r *http.Request, connManager interface{}) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("[AUTH] Login attempt for identifier: %s", req.Identifier)

	if strings.TrimSpace(req.Identifier) == "" || strings.TrimSpace(req.Password) == "" {
		writeJSONError(w, "Username/Email and password are required", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByIdentifier(req.Identifier)
	if err != nil {
		writeJSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
    
    // Check if user found
    if user == nil {
        writeJSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
    }

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Invalidate all existing sessions for single-device enforcement
	err = utils.InvalidateAllUserSessions(user.ID)
	if err != nil {
		log.Printf("[AUTH] Warning: Failed to invalidate old sessions: %v", err)
	}

	// Generate new session ID
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		log.Printf("[AUTH] Login failed for user %s: Session ID generation error - %v", user.Username, err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Extract device info and IP
	deviceInfo := utils.ExtractDeviceInfo(r)
	ipAddress := utils.ExtractIPAddress(r)

	// Create session (stores in Redis + PostgreSQL)
	expiration := time.Now().Add(720 * time.Hour) // 30 days
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
		log.Printf("[AUTH] Login failed for user %s: Session creation error - %v", user.Username, err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate JWT with session ID
	token, err := utils.GenerateJWT(user.ID, user.Username, sessionID)
	if err != nil {
		log.Printf("[AUTH] Login failed for user %s: JWT generation error - %v", user.Username, err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("[AUTH] User logged in successfully: %s (ID: %d)", user.Username, user.ID)

	// Disconnect any existing WebSocket connections for this user
	if cm, ok := connManager.(interface{ DisconnectUser(int64, string) }); ok {
		cm.DisconnectUser(user.ID, "Logged in from another device")
		log.Printf("[AUTH] Disconnected existing session for user %d during login", user.ID)
	}

	// Set JWT in HTTP-only cookie
	utils.SetAuthCookie(w, token)

	// Send response (token included for hybrid approach - frontend can read from response)
	writeJSON(w, AuthResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
	}, http.StatusOK)
}

// HandleLogout handles user logout by clearing the auth cookie
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear the auth cookie
	utils.ClearAuthCookie(w)

	log.Printf("[AUTH] User logged out successfully")

	// Send success response
	writeJSON(w, map[string]string{"message": "Logged out successfully"}, http.StatusOK)
}

// HandleMe returns current user info based on the auth cookie
func HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get token from cookie
	token, err := utils.GetTokenFromCookie(r)
	if err != nil {
		writeJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Validate JWT with session check
	claims, err := utils.ValidateJWTWithSession(token)
	if err != nil {
		if err.Error() == "session invalidated" || err.Error() == "session expired" {
			writeJSONError(w, "Session invalidated or expired", http.StatusUnauthorized)
		} else {
			writeJSONError(w, "Invalid or expired token", http.StatusUnauthorized)
		}
		return
	}

	// Return user info and token (for WebSocket use)
	writeJSON(w, AuthResponse{
		Token:    token,
		UserID:   claims.UserID,
		Username: claims.Username,
	}, http.StatusOK)
}

// Helper functions
func writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, ErrorResponse{Error: message}, status)
}
