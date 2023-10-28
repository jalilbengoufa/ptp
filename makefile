AIR_PATH=$(shell go env GOPATH)/bin/air

build:
	go build -o bin/main main.go

lint:
	gofmt -d **/*.go

run: build
	bin/main

dev:
	$(AIR_PATH)

.PHONY: build lint run dev