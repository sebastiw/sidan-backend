package models

type SideKick struct {
	Number string `json:"number"`
}

//swagger:response Entry
type Entry struct {
	Id             int64      `json:"id"`
	Date           string     `json:"date"`
	Time           string     `json:"time"`
	DateTime       string     `json:"datetime"`
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
