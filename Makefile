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
