package models

import "time"

type Visit struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	IP        string    `gorm:"size:64;not null;index" json:"ip"`
	Tool      string    `gorm:"size:64;not null;index" json:"tool"`
	UserAgent string    `gorm:"size:512;default:''" json:"user_agent"`
	VisitedAt time.Time `gorm:"autoCreateTime;index" json:"visited_at"`
}

func (Visit) TableName() string { return "visits" }

type Admin struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Password  string    `gorm:"size:128;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Admin) TableName() string { return "admins" }

type Session struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"size:128;not null;uniqueIndex" json:"token"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (Session) TableName() string { return "sessions" }
