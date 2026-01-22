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

// DeviceCode operations

func (d *CommonDatabase) CreateDeviceCode(code *models.DeviceCode) error {
	result := d.DB.Create(code)
	return result.Error
}

func (d *CommonDatabase) GetDeviceCodeByUserCode(userCode string) (*models.DeviceCode, error) {
	var code models.DeviceCode
	result := d.DB.Where("user_code = ?", userCode).First(&code)
	if result.Error != nil {
		return nil, result.Error
	}
	
	// Check if expired
	if code.ExpiresAt.Before(time.Now()) {
		d.DB.Delete(&code)
		return nil, errors.New("device code expired")
	}
	
	return &code, nil
}

func (d *CommonDatabase) GetDeviceCodeByDeviceCode(deviceCode string) (*models.DeviceCode, error) {
	var code models.DeviceCode
	result := d.DB.Where("device_code = ?", deviceCode).First(&code)
	if result.Error != nil {
		return nil, result.Error
	}
	
	// Check if expired
	if code.ExpiresAt.Before(time.Now()) {
		d.DB.Delete(&code)
		return nil, errors.New("device code expired")
	}
	
	return &code, nil
}

func (d *CommonDatabase) UpdateDeviceCode(code *models.DeviceCode) error {
	result := d.DB.Save(code)
	return result.Error
}

func (d *CommonDatabase) DeleteDeviceCode(deviceCode string) error {
	result := d.DB.Where("device_code = ?", deviceCode).Delete(&models.DeviceCode{})
	return result.Error
}

func (d *CommonDatabase) CleanupExpiredDeviceCodes() error {
	result := d.DB.Where("expires_at < ?", time.Now()).Delete(&models.DeviceCode{})
	return result.Error
}
