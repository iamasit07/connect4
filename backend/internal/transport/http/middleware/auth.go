package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/iamasit07/4-in-a-row/backend/pkg/auth"
	"github.com/iamasit07/4-in-a-row/backend/pkg/httputil"
)

// SessionValidator abstracts session validation (implemented by session.AuthService)
type SessionValidator interface {
	ValidateToken(tokenString string) (*auth.Claims, error)
	UpdateSessionActivity(sessionID string) error
}

// AuthMiddleware validates JWT token and session via the AuthService (Redis → Postgres fallback)
func AuthMiddleware(next http.HandlerFunc, sv SessionValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract Token (Cookie or Header)
		tokenString, err := httputil.GetTokenFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 2. Validate JWT + Session (checks signature, is_active, and expires_at)
		claims, err := sv.ValidateToken(tokenString)
		if err != nil {
			log.Printf("[AUTH] Token/session validation failed for %s: %v", r.URL.Path, err)
			httputil.ClearAuthCookie(w)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Printf("[AUTH] Success for %s (User: %d)", r.URL.Path, claims.UserID)

		// 3. Update Last Activity (async, best-effort)
		go func(sid string) {
			if err := sv.UpdateSessionActivity(sid); err != nil {
				log.Printf("[AUTH] Failed to update session activity for %s: %v", sid, err)
			}
		}(claims.SessionID)

		// 4. Pass UserID and username to next handler
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "session_id", claims.SessionID)
		ctx = context.WithValue(ctx, "session_expiry", time.Now())
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
