package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joe-black-jb/socket-map-api/internal"
	"github.com/joe-black-jb/socket-map-api/internal/api"
	"golang.org/x/net/html"
)

type Field struct {
	Name                     string `json:"name" dynamodbav:"name"`
	Address_5                string `json:"address_5" dynamodbav:"address_5"`
	Location                 string `json:"location" dynamodbav:"location"`
	PublicWirelessServiceFlg string `json:"public_wireless_service_flg" dynamodbav:"public_wireless_service_flg"`
}

type Starbucks struct {
	Id     string `json:"id" dynamodbav:"id"`
	Fields Field  `json:"fields" dynamodbav:"fields"`
}

var client *dynamodb.Client

func init() {
	fmt.Println("Init")
	cfg, cfgErr := config.LoadDefaultConfig(context.TODO())
	if cfgErr != nil {
		fmt.Println("Load default config error: %v", cfgErr)
		return
	}
	client = dynamodb.NewFromConfig(cfg)
}

func main() {
	// seeder ファイルの読み込み
	file, err := os.Open("seeder/starbucks.json")
	if err != nil {
		fmt.Println("file open err: ", err)
		return
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("read err: ", err)
		return
	}
	var StarbucksData []Starbucks
	json.Unmarshal(bytes, &StarbucksData)

	for _, item := range StarbucksData {
		shopName := item.Fields.Name
		// 電源、Wifi の有無を確認
		url := fmt.Sprintf("https://www.h9v.net/%s/", shopName)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(fmt.Sprintf("「%s」http get error: %v", shopName, err))
			continue
		}
		defer resp.Body.Close()
		// Parse HTML
		htmlData, err := html.Parse(resp.Body)
		if err != nil {
			fmt.Println(fmt.Sprintf("「%s」html parse error: %v", shopName, err))
			continue
		}
		// 説明文を取得
		var text string
		text = extractMetaContent(htmlData, text)
		latLng := strings.Split(item.Fields.Location, ",")
		lat, err := strconv.ParseFloat(latLng[0], 64)
		if err != nil {
			fmt.Println(fmt.Sprintf("「%s」の緯度変換エラー: 緯度「%v」", shopName, lat))
			continue
		}
		lng, err := strconv.ParseFloat(latLng[1], 64)
		if err != nil {
			fmt.Println(fmt.Sprintf("「%s」の緯度変換エラー: 緯度「%v」", shopName, lng))
			continue
		}

		place := internal.Place{}
		place.Name = "スターバックス " + shopName
		place.Address = item.Fields.Address_5
		place.Latitude = lat
		place.Longitude = lng

		if strings.Contains(text, "電源が使える") {
			place.Socket = 1
		}

		if text == "" {
			place.Socket = 2
		}

		if item.Fields.PublicWirelessServiceFlg == "1" {
			place.Wifi = 1
		}

		// 登録
		postPlace, err := api.PostPlace(client, place)
		if err != nil {
			fmt.Println(fmt.Sprintf("「%s」の putItem エラー: %v", shopName, err))
			continue
		}
		fmt.Println(fmt.Sprintf("「%s」登録完了 ⭐️: %v", place.Name, postPlace))
	}
}

// HTMLドキュメントから特定のmetaタグのcontent属性を抽出する関数
func extractMetaContent(n *html.Node, text string) string {
	// metaタグかつdescription属性を持っているかを確認
	if n.Type == html.ElementNode && n.Data == "meta" {
		var name, content string
		for _, attr := range n.Attr {
			if attr.Key == "name" && attr.Val == "description" {
				name = attr.Val
			}
			if attr.Key == "content" {
				content = attr.Val
			}
		}
		if name == "description" && strings.Contains(content, "営業時間") && strings.Contains(content, "電源が") && strings.Contains(content, "Wi-Fiが") {
			// fmt.Println("content: ", content)
			text = content
		}
	}
	// 子ノードを再帰的に探索
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text = extractMetaContent(c, text)
		// fmt.Println("content: ", content)
	}
	return text
}
