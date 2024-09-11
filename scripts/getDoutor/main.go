package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/joe-black-jb/socket-map-api/internal"
)

type count struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type category struct {
	Code        string `json:"code"`
	Image_path  string `json:"image_path"`
	Last_update string `json:"last_update"`
	Level       string `json:"level"`
	Name        string `json:"name"`
}

type coord struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type flagOrText struct {
	Code    string `json:"code"`
	Label   string `json:"label"`
	List_no int    `json:"list_no"`
	Value   bool   `json:"value"`
}

type detail struct {
	Flags []flagOrText
	Texts []flagOrText
}

type item struct {
	Address_code  string     `json:"address_code"`
	Address_name  string     `json:"address_name"`
	Categories    []category `json:"categories"`
	Code          string     `json:"code"`
	Coord         coord      `json:"coord"`
	Details       []detail   `json:"details"`
	External_code string     `json:"external_code"`
	From_date     string     `json:"from_date"`
	Last_update   string     `json:"last_update"`
	List_no       int        `json:"list_no"`
	Name          string     `json:"name"`
	Phone         string     `json:"phone"`
	Postal_code   int        `json:"postal_code"`
	Status        string     `json:"status"`
	To_date       string     `json:"to_date"`
}

type doutor struct {
	Count count  `json:"count"`
	Items []item `json:"items"`
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
	// url := "https://shop.doutor.co.jp/doutor/spot/list?c_d17=1"
	// url := "https://shop.doutor.co.jp/doutor/api/proxy2/shop/list?c_d17=1&exclude-category=05.13&add=detail.image&device=pc&sort=default&random-seed=245247242&datum=wgs84&limit=10&offset=0&ex-code=only.prior&timeStamp=20240829"
	// url := "https://shop.doutor.co.jp/doutor/api/proxy2/shop/list?c_d17=1&page=50&exclude-category=05.13&add=detail.image&device=pc&sort=default&random-seed=245247242&datum=wgs84&limit=20&offset=0&ex-code=only.prior&timeStamp=20240829"
	// 電源あり
	// url := "https://shop.doutor.co.jp/doutor/api/proxy2/shop/list?c_d17=1&page=50&exclude-category=05.13&add=detail.image&device=pc&sort=default&random-seed=245247242&datum=wgs84&limit=500&offset=0&ex-code=only.prior&timeStamp=20240829"
	// 東京都全て
	url := "https://shop.doutor.co.jp/doutor/api/proxy2/shop/list?address=13&page=50&exclude-category=05&add=detail.image&device=pc&sort=default&random-seed=210117245&datum=wgs84&limit=500&offset=0&ex-code=only.prior&timeStamp=20240901"
	resp, httpErr := http.Get(url)

	fmt.Println(url)
	fmt.Println(resp.StatusCode)
	if httpErr != nil {
		log.Fatal("http get error")
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal("read error")
	}
	var Doutor doutor
	json.Unmarshal(body, &Doutor)
	// fmt.Println("Doutor: ", Doutor)
	// fmt.Println("Doutor.items[0]: ", Doutor.Items[0])
	// fmt.Println("Doutor.items[0].Name: ", Doutor.Items[0].Name)

	// jsonファイルに書き出す
	// WriteDoutorJson("doutors.json", Doutor)
	// WriteDoutorJson("tokyo-doutors.json", Doutor)

	// loop
	// var wg sync.WaitGroup
	// fmt.Println("店舗総数: ", len(Doutor.Items))
	for _, item := range Doutor.Items {
		// fmt.Println("item.Name: ", item.Name)
		// wg.Add(1)
		// defer wg.Done()

		///////////// delete ///////////
		// name で検索
		tableName := aws.String("socket_map_places")
		scanItem, scanErr := client.Scan(context.TODO(), &dynamodb.ScanInput{
			TableName: tableName,
			ExpressionAttributeNames: map[string]string{
				"#n": "name",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":name": &types.AttributeValueMemberS{Value: item.Name},
			},
			FilterExpression: aws.String("#n = :name"),
		})
		if scanErr != nil {
			scanErrMsg := fmt.Sprintf("「%s」scan error: %v", item.Name)
			fmt.Println(scanErrMsg)
			return
		}
		// fmt.Println("scanItem 🎾: ", scanItem)

		// fmt.Println("scanItem.Items 🎾: ", scanItem.Items)

		// // scanItem.Items が 0件だったら DB に登録されていない
		// if len(scanItem.Items) > 0 {
		// 	continue
		// }

		//////// 一旦コメントアウト ////////////
		for _, data := range scanItem.Items {
			// fmt.Println("==============")
			// fmt.Println("data: ", data)
			// key := data["id"]
			// fmt.Println("ID: ", key)
			key := map[string]types.AttributeValue{
				"id": data["id"], // プライマリキーの属性名に変更
				// 必要に応じてソートキーも追加
			}
			// fmt.Println("id: ", key)
			// deleteInput := &dynamodb.DeleteItemInput{
			// 	TableName: tableName,
			// 	Key:       key,
			// }
			// _, deleteErr := client.DeleteItem(context.TODO(), deleteInput)
			// if deleteErr != nil {
			// 	deleteNgMsg := fmt.Sprintf("「%s」delete error: %v", item.Name, deleteErr)
			// 	fmt.Println(deleteNgMsg)
			// 	return
			// }
			getItemInput := &dynamodb.GetItemInput{
				TableName: tableName,
				Key:       key,
			}
			// fmt.Println(fmt.Sprintf("「%s」getItem OK ⭐️", item.Name))
			// カフェラミル サンシャインシティ店 が見つからない
			// TODO: getItemResult に戻す
			_, getItemErr := client.GetItem(context.TODO(), getItemInput)
			if getItemErr != nil {
				getItemNgMsg := fmt.Sprintf("「%s」getItem error: %v", item.Name, getItemErr)
				fmt.Println(getItemNgMsg)
				return
			}
			// exists := getItemResult != nil
			// fmt.Println("already exists ?: ", exists)
			// if !exists {
			// 	fmt.Println("登録なし: ", item.Name)
			// }
			// fmt.Println(`getItemResult.Item["socket"]`, getItemResult.Item["socket"])

			place := internal.Place{}

			if HasFreeWiFi(item.Details) {
				place.Wifi = 1
				// wifi = "あり"
			}

			if HasSockets(item.Name, item.Details) {
				place.Socket = 1
				// socket = "あり"
			}

			////////// socket が情報と異なる店舗のみ更新 //////////
			// dynamoSocket := getItemResult.Item["socket"]
			// attributeValue := &types.AttributeValueMemberN{Value: "1"}
			// numberStr := attributeValue.Value
			// number, strconvErr := strconv.Atoi(numberStr)
			// if strconvErr != nil {
			// 	fmt.Println("strconvError converting string to int:", strconvErr)
			// 	continue
			// }

			// if number == 1 && place.Socket == 0 {
			// 	socketStr := strconv.Itoa(place.Socket)
			// 	fmt.Println(item.Name)
			// 	// socket のみ更新
			// 	// プライマリーキーを指定し、あとは更新したいカラムだけ指定する
			// 	updateInput := &dynamodb.UpdateItemInput{
			// 		TableName: tableName,
			// 		Key: key,
			// 		UpdateExpression: aws.String("SET socket = :socket"),
			// 		ExpressionAttributeValues: map[string]types.AttributeValue{
			// 			":socket": &types.AttributeValueMemberN{Value: socketStr},
			// 		},
			// 		ReturnValues: types.ReturnValueUpdatedNew,
			// 	}
			// 	updateResult, updateErr := client.UpdateItem(context.TODO(), updateInput)
			// 	if updateErr != nil {
			// 		fmt.Println(fmt.Sprintf("「%s」update error: %v", item.Name, updateErr))
			// 		continue
			// 	}
			// 	fmt.Println(fmt.Sprintf("「%s」update OK 👍: %v", item.Name, updateResult))
			// }
			////////////////////////////////////////////////////////////

			// fmt.Println("dynamoSocket: ", number)
			// key := map[string]types.AttributeValue{
			// 	"id": data["id"], // プライマリキーの属性名に変更
			// 	// 必要に応じてソートキーも追加
			// }

			// fmt.Println("dynamoSocket: ", dynamoSocket)
			// fmt.Println("place.Socket: ", place.Socket)
			// if getItemResult.Item["socket"]
		}
		//////////////
		//////////////////////////////////

		///////////// post //////////////
		// fmt.Println(item.Name)
		// put item to dynamo
		// place := internal.Place{}
		// place.Name = item.Name
		// place.Address = item.Address_name
		// place.Latitude = item.Coord.Lat
		// place.Longitude = item.Coord.Lon
		// // place.Socket = 1

		// Wifi, コンセント有無確認
		// wifi := "なし"
		// socket := "なし"

		// if HasFreeWiFi(item.Details) {
		// 	place.Wifi = 1
		// 	// wifi = "あり"
		// }

		// if HasSockets(item.Name, item.Details) {
		// 	place.Socket = 1
		// 	// socket = "あり"
		// }

		// fmt.Println(fmt.Sprintf("「%s」wifi: %s, 電源: %s", item.Name, wifi, socket))
		// 「ドトールコーヒーショップ 新宿文化センター通り店」wifi: あり, 電源: なし
		// if item.Name == "ドトールコーヒーショップ 新宿文化センター通り店" {
		// 	fmt.Println(fmt.Sprintf("「%s」wifi: %d, 電源: %d", item.Name, place.Wifi, place.Socket))
		// }

		// result, postErr := api.PostPlace(client, place)
		// if postErr != nil {
		// 	fmt.Println("post err: ", postErr)
		// 	return
		// }
		// fmt.Println("post result: ", result)
		///////////////////////////////////////
	}
	// wg.Wait()
	// fmt.Println("POST 完了⭐️")
}

