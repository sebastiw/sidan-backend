package models

import (
	"fmt"
	"time"
)

// Article represents a news/blog article in the cl_news table
type Article struct {
	Id       int64      `gorm:"column:Id;primaryKey;autoIncrement" json:"id"`
	Header   *string    `gorm:"column:header" json:"header"`
	Body     *string    `gorm:"column:body" json:"body"`
	Date     *string    `gorm:"column:date;type:date" json:"date"`
	Time     *string    `gorm:"column:time;type:time" json:"time"`
	DateTime *time.Time `gorm:"column:datetime;type:datetime;not null;default:now()" json:"datetime"`
}

// TableName specifies the table name for GORM
func (Article) TableName() string {
	return "cl_news"
}

// Fmt formats Article for logging
func (a Article) Fmt() string {
	header := ""
	if a.Header != nil {
		header = *a.Header
	}
	return fmt.Sprintf("Article{Id: %d, Header: %s}", a.Id, header)
}
