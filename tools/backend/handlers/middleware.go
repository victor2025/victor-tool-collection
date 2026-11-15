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
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-VTC-Password")
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

// WriteAuth allows a write when the caller either holds a valid login
// session cookie or supplies the admin password for this request.
func WriteAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) 已登录：直接放行
		if token, err := c.Cookie("vtc_session"); err == nil && token != "" {
			var session models.Session
			if db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error == nil {
				c.Set("session_id", session.ID)
				c.Next()
				return
			}
		}

		// 2) 未登录：校验本次请求携带的密码
		pwd := c.GetHeader("X-VTC-Password")
		if pwd == "" {
			pwd = c.PostForm("password")
		}
		if pwd == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "需要密码"})
			return
		}

		var admin models.Admin
		if err := db.First(&admin).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "服务异常"})
			return
		}
		if pwd != admin.Password {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "密码错误"})
			return
		}
		c.Next()
	}
}
