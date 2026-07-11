package handlers

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// TokenStore holds active tokens in memory with no cleanup goroutine.
// In production, use Redis with TTL.
var (
	activeTokens = make(map[string]bool)
	tokenMu      sync.RWMutex
)

// AddToken records a token so it becomes valid.
func AddToken(token string) {
	tokenMu.Lock()
	activeTokens[token] = true
	tokenMu.Unlock()
}

// RemoveToken invalidates a token.
func RemoveToken(token string) {
	tokenMu.Lock()
	delete(activeTokens, token)
	tokenMu.Unlock()
}

// IsValidToken checks whether a token is currently active.
func IsValidToken(token string) bool {
	tokenMu.RLock()
	defer tokenMu.RUnlock()
	return activeTokens[token]
}

// CORSMiddleware allows cross-origin requests.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AuthMiddleware validates the Bearer token in the Authorization header.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		// Strip "Bearer " prefix
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		if !IsValidToken(token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		c.Set("token", token)
		c.Next()
	}
}
