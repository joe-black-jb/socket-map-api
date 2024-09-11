package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joe-black-jb/socket-map-api/internal"
)

var client *dynamodb.Client
var s3Client *s3.Client

func init() {
	fmt.Println("Init")

	cfg, cfgErr := config.LoadDefaultConfig(context.TODO())
	if cfgErr != nil {
		fmt.Println("Load default config error: %v", cfgErr)
		return
	}
	client = dynamodb.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg)
}

/*
DynamoDB の socket_map_places テーブルから
全レコードを取得し、ローカルに places.json を書き出す
*/
func main() {
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
		// dynamodb の places テーブルの内容を取得
		output, err := client.Scan(context.TODO(), scanInput)
		if err != nil {
			fmt.Println(fmt.Sprintf("scan error: %v", err))
			return
		}
		var batch []internal.Place
		err = attributevalue.UnmarshalListOfMaps(output.Items, &batch)
		if err != nil {
			fmt.Println(fmt.Sprintf("unmarshal error: %v", err))
			return
		}
		places = append(places, batch...)
		fmt.Println(places[:3])

		if output.LastEvaluatedKey == nil {
			break
		}
		lastEvaluatedKey = output.LastEvaluatedKey
	}

	// write json
	filename := "places.json"
	// Open or create the file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(fmt.Sprintf("os create error: %v", err))
		return
	}
	defer file.Close()

	// Convert data to JSON format
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // For pretty-printing
	err = encoder.Encode(places)
	if err != nil {
		fmt.Println(fmt.Sprintf("encode error: %v", err))
		return
	}

	fmt.Printf("Data successfully written to %s\n", filename)
}
