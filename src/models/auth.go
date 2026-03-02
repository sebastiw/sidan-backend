package models

import (
	"time"
)

// AuthState represents OAuth2 state for CSRF protection
type AuthState struct {
	ID           string    `gorm:"primaryKey;size:64" json:"id"`
	Provider     string    `gorm:"size:32;not null" json:"provider"`
	Nonce        string    `gorm:"size:64;not null" json:"nonce"`
	PKCEVerifier string    `gorm:"column:pkce_verifier;size:128" json:"pkce_verifier,omitempty"`
	RedirectURI  string    `gorm:"column:redirect_uri;type:text" json:"redirect_uri,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `gorm:"not null;index" json:"expires_at"`
}

func (AuthState) TableName() string {
	return "auth_states"
}

// DeviceAuthState represents device authorization flow state
type DeviceAuthState struct {
	DeviceCode  string    `gorm:"primaryKey;size:64" json:"device_code"`
	UserCode    string    `gorm:"uniqueIndex;size:32;not null" json:"user_code"`
	Status      string    `gorm:"size:16;not null;default:'pending'" json:"status"` // pending, approved, denied
	MemberNumber *int64    `json:"member_number,omitempty"` // Nullable
	Scopes      string    `gorm:"type:text" json:"scopes"` // JSON or comma-separated
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `gorm:"not null;index" json:"expires_at"`
	LastChecked time.Time `json:"last_checked"`
}

func (DeviceAuthState) TableName() string {
	return "device_auth_states"
}
