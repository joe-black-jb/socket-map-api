package api

import (
	"encoding/json"
	"fmt"
	"io"

	// "log"
	"net/http"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gin-gonic/gin"
	"github.com/joe-black-jb/socket-map-api/internal"
	"github.com/joe-black-jb/socket-map-api/internal/database"
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

func GetPlaces(svc *dynamodb.DynamoDB) (events.APIGatewayProxyResponse, error) {
	fmt.Println("GetPlaces")

	input := &dynamodb.ScanInput{
		TableName: aws.String("socket_map_places"),
	}
	result, scanErr := svc.Scan(input)
	if scanErr != nil {
		fmt.Println("scan err: ", scanErr)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Scan Error",
		}, scanErr
	}

	// 取得したアイテムを Place 構造体に変換
	var places []internal.Place
	unMarshalErr := dynamodbattribute.UnmarshalListOfMaps(result.Items, &places)
	if unMarshalErr != nil {
		fmt.Println("unMarshal err: ", unMarshalErr)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "UnMarshal Error",
		}, unMarshalErr
	}

	// places のスライスを JSON にシリアライズ
	body, marshalErr := json.Marshal(places)
	if marshalErr != nil {
		fmt.Println("failed to marshal places to json: ", marshalErr)
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

func PostPlace(c *gin.Context) {
	var place internal.Place
	if err := c.BindJSON(&place); err != nil {
		fmt.Println("エラー発生❗️: ", err)
		FormatResponse(c, http.StatusBadRequest, err)
		return
	}
	fmt.Println("place⭐️: ", place)
	result := database.Db.Create(&place)
	fmt.Println("result ⭐️: ", result)
	FormatResponse(c, http.StatusOK, place)
	// fmt.Println("place.Name: ", place.Name)

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

func GetStations(svc *dynamodb.DynamoDB) (events.APIGatewayProxyResponse, error) {
	fmt.Println("GetStations")

	input := &dynamodb.ScanInput{
		TableName: aws.String("socket_map_stations"),
	}
	result, scanErr := svc.Scan(input)
	if scanErr != nil {
		fmt.Println("scan err: ", scanErr)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Scan Error",
		}, scanErr
	}

	// 取得したアイテムを Station 構造体に変換
	var stations []internal.Station
	unMarshalErr := dynamodbattribute.UnmarshalListOfMaps(result.Items, &stations)
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
