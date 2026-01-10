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
