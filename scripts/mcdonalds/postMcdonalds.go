package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/joe-black-jb/socket-map-api/internal"
	"github.com/joe-black-jb/socket-map-api/internal/api"
)

type Mcdonald struct {
	ID               int     `json:"id" dynamodbav:"id"`
	Key              string  `json:"key" dynamodbav:"key"`
	Name             string  `json:"name" dynamodbav:"name"`
	Latitude         float64 `json:"latitude" dynamodbav:"latitude"`
	Longitude        float64 `json:"longitude" dynamodbav:"longitude"`
	Address          string  `json:"address" dynamodbav:"address"`
	Marker_index     int     `json:"marker_index" dynamodbav:"marker_index"`
	Condition_values []int   `json:"condition_values" dynamodbav:"condition_values"`
}

var client *dynamodb.Client

func init() {
	fmt.Println("Init")
	// v1
	// sess := session.Must(session.NewSession())
	// svc = dynamodb.New(sess)

	// v2
	cfg, cfgErr := config.LoadDefaultConfig(context.TODO())
	if cfgErr != nil {
		fmt.Println("Load default config error: %v", cfgErr)
		return
	}
	client = dynamodb.NewFromConfig(cfg)
}

func main() {
	// mc.json の読み込み
	file, openErr := os.Open("seeder/mc.json")
	if openErr != nil {
		fmt.Println("file open err: ", openErr)
		return
	}
	defer file.Close()
	fmt.Println("file: ", file)
	bytes, readErr := io.ReadAll(file)
	if readErr != nil {
		fmt.Println("read err: ", readErr)
		return
	}
	var Mcdonalds []Mcdonald
	json.Unmarshal(bytes, &Mcdonalds)

	for _, item := range Mcdonalds {
		// fmt.Println(item)

		shopName := fmt.Sprintf("マクドナルド %s", item.Name)
		// 存在チェック (先頭に「マクドナルド 」をつける)
		tableName := aws.String("socket_map_places")
		scanItem, scanErr := client.Scan(context.TODO(), &dynamodb.ScanInput{
			TableName: tableName,
			ExpressionAttributeNames: map[string]string{
				"#n": "name",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":name": &types.AttributeValueMemberS{Value: shopName},
			},
			FilterExpression: aws.String("#n = :name"),
		})
		if scanErr != nil {
			scanErrMsg := fmt.Sprintf("「%s」scan error: %v", item.Name, scanErr)
			fmt.Println(scanErrMsg)
			return
		}
		// fmt.Println(fmt.Sprintf("「%v」の Items: %v ⭐️", shopName, scanItem.Items))

		// 未登録店舗の登録
		if len(scanItem.Items) == 0 {
			// fmt.Println(item)
			place := internal.Place{}
			place.Name = shopName
			place.Address = item.Address
			place.Latitude = item.Latitude
			place.Longitude = item.Longitude
			// 未登録店舗は 電源がないので Socket = 0 を設定
			place.Socket = 0
			// wifi があるかの判定
			hasWifi := HasWifiMc(item.Condition_values)
			if hasWifi {
				place.Wifi = 1
			}

			// fmt.Println(fmt.Sprintf("「%s」はまだ登録されていません ⭐️ wifi: %d", shopName, place.Wifi))
			result, postErr := api.PostPlace(client, place)
			if postErr != nil {
				fmt.Println("post err: ", postErr)
				return
			}
			fmt.Println(fmt.Sprintf("「%s」登録完了 ⭐️ result: %v", shopName, result))
			continue
		}

		// 登録済み店舗の更新
		//////////////////////////////////////////
		for _, data := range scanItem.Items {
			key := map[string]types.AttributeValue{
				"id": data["id"], // プライマリキーの属性名に変更
			}
			getItemInput := &dynamodb.GetItemInput{
				TableName: tableName,
				Key:       key,
			}
			_, getItemErr := client.GetItem(context.TODO(), getItemInput)
			if getItemErr != nil {
				getItemNgMsg := fmt.Sprintf("「%s」getItem error: %v", item.Name, getItemErr)
				fmt.Println(getItemNgMsg)
				return
			}
			// fmt.Println(fmt.Sprintf("「%v」getItem OK ⭐️", getItemResult))
			// fmt.Println(fmt.Sprintf("「%v」getItem OK ⭐️", getItemResult.Item["name"]))
			// wifi があるかの判定
			hasWifi := HasWifiMc(item.Condition_values)
			// fmt.Println("Wifi はありますか❓: ", hasWifi)
			if hasWifi {
				// wifi のみ更新
				wifiStr := strconv.Itoa(1)
				// プライマリーキーを指定し、あとは更新したいカラムだけ指定する
				updateInput := &dynamodb.UpdateItemInput{
					TableName:        tableName,
					Key:              key,
					UpdateExpression: aws.String("SET wifi = :wifi"),
					ExpressionAttributeValues: map[string]types.AttributeValue{
						":wifi": &types.AttributeValueMemberN{Value: wifiStr},
					},
					ReturnValues: types.ReturnValueUpdatedNew,
				}
				updateResult, updateErr := client.UpdateItem(context.TODO(), updateInput)
				if updateErr != nil {
					fmt.Println(fmt.Sprintf("「%s」update error: %v", item.Name, updateErr))
					continue
				}
				fmt.Println(fmt.Sprintf("「%s」更新完了 👍: %v", shopName, updateResult))
				continue
			}
		}
		//////////////////////////////////////////
	}
}

func HasWifiMc(values []int) bool {
	// fmt.Println("values: ", values)
	wifi := values[0]
	if wifi == 0 {
		return false
	}
	if wifi == 1 {
		return true
	}
	return false
}
