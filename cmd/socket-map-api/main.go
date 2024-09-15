package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/joe-black-jb/socket-map-api/internal/api"
	// "github.com/joe-black-jb/socket-map-api/internal/database"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// DynamoDB との接続
// Init() で実行することで、1つのLambdaにつき1度のみ接続処理を実行する
// var svc *dynamodb.DynamoDB
var dynamoClient *dynamodb.Client

var s3Client *s3.Client

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
	dynamoClient = dynamodb.NewFromConfig(cfg)

	s3Client = s3.NewFromConfig(cfg)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("Received request: %v\n", request)
	path := request.PathParameters["path"]
	key := request.QueryStringParameters["key"]
	fmt.Println("path: ", path)
	fmt.Println("request.Headers: ", request.Headers)

	requestJSON, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Failed to marshal request to JSON")
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       string(err.Error()),
		}, err
	}

	// Print or log the JSON representation of the request
	fmt.Printf("Request JSON: %s\n", requestJSON)

	// Routing
	switch path {
	case "places":
		if key != "" {
			fmt.Println("places route with key: ", key)
			return api.GetPlacesFromCF(request, s3Client)
		}
		fmt.Println("places route dynamoClient: ", dynamoClient)
		return api.GetPlaces(dynamoClient)
	case "stations":
		if key != "" {
			fmt.Println("stations route with key: ", key)
			return api.GetPlacesFromCF(request, s3Client)
		}
		fmt.Println("stations route dynamoClient: ", dynamoClient)
		return api.GetStations(dynamoClient)
	// case "places-cf":
	// 	fmt.Println("stations route dynamoClient: ", dynamoClient)
	// 	return api.GetPlacesFromCF(request, s3Client)
	case "places-bounds":
		fmt.Println("places-bounds route dynamoClient: ", dynamoClient)
		return api.GetPlacesWithBounds(request, dynamoClient)
	default:
		fmt.Println("default")
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string("OK"),
	}, nil
}

func main() {
	fmt.Println("Hello World!")
	// DB接続
	// database.Connect()

	// ルーター起動
	// api.Router()

	// ハンドラー関数実行
	lambda.Start(handler)
}
