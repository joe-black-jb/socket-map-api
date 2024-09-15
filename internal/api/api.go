package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	// "log"
	"net/http"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joe-black-jb/socket-map-api/internal"
)

// func GetPlaces(c *gin.Context) {
// 	Places := &[]internal.Place{}
// 	if err := database.Db.Find(Places).Error; err != nil {
// 		// FormatResponse(c, http.StatusNotFound, err)
// 		c.JSON(http.StatusNotFound, gin.H{"error": err})
// 	}
// 	c.JSON(http.StatusOK, Places)
// 	// FormatResponse(c, http.StatusOK, Places)
// }

func GetPlaces(client *dynamodb.Client) (events.APIGatewayProxyResponse, error) {
	fmt.Println("GetPlaces")

	var places []internal.Place

	// pagination 用
	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		scanInput := &dynamodb.ScanInput{
			TableName: aws.String("socket_map_places"),
		}
		if lastEvaluatedKey != nil {
			scanInput.ExclusiveStartKey = lastEvaluatedKey
		}
		result, err := client.Scan(context.TODO(), scanInput)
		if err != nil {
			fmt.Println("scan err: ", err)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "Scan Error",
			}, err
		}

		var batch []internal.Place
		// 取得したアイテムを Place 構造体に変換
		err = attributevalue.UnmarshalListOfMaps(result.Items, &batch)
		if err != nil {
			fmt.Println("unMarshal err: ", err)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "UnMarshal Error",
			}, err
		}

		places = append(places, batch...)

		if result.LastEvaluatedKey == nil {
			break
		}
		lastEvaluatedKey = result.LastEvaluatedKey
	}

	// places のスライスを JSON にシリアライズ
	body, err := json.Marshal(places)
	if err != nil {
		fmt.Println("failed to marshal places to json: ", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Marshal Error",
		}, err
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(body),
		Headers: map[string]string{
			"Content-type": "application/json",
		},
	}, nil
}

func GetPlacesFromCF(req events.APIGatewayProxyRequest, client *s3.Client) (events.APIGatewayProxyResponse, error) {
	fmt.Println("GetPlacesFromS3")
	cfDomain := os.Getenv("CF_DOMAIN")
	if cfDomain == "" {
		log.Fatal("CF のディストリビューションドメインが未設定です")
	}
	key := req.QueryStringParameters["key"]

	// CF 経由で S3 からファイルを取得
	url := fmt.Sprintf("%s/%s", cfDomain, key)

	// Make the GET request
	resp, err := http.Get(url)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "http get Error",
		}, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "read Error",
		}, err
	}

	var places []internal.Place
	json.Unmarshal(body, &places)

	jsonData, err := json.Marshal(places)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(jsonData),
		Headers: map[string]string{
			"Content-type": "application/json",
		},
	}, err
}

// func PostPlace(c *gin.Context) {
// 	var place internal.Place
// 	if err := c.BindJSON(&place); err != nil {
// 		fmt.Println("エラー発生❗️: ", err)
// 		FormatResponse(c, http.StatusBadRequest, err)
// 		return
// 	}
// 	fmt.Println("place⭐️: ", place)
// 	result := database.Db.Create(&place)
// 	fmt.Println("result ⭐️: ", result)
// 	FormatResponse(c, http.StatusOK, place)
// 	// fmt.Println("place.Name: ", place.Name)
// }

func PostPlace(client *dynamodb.Client, place internal.Place) (events.APIGatewayProxyResponse, error) {
	// fmt.Println("PostPlace")

	id, uuidErr := uuid.NewUUID()
	if uuidErr != nil {
		fmt.Println("uuid create error")
	}
	place.ID = id.String()

	item, marshalErr := attributevalue.MarshalMap(place)
	if marshalErr != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Marshal Error",
		}, marshalErr
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("socket_map_places"),
		Item:      item,
	}
	// 終わったら _ を result にする
	result, putErr := client.PutItem(context.TODO(), input)
	if putErr != nil {
		fmt.Println("put err: ", putErr)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "put Error",
		}, putErr
	}
	// fmt.Println("putItem result ⭐️: ", result)
	// doneMsg := fmt.Sprintf("「%s」登録完了⭐️", place.Name)
	doneMsg := fmt.Sprintf("「%s」result %v ⭐️", place.Name, result)
	fmt.Println(doneMsg)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "OK",
	}, putErr
}

