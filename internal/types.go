package internal

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string `json:"name"`
	Email string `gorm:"unique" json:"email"`
	Password []byte `json:"password"`
	Admin bool `json:"admin"`
}

type Place struct {
	gorm.Model
	Name          string `gorm:"unique" json:"name"`
	Latitude   float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	Image string `json:"image"`
	BusinessHours string `json:"businessHours"`
	Tel string `json:"tel"`
	Url string `json:"url"`
  Memo string `json:"memo"`
	SocketNum int `json:"socketNum"`
	Wifi int `json:"wifi"`
  Smoke int `json:"smoke"`
}

type Credentials struct {
	Email string
	Password string
}

type RegisterUserBody struct {
	Name *string
	Email *string
	Password *string
}

type Login struct {
	Username string
	Token string
}
