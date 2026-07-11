package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"victor-tool-collection/backend/models"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// AuthRequired checks the session cookie against the database.
func AuthRequired(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("vtc_session")
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "未登录"})
			return
		}

		var session models.Session
		if db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "登录已过期"})
			return
		}

		c.Set("session_id", session.ID)
		c.Next()
	}
}
