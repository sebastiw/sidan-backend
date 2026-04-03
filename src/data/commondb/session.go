package commondb

import (
	"errors"
	"time"

	"github.com/sebastiw/sidan-backend/src/models"
)

func (d *CommonDatabase) CreateSession(session *models.Session) error {
	return d.DB.Create(session).Error
}

func (d *CommonDatabase) GetSession(token string) (*models.Session, error) {
	var session models.Session
	result := d.DB.Where("token = ?", token).First(&session)
	if result.Error != nil {
		return nil, result.Error
	}

	if session.ExpiresAt.Before(time.Now()) {
		d.DB.Delete(&session)
		return nil, errors.New("session expired")
	}

	return &session, nil
}

func (d *CommonDatabase) DeleteSession(token string) error {
	return d.DB.Where("token = ?", token).Delete(&models.Session{}).Error
}

func (d *CommonDatabase) CleanupExpiredSessions() error {
	return d.DB.Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error
}
