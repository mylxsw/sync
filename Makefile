Version := $(shell date "+%Y%m%d%H%M")
GitCommit := $(shell git rev-parse HEAD)
DIR := $(shell pwd)
LDFLAGS := "-s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)"

dev: doc build
	./bin/sync

dashboard:
	cd dashboard && npm run serve

run: build
	./bin/sync

build:
	go build -ldflags $(LDFLAGS) -o bin/sync *.go

build-release:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o build/release/sync-linux main.go
	cd dashboard && npm run build

protocol-gen:
	protoc --go_out=plugins=grpc:. protocol/*.proto

doc:
	swag init -g api/provider.go

.PHONY: run build protocol-gen doc dashboard dev build-release
