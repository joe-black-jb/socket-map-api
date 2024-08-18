package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joe-black-jb/socket-map-api/internal"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type GeometryData struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

type Property struct {
	N02_001  string `json:"n02_001"`
	N02_002  string `json:"n02_002"`
	N02_003  string `json:"n02_003"`
	N02_004  string `json:"n02_004"`
	N02_005  string `json:"n02_005"`
	N02_005c string `json:"n02_005c"`
	N02_005g string `json:"n02_005g"`
}

type Feature struct {
	Type       string       `json:"type"`
	Properties Property     `json:"properties"`
	Geometry   GeometryData `json:"geometry"`
}

type Station struct {
	Type     string    `json:"type"`
	Name     string    `json:"name"`
	Features []Feature `json:"features"`
}

// 駅名の重複チェック
func ContainsStation(stations []internal.Station, name string) bool {
	for _, station := range stations {
		if station.Name == name {
			return true
		}
	}
	return false
}

// json ファイルからデータを作る
func main() {
	// jsonファイルの読み込み
	jsonFile, err := os.Open("seeder/stations.geojson")
	if err != nil {
		fmt.Println("JSONファイルを開けません", err)
		return
	}
	defer jsonFile.Close()
	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("JSONデータを読み込めません", err)
		return
	}

	var StationData Station
	json.Unmarshal(jsonData, &StationData)

	// DB接続
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
	dropErr := db.Migrator().DropTable(&internal.Station{})
	if dropErr != nil {
		fmt.Println("エラー: ", dropErr)
	}

	// Migrate the schema
	migrationErr := db.AutoMigrate(&internal.Station{})
	if migrationErr != nil {
		fmt.Println("Migration Error: ", migrationErr)
	}

	Stations := []internal.Station{}

	// ループ処理
	for _, feature := range StationData.Features {
		data := internal.Station{}
		data.Name = feature.Properties.N02_005
		data.Latitude = feature.Geometry.Coordinates[0][1]
		data.Longitude = feature.Geometry.Coordinates[0][0]
		// 名前の重複チェック
		isDuplicate := ContainsStation(Stations, feature.Properties.N02_005)
		if !isDuplicate {
			Stations = append(Stations, data)
			// Batch Create
			createErr := db.Create(&data)
			if createErr != nil {
				fmt.Println("Create Error: ", createErr)
			}
		} else {
			msg := fmt.Sprintf("「%s」は重複しています", feature.Properties.N02_005)
			fmt.Println(msg)
		}
	}

	// JSONファイル出力
	// WriteJson("stations.json", Stations)

	mysql, _ := db.DB()
	mysql.Close()
	fmt.Println("Done!! ⭐️")
}

func WriteJson(fileName string, stations []internal.Station) {
	file, createFileErr := os.Create(fileName)
	if createFileErr != nil {
		fmt.Println("ファイル作成時エラー: ", createFileErr)
		return
	}
	stationsJson, marshalErr := json.Marshal(stations)
	if marshalErr != nil {
		fmt.Println("ファイル解析時エラー: ", marshalErr)
		return
	}
	// ファイル書き込み
	_, writeErr := file.Write(stationsJson)
	if writeErr != nil {
		fmt.Println("ファイル書き込み時エラー: ", writeErr)
		return
	}
}
