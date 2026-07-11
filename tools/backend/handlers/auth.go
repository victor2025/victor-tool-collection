package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"victor-tool-collection/backend/models"
)

type AuthHandler struct {
	DB *gorm.DB
}

type LoginRequest struct {
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// Login verifies password from DB, creates session, sets cookie.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "参数错误"})
		return
	}

	var admin models.Admin
	if err := h.DB.First(&admin).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "服务异常"})
		return
	}

	if req.Password != admin.Password {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "密码错误"})
		return
	}

	// Create session in DB
	token := randomToken()
	session := models.Session{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := h.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "创建会话失败"})
		return
	}

	// Set cookie
	c.SetCookie("vtc_session", token, 86400, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// CheckSession reads cookie, verifies session in DB.
func (h *AuthHandler) CheckSession(c *gin.Context) {
	token, err := c.Cookie("vtc_session")
	if err != nil || token == "" {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "未登录"})
		return
	}

	var session models.Session
	if h.DB.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "登录已过期"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ChangePassword updates password in DB.
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	token, err := c.Cookie("vtc_session")
	if err != nil || token == "" {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "未登录"})
		return
	}

	var session models.Session
	if h.DB.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "登录已过期"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "参数错误"})
		return
	}

	var admin models.Admin
	if err := h.DB.First(&admin).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "服务异常"})
		return
	}

	if req.OldPassword != admin.Password {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "当前密码错误"})
		return
	}

	if len(req.NewPassword) < 4 {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "新密码至少4位"})
		return
	}

	h.DB.Model(&admin).Update("password", req.NewPassword)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Logout deletes the session from DB and clears the cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	token, err := c.Cookie("vtc_session")
	if err == nil && token != "" {
		h.DB.Where("token = ?", token).Delete(&models.Session{})
	}
	c.SetCookie("vtc_session", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func randomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
