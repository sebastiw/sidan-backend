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
