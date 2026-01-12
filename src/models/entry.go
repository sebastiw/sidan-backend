package models

import (
	"time"
)

//swagger:response SideKick
type SideKick struct {
	Id     int64  `gorm:"primaryKey";json:"id"`
	Number string `gorm:"primaryKey";json:"number"`
}

func (SideKick) TableName() string {
  return "cl2003_msgs_kumpaner"
}

//swagger:response Entry
type Entry struct {
	Id             int64      `json:"id"`
	Date           string     `json:"date"`
	Time           string     `json:"time"`
	DateTime       time.Time  `gorm:"column:datetime" json:"datetime"`
	Msg            string     `json:"msg"`
	Status         int64      `json:"status"`
	Cl             int64      `json:"cl"`
	Sig            string     `json:"sig"`
	Email          string     `json:"email"`
	Place          string     `json:"place"`
	Ip             *string    `json:"ip"`
	Host           *string    `json:"host"`
	Olsug          int64      `json:"olsug"`
	Enheter        int64      `json:"enheter"`
	Lat            *float64   `json:"lat"`
	Lon            *float64   `json:"lon"`
	Report         bool       `json:"report"`
	
	// Computed fields from related tables
	Likes          int64       `gorm:"-" json:"likes"` // Count from 2003_likes table
	Secret         bool        `gorm:"-" json:"secret"` // TRUE if ANY permission exists
	PersonalSecret bool        `gorm:"-" json:"personal_secret"` // TRUE if permission with user_id != 0 exists
	
	// Relationships
	SideKicks      []SideKick   `gorm:"foreignKey:Id" json:"sidekicks"`
	LikeRecords    []Like       `gorm:"foreignKey:Id" json:"-"` // Hidden from JSON
	Permissions    []Permission `gorm:"foreignKey:Id" json:"-"` // Hidden from JSON
}

func (Entry) TableName() string {
  return "cl2003_msgs"
}
