package models

import (
	"time"
)

// DeviceCode represents a device authorization flow entry
type DeviceCode struct {
	DeviceCode      string    `gorm:"primaryKey;size:128" json:"device_code"`
	UserCode        string    `gorm:"size:16;uniqueIndex;not null" json:"user_code"`
	VerificationURI string    `gorm:"type:text;not null" json:"verification_uri"`
	ExpiresAt       time.Time `gorm:"not null;index" json:"expires_at"`
	Interval        int       `gorm:"not null" json:"interval"` // Polling interval in seconds
	
	// Approved state
	Approved        bool      `gorm:"default:false" json:"approved"`
	MemberNumber    *int64    `json:"member_number,omitempty"`
	Email           *string   `json:"email,omitempty"`
	Scopes          *string   `gorm:"type:text" json:"scopes,omitempty"` // JSON array
	Provider        string    `gorm:"size:32" json:"provider"`
	
	CreatedAt       time.Time `json:"created_at"`
}

func (DeviceCode) TableName() string {
	return "device_codes"
}
