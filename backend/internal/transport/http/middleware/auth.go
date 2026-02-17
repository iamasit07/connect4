package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iamasit07/4-in-a-row/backend/pkg/auth"
	"github.com/iamasit07/4-in-a-row/backend/pkg/httputil"
)

// SessionValidator abstracts session validation (implemented by session.AuthService)
type SessionValidator interface {
	ValidateToken(tokenString string) (*auth.Claims, error)
	UpdateSessionActivity(sessionID string) error
}

// AuthMiddleware validates JWT token and session via the AuthService (Redis → Postgres fallback)
func AuthMiddleware(sv SessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract Token (Cookie or Header)
		tokenString, err := httputil.GetTokenFromRequest(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// 2. Validate JWT + Session (checks signature, is_active, and expires_at)
		claims, err := sv.ValidateToken(tokenString)
		if err != nil {
			log.Printf("[AUTH] Token/session validation failed for %s: %v", c.Request.URL.Path, err)
			httputil.ClearAuthCookie(c.Writer)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		log.Printf("[AUTH] Success for %s (User: %d)", c.Request.URL.Path, claims.UserID)

		// 3. Update Last Activity (async, best-effort)
		go func(sid string) {
			if err := sv.UpdateSessionActivity(sid); err != nil {
				log.Printf("[AUTH] Failed to update session activity for %s: %v", sid, err)
			}
		}(claims.SessionID)

		// 4. Pass UserID and username to next handler via Gin context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("session_id", claims.SessionID)
		c.Set("session_expiry", time.Now())
		c.Next()
	}
}
