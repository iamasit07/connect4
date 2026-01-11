package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/iamasit07/4-in-a-row/backend/db"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

type SignupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
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

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[AUTH] Error hashing password: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	userID, err := db.CreateUser(req.Username, string(passwordHash))
	if err != nil {
		log.Printf("[AUTH] Error creating user: %v", err)
		writeJSONError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate JWT
	token, err := utils.GenerateJWT(userID, req.Username)
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

	log.Printf("[AUTH] Login attempt for username: %s", req.Username)

	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		writeJSONError(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByUsername(req.Username)
	if err != nil {
		writeJSONError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSONError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		log.Printf("[AUTH] Login failed for user %s: JWT generation error - %v", req.Username, err)
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

	// Validate JWT
	claims, err := utils.ValidateJWT(token)
	if err != nil {
		writeJSONError(w, "Invalid or expired token", http.StatusUnauthorized)
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
