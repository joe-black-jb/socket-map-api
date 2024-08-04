package main

import (
	"fmt"

	"github.com/joe-black-jb/socket-map-api/internal/api"
	"github.com/joe-black-jb/socket-map-api/internal/database"
)

func main() {
	fmt.Println("Hello World!")
	// DB接続
	database.Connect()
	// ルーター起動
	api.Router()
}