package database

import (
	"fmt"
	"strings"
)

type SideKick struct {
	Number string `json:""`
}

func (s SideKick) Fmt() string {
	return s.Number
}
//swagger:response Entry
type Entry struct {
	Id             int64      `json:"id"`
	Date           string     `json:"date"`
	Time           string     `json:"time"`
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
	SideKicks      []SideKick `json:"sidekicks"`
}

func (e Entry) Fmt() string {
	sk := make([]string, 0)
	for _, n := range e.SideKicks {
		sk = append(sk, n.Fmt())
	}
	s := make([]string, 0)
	s = add_i(s, "Id", e.Id)
	s = add_s(s, "Date", e.Date)
	s = add_s(s, "Time", e.Time)
	s = add_s(s, "Msg", e.Msg)
	s = add_i(s, "Status", e.Status)
	s = add_i(s, "Cl", e.Cl)
	s = add_s(s, "Sig", e.Sig)
	s = add_s(s, "Email", e.Email)
	s = add_s(s, "Place", e.Place)
	s = add_sp(s, "Ip", e.Ip)
	s = add_sp(s, "Host", e.Host)
	s = add_i(s, "Olsug", e.Olsug)
	s = add_i(s, "Enheter", e.Enheter)
	s = add_fp(s, "Lat", e.Lat)
	s = add_fp(s, "Lon", e.Lon)
	s = add_b(s, "Report", e.Report)
	s = add_i(s, "Likes", e.Likes)
	s = add_b(s, "Secret", e.Secret)
	s = add_b(s, "PersonalSecret", e.PersonalSecret)
	s = add_s(s, "SideKicks", "["+strings.Join(sk, ",")+"]")

	return fmt.Sprintf("Entry{%s}", strings.Join(s, ", "))
}
