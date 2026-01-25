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

// Device Auth operations

func (d *CommonDatabase) CreateDeviceAuthState(state *models.DeviceAuthState) error {
	result := d.DB.Create(state)
	return result.Error
}

func (d *CommonDatabase) GetDeviceAuthStateByDeviceCode(code string) (*models.DeviceAuthState, error) {
	var state models.DeviceAuthState
	result := d.DB.Where("device_code = ?", code).First(&state)
	if result.Error != nil {
		return nil, result.Error
	}
	
	// Check if expired
	if state.ExpiresAt.Before(time.Now()) {
		// Don't delete immediately, let cleanup job handle it or handle it in handler
		return nil, errors.New("state expired")
	}
	
	return &state, nil
}

func (d *CommonDatabase) GetDeviceAuthStateByUserCode(code string) (*models.DeviceAuthState, error) {
	var state models.DeviceAuthState
	result := d.DB.Where("user_code = ?", code).First(&state)
	if result.Error != nil {
		return nil, result.Error
	}
	
	if state.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("state expired")
	}
	
	return &state, nil
}

func (d *CommonDatabase) UpdateDeviceAuthState(state *models.DeviceAuthState) error {
	result := d.DB.Save(state)
	return result.Error
}
