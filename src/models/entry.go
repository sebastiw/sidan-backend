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
	DateTime       time.Time  `gorm:"type:time" json:"datetime"` // TODO: convert from "2020-04-15 18:50:13" to ISO.
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
	Likes          int64      `json:"likes"`
	Secret         bool       `json:"secret"`
	PersonalSecret bool       `json:"personal_secret"`
	SideKicks      []SideKick `gorm:"foreignKey:Id";json:"sidekicks"`
}

func (Entry) TableName() string {
  return "cl2003_msgs"
}
