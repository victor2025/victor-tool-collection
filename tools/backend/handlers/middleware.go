package handlers

import (
	"net/http"

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

// 安全响应头统一由 nginx 注入（deploy/nginx/vtc-security-headers.conf）。
// 后端只监听 127.0.0.1，这里不再重复设置，避免响应中出现重复的安全头。

// AuthRequired checks the session cookie against the database.
func AuthRequired(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session, ok := sessionFromCookie(db, c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "未登录"})
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
		if session, ok := sessionFromCookie(db, c); ok {
			c.Set("session_id", session.ID)
			c.Next()
			return
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
		if pass, needsUpgrade := verifyPassword(admin.Password, pwd); !pass {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "error": "密码错误"})
			return
		} else if needsUpgrade {
			// 历史明文密码：校验通过后顺手升级
			if hashed, err := HashPassword(pwd); err == nil {
				db.Model(&admin).Update("password", hashed)
			}
		}
		c.Next()
	}
}
