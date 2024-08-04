package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joe-black-jb/socket-map-api/internal"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Users = []internal.User{
	// Password: pass のハッシュ値
	{
		Name: "サンプルユーザ", 
		Password: []byte("$2a$10$lbnP92Wdad2olUA18I1Xbe21Zuma6eoriPCmohCxAku8Bdzo3.SL2"), 
		Email: "sample@sample.com",
		Admin: false,
	},
}

var Places = []internal.Place{
	{
		Name: "カフェ・ベローチェ 京成船橋駅前店",
		Latitude: 35.69985840379561, 
		Longitude: 139.98670535289517,
	},
	{
		Name: "ドトールコーヒーショップ 船橋駅南口店",
		Latitude: 35.70107504987384,
		Longitude: 139.98581438173127,
	},
}


func main() {
	enverr := godotenv.Load()
	if enverr != nil {
		log.Fatal("Error loading .env file")
	}

	dbuser := os.Getenv("MYSQL_USER")
	dbpass := os.Getenv("MYSQL_ROOT_PASSWORD")
	dbname := os.Getenv("MYSQL_DATABASE")
	// docker コンテナを立ち上げている場合、ホスト名は 127.0.0.1 ではなくサービス名（db）
	// コンテナ外でスクリプト実行想定のため、ホスト名は 127.0.0.1 にする
	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbuser, dbpass, dbname)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	// Drop Table
	db.Migrator().DropTable(&internal.User{})
	db.Migrator().DropTable(&internal.Place{})

	// Migrate the schema
	db.AutoMigrate(&internal.User{}, &internal.Place{})

	// Batch Create
	db.Create(&Users)
	db.Create(&Places)

	mysql, _ := db.DB()
	mysql.Close()
	fmt.Println("Done!! ⭐️")
}
