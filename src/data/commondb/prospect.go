package commondb

import (
	"github.com/sebastiw/sidan-backend/src/models"
)

func (d *CommonDatabase) CreateProspect(prospect *models.Prospect) (*models.Prospect, error) {
	result := d.DB.Create(&prospect)
	if result.Error != nil {
		return nil, result.Error
	}
	return prospect, nil
}

func (d *CommonDatabase) ReadProspect(id int64) (*models.Prospect, error) {
	var prospect models.Prospect
	result := d.DB.First(&prospect, models.Prospect{Id: id})
	if result.Error != nil {
		return nil, result.Error
	}
	return &prospect, nil
}

func (d *CommonDatabase) ReadProspects(status string) ([]models.Prospect, error) {
	var prospects []models.Prospect
	db := d.DB.Order("number desc")
	if status != "" {
		db = db.Where("status = ?", status)
	}
	result := db.Find(&prospects)
	if result.Error != nil {
		return nil, result.Error
	}
	return prospects, nil
}

func (d *CommonDatabase) UpdateProspect(prospect *models.Prospect) (*models.Prospect, error) {
	result := d.DB.Save(&prospect)
	if result.Error != nil {
		return nil, result.Error
	}
	return prospect, nil
}

func (d *CommonDatabase) DeleteProspect(prospect *models.Prospect) (*models.Prospect, error) {
	result := d.DB.Delete(&prospect)
	if result.Error != nil {
		return nil, result.Error
	}
	return prospect, nil
}
