package main

import (
	// "encoding/json"
	"fmt"
	// "io"
	"log"
	"os"

	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	"github.com/joe-black-jb/socket-map-api/internal"
	"github.com/joho/godotenv"
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

var Places = []internal.Place{
	{
		Name:      "カフェ・ベローチェ 京成船橋駅前店",
		Address:   "〒273-0005 千葉県船橋市本町４丁目４４−２５ ルネライラタワー船橋 １Ｆ",
		Latitude:  35.69985840379561,
		Longitude: 139.98670535289517,
	},
	{
		Name:      "ドトールコーヒーショップ 船橋駅南口店",
		Address:   "〒273-0005 千葉県船橋市本町１丁目３−１ 船橋FACE 1F",
		Latitude:  35.70107504987384,
		Longitude: 139.98581438173127,
	},
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

// DynamoDB との接続
// Init() で実行することで、1つのLambdaにつき1度のみ接続処理を実行する
var svc *dynamodb.DynamoDB

var Env = os.Getenv("ENV")
var accessKey = ""
var secretAccessKey = ""

func init() {
	fmt.Println("Init")
	if Env == "" || Env == "local" {
		envErr := godotenv.Load()
		if envErr != nil {
			log.Fatal("Error loading .env file err: ", envErr)
		}
		accessKey = os.Getenv("ACCESS_KEY")
		secretAccessKey = os.Getenv("SECRET_ACCESS_KEY")
	}
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
		Credentials: credentials.NewStaticCredentials(
			accessKey,
			secretAccessKey,
			"",
		),
	}))
	svc = dynamodb.New(sess)
}

func main() {
	// // jsonファイルの読み込み
	// jsonFile, err := os.Open("stations.json")
	// if err != nil {
	// 	fmt.Println("JSONファイルを開けません", err)
	// 	return
	// }
	// defer jsonFile.Close()
	// jsonData, err := io.ReadAll(jsonFile)
	// if err != nil {
	// 	fmt.Println("JSONデータを読み込めません", err)
	// 	return
	// }

	// Stations := []internal.Station{}
	// json.Unmarshal(jsonData, &Stations)

	// // ループ処理
	// for _, station := range Stations {

	// 	// Station 構造体を DynamoDB アイテムに変換
	// 	av, err := dynamodbattribute.MarshalMap(station)
	// 	if err != nil {
	// 		log.Fatalf("failed to marshal station: %v", err)
	// 	}
	// 	// Dynamo テーブルに値を追加
	// 	_, putErr := svc.PutItem(&dynamodb.PutItemInput{
	// 		TableName: aws.String("socket_map_stations"),
	// 		Item:      av,
	// 	})

	// 	if putErr != nil {
	// 		fmt.Println("put Error: ", putErr)
	// 		continue
	// 	}
	// 	successMsg := fmt.Sprintf("「%s」追加完了", station.Name)
	// 	fmt.Println(successMsg)
	// }

	// // goroutine put stations
	// var wg sync.WaitGroup
	// for _, station := range Stations {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		// Station 構造体を DynamoDB アイテムに変換
	// 		av, err := dynamodbattribute.MarshalMap(station)
	// 		if err != nil {
	// 			log.Fatalf("failed to marshal station: %v", err)
	// 		}
	// 		// Dynamo テーブルに値を追加
	// 		_, putErr := svc.PutItem(&dynamodb.PutItemInput{
	// 			TableName: aws.String("socket_map_stations"),
	// 			Item:      av,
	// 		})

	// 		if putErr != nil {
	// 			fmt.Println("put Error: ", putErr)
	// 			continue
	// 		}
	// 		successMsg := fmt.Sprintf("「%s」追加完了", station.Name)
	// 		fmt.Println(successMsg)
	// 	}()
	// }

	// put places
	var placesWg sync.WaitGroup

	for _, place := range Places {
		placesWg.Add(1)
		go func() {
			defer placesWg.Done()
			id, uuidErr := uuid.NewUUID()
			if uuidErr != nil {
				fmt.Println("uuid create error")
			}
			fmt.Println("uuid: ", id)
			place.ID = id.String()

			av, err := dynamodbattribute.MarshalMap(place)
			if err != nil {
				log.Fatalf("failed to marshal place: %v", err)
			}
			// Dynamo テーブルに値を追加
			_, putErr := svc.PutItem(&dynamodb.PutItemInput{
				TableName: aws.String("socket_map_places"),
				Item:      av,
			})

			if putErr != nil {
				fmt.Println("put Error: ", putErr)
				return
			}
			successMsg := fmt.Sprintf("「%s」追加完了", place.Name)
			fmt.Println(successMsg)
		}()
	}
	placesWg.Wait()
	fmt.Println("Done ⭐️")
}
