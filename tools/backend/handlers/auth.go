package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"victor-tool-collection/backend/models"
)

// 登录尝试限流：同一 IP 10 分钟内最多 8 次（防绕过 nginx 直连后端）。
var loginLimiter = newRateLimiter(10*time.Minute, 8)

const sessionCookieName = "vtc_session"

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
	ip := clientIP(c)
	if ok, wait := loginLimiter.Allow(ip); !ok {
		c.Header("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
		c.JSON(http.StatusTooManyRequests, gin.H{"ok": false, "error": "尝试过于频繁，请稍后再试"})
		return
	}

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

	ok, needsUpgrade := verifyPassword(admin.Password, req.Password)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "密码错误"})
		return
	}

	// 历史明文密码：登录成功时顺手升级为 bcrypt
	if needsUpgrade {
		if hashed, err := HashPassword(req.Password); err == nil {
			h.DB.Model(&admin).Update("password", hashed)
		}
	}

	// 清理过期会话，避免 sessions 表只增不减
	h.DB.Where("expires_at < ?", time.Now()).Delete(&models.Session{})

	token := randomToken()
	session := models.Session{Token: token, ExpiresAt: time.Now().Add(24 * time.Hour)}
	if err := h.DB.Create(&session).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "创建会话失败"})
		return
	}

	setSessionCookie(c, token, 86400)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// CheckSession reads cookie, verifies session in DB.
func (h *AuthHandler) CheckSession(c *gin.Context) {
	if _, ok := sessionFromCookie(h.DB, c); !ok {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "未登录"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ChangePassword updates password in DB, then invalidates all other sessions.
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	current, ok := sessionFromCookie(h.DB, c)
	if !ok {
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

	if pass, _ := verifyPassword(admin.Password, req.OldPassword); !pass {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "当前密码错误"})
		return
	}

	// bcrypt 只使用前 72 字节；下限提到 8 位，避免继续使用弱口令
	if len(req.NewPassword) < 8 {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "新密码至少 8 位"})
		return
	}
	if len(req.NewPassword) > 72 {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "新密码过长（最多 72 字节）"})
		return
	}

	hashed, err := HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "密码处理失败"})
		return
	}
	if err := h.DB.Model(&admin).Update("password", hashed).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": false, "error": "保存失败"})
		return
	}

	// 改密后失效其它设备上的会话，只保留当前这一个
	h.DB.Where("id <> ?", current.ID).Delete(&models.Session{})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Logout deletes the session from DB and clears the cookie.
func (h *AuthHandler) Logout(c *gin.Context) {
	if token, err := c.Cookie(sessionCookieName); err == nil && token != "" {
		h.DB.Where("token = ?", token).Delete(&models.Session{})
	}
	setSessionCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// sessionFromCookie 校验 cookie 里的会话是否仍然有效。
func sessionFromCookie(db *gorm.DB, c *gin.Context) (models.Session, bool) {
	token, err := c.Cookie(sessionCookieName)
	if err != nil || token == "" {
		return models.Session{}, false
	}
	var session models.Session
	if db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error != nil {
		return models.Session{}, false
	}
	return session, true
}

// setSessionCookie 统一下发 cookie。仅当请求确实走 HTTPS（含 Cloudflare 回源）时才带 Secure，
// 这样局域网 http://ip:8001 调试不会因 Secure 标记而登录失败。
func setSessionCookie(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, token, maxAge, "/", "", isHTTPS(c), true)
}

func isHTTPS(c *gin.Context) bool {
	if c.Request != nil && c.Request.TLS != nil {
		return true
	}
	proto := c.GetHeader("X-Forwarded-Proto")
	if i := strings.Index(proto, ","); i >= 0 {
		proto = proto[:i]
	}
	return strings.EqualFold(strings.TrimSpace(proto), "https")
}

// clientIP 优先使用 nginx 覆盖写入的 X-Real-IP，避免客户端伪造来源 IP。
func clientIP(c *gin.Context) string {
	if ip := strings.TrimSpace(c.GetHeader("X-Real-IP")); ip != "" {
		return ip
	}
	return c.ClientIP()
}

func randomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Printf("randomToken: %v", err)
	}
	return hex.EncodeToString(b)
}
