package models

import "fmt"

// Arr represents an event/arrangemang in the cl2015_arrsidan table
type Arr struct {
	Id          int64   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Namn        *string `gorm:"column:namn" json:"namn"`
	StartDate   *string `gorm:"column:start_date" json:"start_date"`
	Plats       *string `gorm:"column:plats" json:"plats"`
	Organisator *string `gorm:"column:organisator" json:"organisator"`
	Deltagare   *string `gorm:"column:deltagare" json:"deltagare"`
	Kanske      *string `gorm:"column:kanske" json:"kanske"`
	Hetsade     *string `gorm:"column:hetsade" json:"hetsade"`
	Losen       *string `gorm:"column:losen" json:"losen"`
	Fularr      *string `gorm:"column:fularr" json:"fularr"`
}

// TableName specifies the table name for GORM
func (Arr) TableName() string {
	return "cl2015_arrsidan"
}

// Fmt formats Arr for logging
func (a Arr) Fmt() string {
	namn := ""
	if a.Namn != nil {
		namn = *a.Namn
	}
	plats := ""
	if a.Plats != nil {
		plats = *a.Plats
	}
	return fmt.Sprintf("Arr{Id: %d, Namn: %s, Plats: %s}", a.Id, namn, plats)
}
