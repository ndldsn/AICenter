.PHONY: build run dev test clean migrate

APP_NAME=aicenter
BUILD_DIR=./bin

build:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME).exe ./cmd/aicenter

run: build
	$(BUILD_DIR)/$(APP_NAME).exe

dev:
	go run ./cmd/aicenter

test:
	go test -v ./...

clean:
	rm -rf $(BUILD_DIR)
	go clean

migrate:
	go run ./cmd/migrate

deps:
	go mod download
	go mod tidy

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
