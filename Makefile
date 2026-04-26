.PHONY: run build test tidy lint clean

BINARY=bin/api
MAIN=./cmd/api

## run: start the development server
run:
	go run $(MAIN)

## build: compile the application binary
build:
	go build -o $(BINARY) $(MAIN)

## test: run all unit tests
test:
	go test -v ./...

## tidy: tidy and verify go modules
tidy:
	go mod tidy
	go mod verify

## lint: run golangci-lint (install separately)
lint:
	golangci-lint run ./...

## clean: remove build artefacts
clean:
	rm -rf bin/
