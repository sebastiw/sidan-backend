package commondb

import (
	"strconv"
	"errors"

	"github.com/sebastiw/sidan-backend/src/models"
)

func (d *CommonDatabase) GetUserFromEmails(emails []string) (*models.User, error) {
	var user models.User
	result := d.DB.Where("email IN ? AND isvalid = true", emails).First(&user)

	if result.Error != nil {
		return nil,result.Error
	}
	return &user,nil
}

func (d *CommonDatabase) GetUserFromLogin(username string, password string) (*models.User, error) {
	var user models.User

	types := []byte{'#', 'P', 'S', 'p', 's'}
	if !contains(types, username[0])  {
		return nil,errors.New("Username not starting with '#', 'P', 'S'")
	}
	number, err := strconv.ParseInt(username[1:], 10, 64)
	if err != nil {
		return nil,err
	}

	result := d.DB.Where("password_classic = ? AND isvalid = true", password).First(&user, models.User{Number: number})

	if result.Error != nil {
		return nil,result.Error
	}
	return &user,nil
}

func contains(s []byte, str byte) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

