package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iamasit07/4-in-a-row/backend/internal/config"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// If no origin header (like from curl or same-origin), allow the request
		if origin == "" {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Allow-Credentials", "true")

			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusOK)
				return
			}
			c.Next()
			return
		}

		// Check if origin is in allowed list
		allowed := false
		for _, allowedOrigin := range config.AppConfig.AllowedOrigins {
			if allowedOrigin == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				allowed = true
				break
			}
		}

		// If origin not allowed, log and reject
		if !allowed {
			log.Printf("[CORS] Origin '%s' not in allowed list: %v", origin, config.AppConfig.AllowedOrigins)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Origin not allowed"})
			return
		}

		// Set CORS headers for allowed origins
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		// Handle preflight OPTIONS requests
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
