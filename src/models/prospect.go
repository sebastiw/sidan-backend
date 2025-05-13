package models

import (
	"fmt"
	"strings"
)

//swagger:response Prospect
type Prospect struct {
	Id      int64  `json:"id"`
	Status  string `json:"status"`
	Number  int64  `json:"number"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	History string `json:"history"`
}

func (Prospect) TableName() string {
  return "cl2007_prospects"
}

func (p Prospect) Fmt() string {
	s := make([]string, 0)
	s = addI(s, "Id", p.Id)
	s = addS(s, "Status", p.Status)
	s = addI(s, "Number", p.Number)
	s = addS(s, "Name", p.Name)
	s = addS(s, "Email", p.Email)
	s = addS(s, "Phone", p.Phone)
	s = addS(s, "History", p.History)
	return fmt.Sprintf("Prospect{%s}", strings.Join(s, ", "))
}
