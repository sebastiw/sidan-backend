package models

import "time"

// Session stores a long-lived sidan refresh token issued after web OAuth2 login
type Session struct {
	Token        string    `gorm:"primaryKey;size:64"`
	MemberNumber int64     `gorm:"not null;index"`
	Email        string    `gorm:"size:255;not null"`
	Provider     string    `gorm:"size:32;not null"`
	ExpiresAt    time.Time `gorm:"not null;index"`
	CreatedAt    time.Time
}

func (Session) TableName() string {
	return "oauth2_sessions"
}
