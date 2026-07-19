package models

import "time"

type Visit struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	IP        string    `gorm:"size:64;not null;index" json:"ip"`
	Tool      string    `gorm:"size:64;not null;index" json:"tool"`
	UserAgent string    `gorm:"size:512;default:''" json:"user_agent"`
	DeviceID  string    `gorm:"size:64;default:'';index" json:"device_id"`
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

// DeviceLabel stores human-readable labels for device identifiers.
type DeviceLabel struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DeviceID  string    `gorm:"size:64;not null;uniqueIndex" json:"device_id"`
	Label     string    `gorm:"size:128;not null" json:"label"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (DeviceLabel) TableName() string { return "device_labels" }
