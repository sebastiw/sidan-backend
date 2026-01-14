package mysqldb

import (
	"github.com/sebastiw/sidan-backend/src/models"
)

func (d *MySQLDatabase) CreateArr(arr *models.Arr) (*models.Arr, error) {
	return d.CommonDB.CreateArr(arr)
}

func (d *MySQLDatabase) ReadArr(id int64) (*models.Arr, error) {
	return d.CommonDB.ReadArr(id)
}

func (d *MySQLDatabase) ReadArrs(take int, skip int) ([]models.Arr, error) {
	return d.CommonDB.ReadArrs(take, skip)
}

func (d *MySQLDatabase) UpdateArr(arr *models.Arr) (*models.Arr, error) {
	return d.CommonDB.UpdateArr(arr)
}

func (d *MySQLDatabase) DeleteArr(arr *models.Arr) (*models.Arr, error) {
	return d.CommonDB.DeleteArr(arr)
}
