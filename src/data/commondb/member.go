package commondb

import (
	"github.com/sebastiw/sidan-backend/src/models"
)

func (d *CommonDatabase) CreateMember(member *models.Member) (*models.Member, error) {
	result := d.DB.Create(&member)
	if result.Error != nil {
		return nil, result.Error
	}
	return member, nil
}

func (d *CommonDatabase) ReadMember(id int64) (*models.Member, error) {
	var member models.Member

	result := d.DB.First(&member, models.Member{Id: id})

	if result.Error != nil {
		return nil, result.Error
	}
	return &member, nil
}

func (d *CommonDatabase) ReadMemberByNumber(number int64) (*models.Member, error) {
	var member models.Member

	result := d.DB.First(&member, models.Member{Number: number})

	if result.Error != nil {
		return nil, result.Error
	}
	return &member, nil
}

func (d *CommonDatabase) ReadMembers(onlyValid bool) ([]models.Member, error) {
	var members []models.Member

	result := d.DB.Where("isvalid = 1").Order("number desc").Find(&members)

	if result.Error != nil {
		return nil, result.Error
	}
	return members, nil
}

func (d *CommonDatabase) UpdateMember(member *models.Member) (*models.Member, error) {
	result := d.DB.Save(&member)

	if result.Error != nil {
		return nil, result.Error
	}
	return member, nil
}

func (d *CommonDatabase) DeleteMember(member *models.Member) (*models.Member, error) {
	result := d.DB.Delete(&member)

	if result.Error != nil {
		return nil, result.Error
	}
	return member, nil
}
