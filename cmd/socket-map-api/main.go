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
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// DynamoDB との接続
// Init() で実行することで、1つのLambdaにつき1度のみ接続処理を実行する
var svc *dynamodb.DynamoDB

func init() {
	fmt.Println("Init")
	sess := session.Must(session.NewSession())
	svc = dynamodb.New(sess)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("Received request: %v\n", request)
	path := request.PathParameters["path"]
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
		fmt.Println("places route svc: ", svc)
		return api.GetPlaces(svc)
	case "stations":
		fmt.Println("stations route svc: ", svc)
		return api.GetStations(svc)
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
