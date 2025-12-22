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

	// Send response
	writeJSON(w, AuthResponse{
		Token:    token,
		UserID:   userID,
		Username: req.Username,
	}, http.StatusCreated)
}

// HandleLogin handles user login
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		writeJSONError(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := db.GetUserByUsername(req.Username)
	if err != nil {
		log.Printf("[AUTH] Error getting user: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		writeJSONError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		writeJSONError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	token, err := utils.GenerateJWT(user.ID, user.Username)
	if err != nil {
		log.Printf("[AUTH] Error generating JWT: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("[AUTH] User logged in successfully: %s (ID: %d)", user.Username, user.ID)

	// Send response
	writeJSON(w, AuthResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
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
