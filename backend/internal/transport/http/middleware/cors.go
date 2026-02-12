package middleware

import (
	"log"
	"net/http"

	"github.com/iamasit07/4-in-a-row/backend/internal/config"
)

func EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// If no origin header (like from curl or same-origin), allow the request
		if origin == "" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		
		// Check if origin is in allowed list
		allowed := false
		for _, allowedOrigin := range config.AppConfig.AllowedOrigins {
			if allowedOrigin == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				allowed = true
				break
			}
		}
		
		// If origin not allowed, log and reject
		if !allowed {
			log.Printf("[CORS] Origin '%s' not in allowed list: %v", origin, config.AppConfig.AllowedOrigins)
			http.Error(w, "Origin not allowed", http.StatusForbidden)
			return
		}

		// Set CORS headers for allowed origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
