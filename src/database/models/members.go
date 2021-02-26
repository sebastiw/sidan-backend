package database

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

