package commondb

import (
	"errors"
	"time"

	"github.com/sebastiw/sidan-backend/src/models"
)

// AuthState operations

func (d *CommonDatabase) CreateAuthState(state *models.AuthState) error {
	result := d.DB.Create(state)
	return result.Error
}

func (d *CommonDatabase) GetAuthState(id string) (*models.AuthState, error) {
	var state models.AuthState
	result := d.DB.Where("id = ?", id).First(&state)
	if result.Error != nil {
		return nil, result.Error
	}
	
	// Check if expired
	if state.ExpiresAt.Before(time.Now()) {
		d.DB.Delete(&state) // cleanup expired state
		return nil, errors.New("state expired")
	}
	
	return &state, nil
}

func (d *CommonDatabase) DeleteAuthState(id string) error {
	result := d.DB.Where("id = ?", id).Delete(&models.AuthState{})
	return result.Error
}

func (d *CommonDatabase) CleanupExpiredAuthStates() error {
	result := d.DB.Where("expires_at < ?", time.Now()).Delete(&models.AuthState{})
	return result.Error
}

// AuthToken operations

func (d *CommonDatabase) CreateAuthToken(token *models.AuthToken) error {
	result := d.DB.Create(token)
	return result.Error
}

func (d *CommonDatabase) GetAuthToken(memberID int64, provider string) (*models.AuthToken, error) {
	var token models.AuthToken
	result := d.DB.Where("member_id = ? AND provider = ?", memberID, provider).First(&token)
	if result.Error != nil {
		return nil, result.Error
	}
	return &token, nil
}

func (d *CommonDatabase) GetAuthTokenByMemberID(memberID int64) ([]models.AuthToken, error) {
	var tokens []models.AuthToken
	result := d.DB.Where("member_id = ?", memberID).Find(&tokens)
	if result.Error != nil {
		return nil, result.Error
	}
	return tokens, nil
}

func (d *CommonDatabase) UpdateAuthToken(token *models.AuthToken) error {
	// Use Updates to avoid updating zero-value created_at
	result := d.DB.Model(token).Updates(map[string]interface{}{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"token_type":    token.TokenType,
		"expires_at":    token.ExpiresAt,
		"scopes":        token.Scopes,
		"updated_at":    time.Now(),
	})
	return result.Error
}

func (d *CommonDatabase) DeleteAuthToken(memberID int64, provider string) error {
	result := d.DB.Where("member_id = ? AND provider = ?", memberID, provider).Delete(&models.AuthToken{})
	return result.Error
}

func (d *CommonDatabase) DeleteAllAuthTokens(memberID int64) error {
	result := d.DB.Where("member_id = ?", memberID).Delete(&models.AuthToken{})
	return result.Error
}

// AuthProviderLink operations

func (d *CommonDatabase) CreateAuthProviderLink(link *models.AuthProviderLink) error {
	result := d.DB.Create(link)
	return result.Error
}

func (d *CommonDatabase) GetAuthProviderLink(provider, providerUserID string) (*models.AuthProviderLink, error) {
	var link models.AuthProviderLink
	result := d.DB.Where("provider = ? AND provider_user_id = ?", provider, providerUserID).First(&link)
	if result.Error != nil {
		return nil, result.Error
	}
	return &link, nil
}

func (d *CommonDatabase) GetAuthProviderLinksByMemberID(memberID int64) ([]models.AuthProviderLink, error) {
	var links []models.AuthProviderLink
	result := d.DB.Where("member_id = ?", memberID).Find(&links)
	if result.Error != nil {
		return nil, result.Error
	}
	return links, nil
}

func (d *CommonDatabase) GetMemberByProviderEmail(provider, email string) (*models.Member, error) {
	var link models.AuthProviderLink
	result := d.DB.Where("provider = ? AND provider_email = ? AND email_verified = true", provider, email).First(&link)
	if result.Error != nil {
		return nil, result.Error
	}
	
	var member models.Member
	result = d.DB.Where("id = ?", link.MemberID).First(&member)
	if result.Error != nil {
		return nil, result.Error
	}
	
	return &member, nil
}

func (d *CommonDatabase) DeleteAuthProviderLink(provider, providerUserID string) error {
	result := d.DB.Where("provider = ? AND provider_user_id = ?", provider, providerUserID).Delete(&models.AuthProviderLink{})
	return result.Error
}

// AuthSession operations

func (d *CommonDatabase) CreateAuthSession(session *models.AuthSession) error {
	result := d.DB.Create(session)
	return result.Error
}

func (d *CommonDatabase) GetAuthSession(id string) (*models.AuthSession, error) {
	var session models.AuthSession
	result := d.DB.Where("id = ?", id).First(&session)
	if result.Error != nil {
		return nil, result.Error
	}
	
	// Check if expired
	if session.IsExpired() {
		d.DB.Delete(&session) // cleanup expired session
		return nil, errors.New("session expired")
	}
	
	return &session, nil
}

func (d *CommonDatabase) UpdateAuthSession(session *models.AuthSession) error {
	result := d.DB.Save(session)
	return result.Error
}

func (d *CommonDatabase) DeleteAuthSession(id string) error {
	result := d.DB.Where("id = ?", id).Delete(&models.AuthSession{})
	return result.Error
}

func (d *CommonDatabase) DeleteAllAuthSessions(memberID int64) error {
	result := d.DB.Where("member_id = ?", memberID).Delete(&models.AuthSession{})
	return result.Error
}

func (d *CommonDatabase) CleanupExpiredAuthSessions() error {
	result := d.DB.Where("expires_at < ?", time.Now()).Delete(&models.AuthSession{})
	return result.Error
}

func (d *CommonDatabase) TouchAuthSession(id string) error {
	result := d.DB.Model(&models.AuthSession{}).
		Where("id = ?", id).
		Update("last_activity", time.Now())
	return result.Error
}