/*
place.Name ❗️:  ドトールコーヒーショップ 四谷３丁目店
Wifi:  あり
コンセント:  なし

place.Name ❗️:  ドトールコーヒーショップ 祐天寺店
Wifi:  あり
コンセント:  なし

place.Name ❗️:  ドトールコーヒーショップ 国際展示場駅店
Wifi:  なし
コンセント:  なし

place.Name ❗️:  神乃珈琲 赤坂店
Wifi:  なし
コンセント:  なし
*/

func HasFreeWiFi(details []detail) bool {
	for _, detail := range details {
		for _, flag := range detail.Flags {
			if flag.Label == "FREE Wi-Fi" && flag.Value {
				return true
			}
		}
	}
	return false
}

func HasSockets(name string, details []detail) bool {
	for _, detail := range details {
		for _, flag := range detail.Flags {
			if flag.Label == "コンセント" && flag.Value {
				return true
			}
		}
	}
	return false
}

func WriteDoutorJson(fileName string, data interface{}) {
	file, createFileErr := os.Create(fileName)
	if createFileErr != nil {
		fmt.Println("ファイル作成時エラー: ", createFileErr)
		return
	}
	jsonData, marshalErr := json.Marshal(data)
	if marshalErr != nil {
		fmt.Println("ファイル解析時エラー: ", marshalErr)
		return
	}
	// ファイル書き込み
	_, writeErr := file.Write(jsonData)
	if writeErr != nil {
		fmt.Println("ファイル書き込み時エラー: ", writeErr)
		return
	}
}
