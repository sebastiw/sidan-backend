package commondb

import (
	"github.com/sebastiw/sidan-backend/src/models"
)

func (d *CommonDatabase) CreateArr(arr *models.Arr) (*models.Arr, error) {
	result := d.DB.Create(arr)
	if result.Error != nil {
		return nil, result.Error
	}
	return arr, nil
}

func (d *CommonDatabase) ReadArr(id int64) (*models.Arr, error) {
	var arr models.Arr
	result := d.DB.First(&arr, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &arr, nil
}

func (d *CommonDatabase) ReadArrs(take int, skip int) ([]models.Arr, error) {
	var arrs []models.Arr
	result := d.DB.Order("id DESC").Limit(take).Offset(skip).Find(&arrs)
	if result.Error != nil {
		return nil, result.Error
	}
	return arrs, nil
}

func (d *CommonDatabase) UpdateArr(arr *models.Arr) (*models.Arr, error) {
	result := d.DB.Save(arr)
	if result.Error != nil {
		return nil, result.Error
	}
	return arr, nil
}

func (d *CommonDatabase) DeleteArr(arr *models.Arr) (*models.Arr, error) {
	result := d.DB.Delete(arr)
	if result.Error != nil {
		return nil, result.Error
	}
	return arr, nil
}
