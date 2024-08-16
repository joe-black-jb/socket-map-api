package internal

import (
	"gorm.io/gorm"
)

type Response struct {
	StatusCode int      `json:"statusCode"`
	Data   interface{} `json:"data"`
}

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
	Address          string `json:"address"`
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

type Station struct {
	gorm.Model
	Name string `gorm:"unique" json:"name"`
	Latitude   float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
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

type OsmPlaceDetail struct {
	PlaceId int `json:"place_id"`
	Licence string `json:"licence"`
	OsmType string `json:"osm_type"`
	OsmId int `json:"osm_id"`
	Lat string `json:"lat"`
	Lon string `json:"lon"`
	Class string `json:"class"`
	Type string `json:"type"`
	PlaceRank int `json:"place_rank"`
	Importance float64 `json:"importance"`
	AddressType string `json:"addresstype"`
	Name string `json:"name"`
	DisplayName string `json:"display_name"`
	BoundingBox []string `json:"boundingbox"`
}
