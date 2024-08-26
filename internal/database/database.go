package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Db *gorm.DB

var ProdDb *sql.DB

func Connect() {
	fmt.Println("DBとの接続処理⭐️")

	Env := os.Getenv("ENV")
	fmt.Println("環境: ", Env)

	if Env == "" || Env == "local" {
		fmt.Println("Local ⭐️")
		envErr := godotenv.Load()
		if envErr != nil {
			log.Fatal("Error loading .env file err: ", envErr)
		}
		DbUser := os.Getenv("MYSQL_USER")
		DbPass := os.Getenv("MYSQL_ROOT_PASSWORD")
		DbName := os.Getenv("MYSQL_DATABASE")
		DbHost := os.Getenv("MYSQL_HOST")

		// docker コンテナを立ち上げている場合、ホスト名は 127.0.0.1 ではなくサービス名（db）
		dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", DbUser, DbPass, DbHost, DbName)
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal("failed to connect database")
		}

		fmt.Println("Connected to Local Database ⭐️")
		Db = db
	}

	if Env == "production" {
		if os.Getenv("USE_DOT_ENV") == "true" {
			envErr := godotenv.Load()
			if envErr != nil {
				log.Fatal("Error loading .env file err: ", envErr)
			}
		}
		ProdDbEndpoint := os.Getenv("PROD_DB_ENDPOINT")
		ProdDbName := os.Getenv("PROD_DB_NAME")
		secretName := os.Getenv("PROD_DB_SECRET_NAME")
		region := os.Getenv("REGION")

		config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
		if err != nil {
			fmt.Println("LoadDefaultConfig Err")
			log.Fatal(err)
		}

		// Create Secrets Manager client
		svc := secretsmanager.NewFromConfig(config)

		input := &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(secretName),
			VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
		}

		result, err := svc.GetSecretValue(context.TODO(), input)
		if err != nil {
			// For a list of exceptions thrown, see
			// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_GetSecretValue.html
			fmt.Println("GetSecretValue err")
			log.Fatal(err.Error())
		}

		fmt.Println("result: ", result)

		// Decrypts secret using the associated KMS key.
		var secretString string = *result.SecretString

		type DbSecret struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		dbSecret := DbSecret{}

		json.Unmarshal([]byte(secretString), &dbSecret)

		// Your code goes here.

		// Connect DB
		// dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbSecret.Username, dbSecret.Password, ProdDbEndpoint, ProdDbName)
		// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		// if err != nil {
		// 	log.Fatal("failed to connect database")
		// }

		// https://docs.aws.amazon.com/ja_jp/AmazonRDS/latest/UserGuide/UsingWithRDS.IAMDBAuth.Connecting.Go.html
		var dbName string = ProdDbName
		var dbUser string = dbSecret.Username
		var dbHost string = ProdDbEndpoint
		var dbPort int = 3306
		var dbEndpoint string = fmt.Sprintf("%s:%d", dbHost, dbPort)

		authenticationToken, err := auth.BuildAuthToken(
			context.TODO(), dbEndpoint, region, dbUser, config.Credentials)
		if err != nil {
			fmt.Println("BuildAuthToken Err")
			panic("failed to create authentication token: " + err.Error())
		}

		// dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=true&allowCleartextPasswords=true",
		// 	dbUser, authenticationToken, dbEndpoint, dbName,
		// )
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUser, authenticationToken, dbEndpoint, dbName,
		)

		fmt.Println("本番用DBとの接続を開始します")
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}

		err = db.Ping()
		if err != nil {
			panic(err)
		}

		fmt.Println("Connected to Local Database ⭐️")
		ProdDb = db
	}
}
