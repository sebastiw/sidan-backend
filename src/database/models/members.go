package database

import(
	"fmt"
	"strings"
)

type Member struct {
	Id int64 `json:"id"`
	Number *string `json:"number"`
	Name *string `json:"name"`
	Email *string `json:"email"`
	Im string `json:"im"`
	Phone *string `json:"phone"`
	Adress *string `json:"address"`
	Adressurl *string `json:"address_url"`
	Title *string `json:"title"`
	History *string `json:"history"`
	Picture *string `json:"picture"`
	Password *string `json:"password"`
	Isvalid *bool `json:"is_valid"`
	Password_classic *string `json:"password_classic"`
	Password_classic_resetstring *string `json:"password_classic_resetstring"`
	Password_resetstring *string `json:"password_resetstring"`
}

func (m Member) Fmt() string {
	s := make([]string, 0)
	isvalid := true
	s = add_i(s, "Id", m.Id)
	s = add_sp(s, "Number", m.Number)
	s = add_sp(s, "Name", m.Name)
	s = add_sp(s, "Email", m.Email)
	s = add_s(s, "Im", m.Im)
	s = add_sp(s, "Phone", m.Phone)
	s = add_sp(s, "Adress", m.Adress)
	s = add_sp(s, "Adressurl", m.Adressurl)
	s = add_sp(s, "Title", m.Title)
	s = add_sp(s, "History", m.History)
	s = add_sp(s, "Picture", m.Picture)
	s = add_bp(s, "Isvalid", m.Isvalid)
	return fmt.Sprintf("Member{%s, Isvalid: %t}", strings.Join(s, ", "), isvalid)
}
