.PHONY: build test test-cover lint clean install release-local

BINARY_NAME=springcli
VERSION ?= dev

build:
	go build -ldflags="-X main.Version=$(VERSION)" -o $(BINARY_NAME) .

test:
	go test ./... -v -race

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

lint:
	golangci-lint run

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe coverage.out
	rm -f springcli-linux-* springcli-darwin-* springcli-windows-*

install:
	go install -ldflags="-X main.Version=$(VERSION)" .

release-local:
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o springcli-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o springcli-linux-arm64 .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o springcli-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o springcli-darwin-arm64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o springcli-windows-amd64.exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o springcli-windows-arm64.exe .
