package models

import (
	"fmt"
	"strings"
)

//swagger:response Member
type Member struct {
	Id                           int64   `json:"id"`
	Number                       int64   `json:"number"`
	Name                         *string `json:"name"`
	Email                        *string `json:"email"`
	Im                           string  `json:"im"`
	Phone                        *string `json:"phone"`
	Adress                       *string `json:"address"`
	Adressurl                    *string `json:"address_url"`
	Title                        *string `json:"title"`
	History                      *string `json:"history"`
	Picture                      *string `json:"picture"`
	Password                     *string `json:"password"`
	Isvalid                      *bool   `json:"is_valid"`
	Password_classic             *string `json:"password_classic"`
	Password_classic_resetstring *string `json:"password_classic_resetstring"`
	Password_resetstring         *string `json:"password_resetstring"`
}

func (Member) TableName() string {
  return "cl2007_members"
}

func (m Member) Fmt() string {
	s := make([]string, 0)
	isvalid := true
	s = addI(s, "Id", m.Id)
	s = addI(s, "Number", m.Number)
	s = addSp(s, "Name", m.Name)
	s = addSp(s, "Email", m.Email)
	s = addS(s, "Im", m.Im)
	s = addSp(s, "Phone", m.Phone)
	s = addSp(s, "Adress", m.Adress)
	s = addSp(s, "Adressurl", m.Adressurl)
	s = addSp(s, "Title", m.Title)
	s = addSp(s, "History", m.History)
	s = addSp(s, "Picture", m.Picture)
	s = addBp(s, "Isvalid", m.Isvalid)
	return fmt.Sprintf("Member{%s, Isvalid: %t}", strings.Join(s, ", "), isvalid)
}

type MemberLite struct {
	Id     int64   `json:"id"`
	Number *int64 `json:"number"`
	Title  *string `json:"title"`
}
