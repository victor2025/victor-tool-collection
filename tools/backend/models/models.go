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

// Score is one piece of sheet music. Its pages live in ScoreImage rows,
// either mirrored local files ("local") or remote URLs ("remote").
type Score struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Title     string     `gorm:"size:255;not null;index" json:"title"`
	Composer  string     `gorm:"size:128;default:''" json:"composer"`
	Type      string     `gorm:"size:32;default:'吉他谱';index" json:"type"`
	Note      string     `gorm:"size:512;default:''" json:"note"`
	CreatedAt time.Time  `gorm:"autoCreateTime;index" json:"created_at"`
	Images    []ScoreImage `gorm:"foreignKey:ScoreID" json:"images"`
}

func (Score) TableName() string { return "scores" }

// ScoreImage is one page of a score.
// URL is what the browser loads: "/scores/files/x.png" (local) or an http(s) URL.
type ScoreImage struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ScoreID    uint      `gorm:"not null;index" json:"score_id"`
	URL        string    `gorm:"size:512;not null" json:"url"`
	Source     string    `gorm:"size:16;default:'local'" json:"source"`
	SourceURL  string    `gorm:"size:512;default:''" json:"source_url"`
	StoredName string    `gorm:"size:128;default:''" json:"-"`
	SortOrder  int       `gorm:"default:0" json:"sort_order"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ScoreImage) TableName() string { return "score_images" }
