package internal

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Response struct {
	StatusCode int         `json:"statusCode"`
	Data       interface{} `json:"data"`
}

type User struct {
	// gorm.Model
	ID        string    `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `json:"name"`
	Email     string    `gorm:"unique" json:"email"`
	Password  []byte    `json:"password"`
	Admin     bool      `json:"admin"`
}

type Place struct {
	// gorm.Model
	ID            string    `gorm:"primaryKey" json:"id" dynamodbav:"id"`
	CreatedAt     time.Time `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
	Name          string    `gorm:"unique" json:"name" dynamodbav:"name"`
	Address       string    `json:"address" dynamodbav:"address"`
	Latitude      float64   `json:"latitude" dynamodbav:"latitude"`
	Longitude     float64   `json:"longitude" dynamodbav:"longitude"`
	Image         string    `json:"image" dynamodbav:"image"`
	BusinessHours string    `json:"businessHours" dynamodbav:"businessHours"`
	Tel           string    `json:"tel" dynamodbav:"tel"`
	Url           string    `json:"url" dynamodbav:"url"`
	Memo          string    `json:"memo" dynamodbav:"memo"`
	Socket        int       `json:"socket" dynamodbav:"socket"`       // 0: なし, 1: あり, 2: 不明
	SocketNum     int       `json:"socketNum" dynamodbav:"socketNum"` // 0: なし, 1: あり, 2: 不明
	Wifi          int       `json:"wifi" dynamodbav:"wifi"`
	Smoke         int       `json:"smoke" dynamodbav:"smoke"`
}

type Station struct {
	// gorm.Model
	ID        string    `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `gorm:"unique" json:"name"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
}

type Credentials struct {
	Email    string
	Password string
}

type RegisterUserBody struct {
	Name     *string
	Email    *string
	Password *string
}

type Login struct {
	Username string
	Token    string
}

type OsmPlaceDetail struct {
	PlaceId     int      `json:"place_id"`
	Licence     string   `json:"licence"`
	OsmType     string   `json:"osm_type"`
	OsmId       int      `json:"osm_id"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	Class       string   `json:"class"`
	Type        string   `json:"type"`
	PlaceRank   int      `json:"place_rank"`
	Importance  float64  `json:"importance"`
	AddressType string   `json:"addresstype"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	BoundingBox []string `json:"boundingbox"`
}

type BucketBasics struct {
	S3Client *s3.Client
}
