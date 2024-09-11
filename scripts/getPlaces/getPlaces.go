package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/joe-black-jb/socket-map-api/internal"
	"github.com/joe-black-jb/socket-map-api/internal/api"
	"github.com/joho/godotenv"
	"golang.org/x/net/html"
	"googlemaps.github.io/maps"
)

var Env = os.Getenv("ENV")

var dynamoClient *dynamodb.Client

func init() {
	cfg, cfgErr := config.LoadDefaultConfig(context.TODO())
	if cfgErr != nil {
		fmt.Println("Load default config error: %v", cfgErr)
		return
	}
	dynamoClient = dynamodb.NewFromConfig(cfg)
}

func main() {

	if Env == "" || Env == "local" {
		// Read Environment Variables
		envErr := godotenv.Load()
		if envErr != nil {
			log.Fatal("Error loading .env file err: ", envErr)
		}
		// Set Google Maps API key
		googleMapApiKey := os.Getenv("GOOGLE_MAP_API_KEY")
		if googleMapApiKey == "" {
			log.Fatal("Google Maps の API キーが設定されていません")
		}
		// Create Google Maps client
		client, clientErr := maps.NewClient(maps.WithAPIKey(googleMapApiKey))
		if clientErr != nil {
			log.Fatal("cErr: ", clientErr)
		}
		// Get HTML
		for i := 2; i <= 178; i++ {
			url := fmt.Sprintf("https://www.justnoles.com/page/%d/?search_keywords&search_keywords_operator=and&search_cat2=2087&search_cat3=0", i)
			// fmt.Println("url: ", url)
			// 1ページ目
			// url := "https://www.justnoles.com/?search_keywords=&search_keywords_operator=and&search_cat2=2087&search_cat3=0"
			// 2ページ目
			// url := "https://www.justnoles.com/page/179/?search_keywords&search_keywords_operator=and&search_cat2=2087&search_cat3=0"
			resp, getErr := http.Get(url)
			if getErr != nil {
				fmt.Println("http Get Error: ", getErr)
				return
			}
			defer resp.Body.Close()

			// Parse HTML
			doc, parseErr := html.Parse(resp.Body)
			if parseErr != nil {
				fmt.Println("Parse Error: ", parseErr)
				return
			}
			// f(doc)
			var shopsArray []string
			shopNames := findTitle(doc, shopsArray)
			// fmt.Println("shopNames: ", shopNames)
			// DynamoDB への登録処理
			for _, shopName := range shopNames {
				closed := strings.Contains(shopName, "【閉店】")
				if !closed {
					GetPlacesFromGoogle(dynamoClient, client, shopName)
				}
			}
		}
	} else {
		fmt.Println("環境変数 (ENV) がない、もしくは local 以外の値が設定されています")
	}
}

// func f(n *html.Node) {
// 	// fmt.Println("n.Type: ", n.Type)
// 	// fmt.Println("n.Data: ", n.Data)
// 	if n.Type == html.ElementNode && n.Data == "h3" {
// 		for _, attr := range n.Attr {
// 			fmt.Println("attr: ", attr)
// 			// if attr.Key == "href" {
// 			// 		fmt.Println("Found link:", attr.Val)
// 			// }
// 		}
// 	}
// }

// 関数: 指定されたノードの属性から特定の属性値を取得
func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

// 関数: 再帰的にHTMLノードを探索し、<h3 class="title">を見つける
func findTitle(node *html.Node, shops []string) []string {
	if node.Type == html.ElementNode && node.Data == "h3" {
		// class属性が"title"か確認
		if class := getAttr(node, "class"); strings.Contains(class, "title") {
			// 子ノードを探索してテキストを取得
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					shopName := fmt.Sprintf("マクドナルド %s", c.Data)
					shops = append(shops, shopName)
				}
			}
		}
	}

	// 子ノードを再帰的に探索
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		shops = findTitle(c, shops)
	}
	// fmt.Println("shops: ", shops)
	return shops
}

func GetPlacesFromGoogle(dynamoClient *dynamodb.Client, client *maps.Client, shopName string) {
	// fmt.Println("GetPlacesFromGoogle start ⭐️")

	r := &maps.TextSearchRequest{
		Query: shopName,
	}
	// Google Maps Text Search
	searchResp, searchErr := client.TextSearch(context.TODO(), r)
	if searchErr != nil {
		fmt.Println("searchErr: ", searchErr)
		return
	}
	// fmt.Println("searchResp.Results[0]: ", searchResp.Results[0])
	result := searchResp.Results[0]
	// fmt.Println("searchResp.Results[0].Geometry.Location: ", searchResp.Results[0].Geometry.Location)

	// Create UUID
	id, uuidErr := uuid.NewUUID()
	if uuidErr != nil {
		fmt.Println("uuid create error")
	}
	// fmt.Println("uuid: ", id)

	var place internal.Place

	place.ID = id.String()
	place.Name = shopName
	place.Address = result.FormattedAddress
	place.Latitude = result.Geometry.Location.Lat
	place.Longitude = result.Geometry.Location.Lng
	place.Socket = 1

	postPlace, err := api.PostPlace(dynamoClient, place)
	if err != nil {
		log.Fatalf("failed to marshal place: %v", err)
	}
	fmt.Println(fmt.Sprintf("「%s」登録完了 ⭐️: %v", place.Name, postPlace))
}
