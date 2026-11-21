package handlers

import (
	"crypto/subtle"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"victor-tool-collection/backend/models"
)

// bcryptCost 是密码哈希的代价因子。树莓派上 10 约 30~60ms，管理后台登录够用。
const bcryptCost = 10

// IsBcryptHash 判断存储值是否为 bcrypt 哈希（$2a$ / $2b$ / $2y$ 前缀）。
func IsBcryptHash(s string) bool {
	return strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$") || strings.HasPrefix(s, "$2y$")
}

// HashPassword 生成 bcrypt 哈希。
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// verifyPassword 校验明文密码。stored 既可能是 bcrypt 哈希，也可能是历史遗留的明文。
// 返回是否通过，以及是否需要在本次成功后升级存储形态。
func verifyPassword(stored, plain string) (ok bool, needsUpgrade bool) {
	if stored == "" || plain == "" {
		return false, false
	}
	if IsBcryptHash(stored) {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(plain)) == nil, false
	}
	// 历史明文：常量时间比较，避免时序侧信道
	if subtle.ConstantTimeCompare([]byte(stored), []byte(plain)) == 1 {
		return true, true
	}
	return false, false
}

// MigrateAdminPassword 在启动时把历史明文密码一次性升级为 bcrypt 哈希。
func MigrateAdminPassword(db *gorm.DB) error {
	var admin models.Admin
	if err := db.First(&admin).Error; err != nil {
		return err
	}
	if IsBcryptHash(admin.Password) {
		return nil
	}
	h, err := HashPassword(admin.Password)
	if err != nil {
		return err
	}
	if err := db.Model(&admin).Update("password", h).Error; err != nil {
		return err
	}
	log.Println("admin 密码已升级为 bcrypt 哈希")
	return nil
}
