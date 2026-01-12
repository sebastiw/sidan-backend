package mysqldb

import (
       "github.com/sebastiw/sidan-backend/src/models"
)

func (d *MySQLDatabase) CreateMember(member *models.Member) (*models.Member, error) {
	return d.CommonDB.CreateMember(member)
}

func (d *MySQLDatabase) ReadMember(id int64) (*models.Member, error) {
	return d.CommonDB.ReadMember(id)
}

func (d *MySQLDatabase) ReadMemberByNumber(number int64) (*models.Member, error) {
	return d.CommonDB.ReadMemberByNumber(number)
}

func (d *MySQLDatabase) ReadMembers(onlyValid bool) ([]models.Member, error) {
	return d.CommonDB.ReadMembers(onlyValid)
}

func (d *MySQLDatabase) UpdateMember(member *models.Member) (*models.Member, error) {
	return d.CommonDB.UpdateMember(member)
}

func (d *MySQLDatabase) DeleteMember(member *models.Member) (*models.Member, error) {
	return d.CommonDB.DeleteMember(member)
}
