package models

import (
	"database/sql/driver"
	"encoding/json"
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

// StringArray is a custom type for JSON array of strings in MySQL
type StringArray []string

// Scan implements sql.Scanner interface
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = StringArray{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer interface
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return json.Marshal([]string{})
	}
	return json.Marshal(s)
}

// AuthToken represents OAuth2 tokens stored in database
type AuthToken struct {
	ID           int64       `gorm:"primaryKey" json:"id"`
	MemberID     int64       `gorm:"not null;uniqueIndex:unique_member_provider" json:"member_id"`
	Provider     string      `gorm:"size:32;not null;uniqueIndex:unique_member_provider" json:"provider"`
	AccessToken  string      `gorm:"column:access_token;type:text;not null" json:"-"` // encrypted, never expose
	RefreshToken *string     `gorm:"column:refresh_token;type:text" json:"-"`         // encrypted, never expose
	TokenType    string      `gorm:"column:token_type;size:32;default:Bearer" json:"token_type"`
	ExpiresAt    *time.Time  `gorm:"column:expires_at;index" json:"expires_at,omitempty"`
	Scopes       StringArray `gorm:"type:json" json:"scopes"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

func (AuthToken) TableName() string {
	return "auth_tokens"
}

// IsExpired checks if the token has expired
func (t *AuthToken) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false // no expiry means token doesn't expire
	}
	return t.ExpiresAt.Before(time.Now())
}

// AuthProviderLink represents the link between a member and OAuth2 provider
type AuthProviderLink struct {
	ID             int64     `gorm:"primaryKey" json:"id"`
	MemberID       int64     `gorm:"not null;index" json:"member_id"`
	Provider       string    `gorm:"size:32;not null;uniqueIndex:unique_provider_user" json:"provider"`
	ProviderUserID string    `gorm:"column:provider_user_id;size:255;not null;uniqueIndex:unique_provider_user" json:"provider_user_id"`
	ProviderEmail  string    `gorm:"column:provider_email;size:255;not null;index:idx_provider_email" json:"provider_email"`
	EmailVerified  bool      `gorm:"column:email_verified;default:false" json:"email_verified"`
	LinkedAt       time.Time `gorm:"column:linked_at" json:"linked_at"`
}

func (AuthProviderLink) TableName() string {
	return "auth_provider_links"
}

// SessionData represents the JSON data stored in auth_sessions
type SessionData struct {
	Scopes   []string `json:"scopes"`
	Provider string   `json:"provider"`
}

// Scan implements sql.Scanner interface
func (s *SessionData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer interface
func (s SessionData) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// AuthSession represents a user session
type AuthSession struct {
	ID           string       `gorm:"primaryKey;size:128" json:"id"`
	MemberID     int64        `gorm:"not null;index" json:"member_id"`
	Data         *SessionData `gorm:"type:json" json:"data"`
	CreatedAt    time.Time    `json:"created_at"`
	ExpiresAt    time.Time    `gorm:"not null;index" json:"expires_at"`
	LastActivity time.Time    `gorm:"column:last_activity" json:"last_activity"`
}

func (AuthSession) TableName() string {
	return "auth_sessions"
}

// IsExpired checks if the session has expired
func (s *AuthSession) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}
