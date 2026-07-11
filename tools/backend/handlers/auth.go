package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"victor-tool-collection/backend/models"
)

type AuthHandler struct {
	DB       *gorm.DB
	Password string
}

type LoginRequest struct {
	Password string `json:"password"`
}

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

	// Plain text comparison
	if req.Password != admin.Password {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "密码错误"})
		return
	}

	token := randomToken()
	AddToken(token)
	c.JSON(http.StatusOK, gin.H{"ok": true, "token": token})
}

type CheckSessionRequest struct {
	Token string `json:"token"`
}

func (h *AuthHandler) CheckSession(c *gin.Context) {
	var req CheckSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "参数错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": IsValidToken(req.Token)})
}

type ChangePasswordRequest struct {
	Token       string `json:"token"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": "参数错误"})
		return
	}

	if !IsValidToken(req.Token) {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "未登录"})
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

func randomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
