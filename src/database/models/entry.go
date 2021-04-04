package database

import(
	"fmt"
	"strings"
)

type Entry struct {
  Id int64 `json:""`
  Date string `json:""`
  Time string `json:""`
  Msg string `json:""`
  Status int64 `json:""`
  Cl int64 `json:""`
  Sig string `json:""`
  Email string `json:""`
  Place string `json:""`
  Ip *string `json:""`
  Host *string `json:""`
  Olsug int64 `json:""`
  Enheter int64 `json:""`
  Lat *float64 `json:""`
  Lon *float64 `json:""`
  Report bool `json:""`
}

func (e Entry) Fmt() string {
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

	return fmt.Sprintf("Entry{%s}", strings.Join(s, ", "))
}
