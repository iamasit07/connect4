package handlers

import (
	"net/http"

	"github.com/iamasit07/4-in-a-row/backend/db"
	"github.com/iamasit07/4-in-a-row/backend/utils"
)

// HandleSessionHistory returns the session history for the authenticated user
func HandleSessionHistory(w http.ResponseWriter, r *http.Request) {
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

	// Validate JWT with session
	claims, err := utils.ValidateJWTWithSession(token)
	if err != nil {
		writeJSONError(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	// Get session history (last 10 sessions)
	sessions, err := db.GetUserSessionHistory(claims.UserID, 10)
	if err != nil {
		writeJSONError(w, "Failed to fetch session history", http.StatusInternalServerError)
		return
	}

	writeJSON(w, sessions, http.StatusOK)
}
