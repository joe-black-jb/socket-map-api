.PHONY: zip terraform localstack lint fmt

# 変数
GO := go
APP_DIR := ./scripts
STATION_APP_DIR := ./scripts/migrateStations
APP_NAME := migration
STATION_APP_NAME := migrateStations

migrate:
	$(GO) run $(APP_DIR)/$(APP_NAME).go

lint:
	go vet ./...

fmt:
	go fmt ./...

station:
	$(GO) run $(STATION_APP_DIR)/$(STATION_APP_NAME).go

deploy:
	sh ./scripts/deploy.sh $(TARGET)

dynamo:
	$(GO) run ./scripts/migrateToDynamo/migrateToDynamo.go

getPlaces:
	$(GO) run ./scripts/getPlaces/getPlaces.go

getDoutor:
	$(GO) run ./scripts/getDoutor/main.go

mcdonalds:
	$(GO) run ./scripts/mcdonalds/postMcdonalds.go

starbucks:
	$(GO) run ./scripts/starbucks/postStarbucks.go

writePlaces:
	$(GO) run ./scripts/writePlacesJson/writePlacesJson.go

localstack:
	@docker-compose up -d

down:
	@docker-compose down

start:
	@docker-compose start

stop:
	@docker-compose stop

zip:
	@GOOS=linux GOARCH=amd64 go build -o terraform/localstack/main cmd/socket-map-api/main.go
	@zip terraform/localstack/main.zip terraform/localstack/main
	@rm terraform/localstack/main

terraform:
	@tflocal -chdir=terraform/localStack init
	@tflocal -chdir=terraform/localStack apply --auto-approve


