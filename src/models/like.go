package models

import (
	"time"
)

// Like represents a like on an entry
type Like struct {
	Date time.Time `gorm:"column:date;type:date" json:"date"`
	Time time.Time `gorm:"column:time;type:time" json:"time"`
	Id   int64     `gorm:"column:id" json:"id"`  // References cl2003_msgs.id
	Sig  string    `gorm:"column:sig" json:"sig"`
	Host string    `gorm:"column:host" json:"host"`
}

func (Like) TableName() string {
	return "2003_likes"
}