func SearchPlace(c *gin.Context) {
	q := c.Query("q")
	fmt.Println("q 🎾: ", q)
	fmt.Println("&q 🎾: ", &q)
	if q == "" {
		errStr := "検索したい場所の名前を指定してください"
		fmt.Println(errStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": errStr})
		return
	}
	encodedQ := url.QueryEscape(q)

	var result []internal.OsmPlaceDetail
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json", encodedQ)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("エラー❗️: ", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading the response body:", err)
		return
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Error while unmarshaling the response ❗️:", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	// fmt.Println("レスポンス⭐️: ", resp)
	fmt.Println("result ⭐️: ", result)
	c.JSON(http.StatusOK, result)
}

// func GetStations(c *gin.Context) {
// 	Stations := &[]internal.Station{}
// 	if err := database.Db.Find(Stations).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
// 	}
// 	c.JSON(http.StatusOK, Stations)
// }

func GetStations(client *dynamodb.Client) (events.APIGatewayProxyResponse, error) {
	fmt.Println("GetStations")

	input := &dynamodb.ScanInput{
		TableName: aws.String("socket_map_stations"),
	}
	result, scanErr := client.Scan(context.TODO(), input)
	if scanErr != nil {
		fmt.Println("scan err: ", scanErr)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Scan Error",
			Headers: map[string]string{
				"Content-type": "application/json",
			},
		}, scanErr
	}

	// 取得したアイテムを Station 構造体に変換
	var stations []internal.Station
	unMarshalErr := attributevalue.UnmarshalListOfMaps(result.Items, &stations)
	if unMarshalErr != nil {
		fmt.Println("unMarshal err: ", unMarshalErr)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "UnMarshal Error",
		}, unMarshalErr
	}

	// stations のスライスを JSON にシリアライズ
	body, marshalErr := json.Marshal(stations)
	if marshalErr != nil {
		fmt.Println("failed to marshal stations to json: ", marshalErr)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Marshal Error",
		}, marshalErr
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(body),
	}, nil
}

func GetPlacesWithBounds(req events.APIGatewayProxyRequest, client *dynamodb.Client)(events.APIGatewayProxyResponse, error){
	latMin, err := strconv.ParseFloat(req.QueryStringParameters["lat_min"], 64)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "invalid parameter",
		}, nil
	}
	latMax, err := strconv.ParseFloat(req.QueryStringParameters["lat_max"], 64)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "invalid parameter",
		}, nil
	}
	lngMin, err := strconv.ParseFloat(req.QueryStringParameters["lng_min"], 64)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "invalid parameter",
		}, nil
	}
	lngMax, err := strconv.ParseFloat(req.QueryStringParameters["lng_max"], 64)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "invalid parameter",
		}, nil
	}

	fmt.Println("query: ", req.QueryStringParameters)

	var response *dynamodb.ScanOutput

	var places []internal.Place

	latCondition := expression.Name("latitude").Between(expression.Value(latMin), expression.Value(latMax))
	lngCondition := expression.Name("longitude").Between(expression.Value(lngMin), expression.Value(lngMax))
	filtEx := expression.And(latCondition, lngCondition)

	expr, err := expression.NewBuilder().WithFilter(filtEx).Build()
	if err != nil {
		fmt.Println("expression build err: ", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "expression build err",
		}, err
	}

	scanPaginator := dynamodb.NewScanPaginator(client, &dynamodb.ScanInput{
		TableName:                 aws.String("socket_map_places"),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		// ProjectionExpression:      expr.Projection(),
	})
	for scanPaginator.HasMorePages() {
		response, err = scanPaginator.NextPage(context.TODO())
		if err != nil {
			log.Printf("Couldn't scan for movies released between %v and %v. Here's why: %v\n", latMin, latMax, err)
			break
		} else {
			var placePage []internal.Place
			err = attributevalue.UnmarshalListOfMaps(response.Items, &placePage)
			if err != nil {
				log.Printf("Couldn't unmarshal query response. Here's why: %v\n", err)
				break
			} else {
				places = append(places, placePage...)
			}
		}
	}
	body, err := json.Marshal(places)
	if err != nil {
		fmt.Println("failed to marshal places to json: ", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Marshal Error",
		}, err
	}
	fmt.Println("len(places): ", len(places))
	fmt.Println("body string: ", string(body))
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(body),
		Headers: map[string]string{
			"Content-type": "application/json",
		},
	}, nil
}
