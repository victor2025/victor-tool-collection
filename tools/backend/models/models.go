package models

import "time"

// Visit records one page visit.
type Visit struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	IP        string    `gorm:"size:64;not null;index" json:"ip"`
	Tool      string    `gorm:"size:64;not null;index" json:"tool"`
	VisitedAt time.Time `gorm:"autoCreateTime;index" json:"visited_at"`
}

func (Visit) TableName() string { return "visits" }

// Admin stores the password hash (plain-text for this simple setup).
type Admin struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Password  string    `gorm:"size:128;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Admin) TableName() string { return "admins" }
