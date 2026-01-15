package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/internal/repository/postgres"
	"github.com/iamasit07/4-in-a-row/backend/pkg/auth"
	"github.com/iamasit07/4-in-a-row/backend/pkg/httputil"
)

// AuthMiddleware validates JWT token from cookie
// AuthMiddleware wraps the next handler and validates the JWT + Session Status
func AuthMiddleware(next http.HandlerFunc, sessionRepo *postgres.SessionRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract Token (Cookie or Header)
		tokenString, err := httputil.GetTokenFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 2. Validate JWT Signature (Stateless check)
		claims, err := auth.ValidateJWT(tokenString)
		if err != nil {
			httputil.ClearAuthCookie(w)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// 3. Validate Session in DB (Stateful check - The Logic You Were Missing!)
		// This replaces utils.ValidateJWTWithSession
		session, err := sessionRepo.GetSessionByID(claims.SessionID)
		if err != nil || session == nil {
			httputil.ClearAuthCookie(w)
			http.Error(w, "Session invalid", http.StatusUnauthorized)
			return
		}

		// Check if session is explicitly deactivated (e.g. logout/ban)
		if !session.IsActive {
			httputil.ClearAuthCookie(w)
			http.Error(w, "Session logged out", http.StatusUnauthorized)
			return
		}

		// Check expiry
		if time.Now().After(session.ExpiresAt) {
			httputil.ClearAuthCookie(w)
			http.Error(w, "Session expired", http.StatusUnauthorized)
			return
		}

		// 4. Update Last Activity (Optional, but good for tracking)
		// Run in background to not block request
		go sessionRepo.UpdateSessionActivity(claims.SessionID)

		// 5. Pass UserID to next handler
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
