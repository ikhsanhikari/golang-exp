#!/bin/bash
APPNAME="molanobar"

all: test build run

build:
	@echo "Building application..."
	@go build -v -o molanobar cmd/serv/main.go
	@echo "😁 Success 😁"

run:
	@go run cmd/serv/main.go cmd/serv/config.go

install:
	@go mod tidy

test:
	@go test ./... -cover -race