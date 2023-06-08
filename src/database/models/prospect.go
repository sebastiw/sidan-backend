package database

import(
	"fmt"
	"strings"
)
// swagger:model

type Prospect struct {
	Id int64 `json:"id"`
	Status string `json:"status"`
	Number int64 `json:"number"`
	Name string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	History string `json:"history"`
}

func (p Prospect) Fmt() string {
	s := make([]string, 0)
	s = add_i(s, "Id", p.Id)
	s = add_s(s, "Status", p.Status)
	s = add_i(s, "Number", p.Number)
	s = add_s(s, "Name", p.Name)
	s = add_s(s, "Email", p.Email)
	s = add_s(s, "Phone", p.Phone)
	s = add_s(s, "History", p.History)
	return fmt.Sprintf("Prospect{%s}", strings.Join(s, ", "))
}
