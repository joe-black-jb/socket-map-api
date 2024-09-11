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
	sh ./scripts/deploy.sh

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
