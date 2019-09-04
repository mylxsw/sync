Version := $(shell date "+%Y%m%d%H%M")
GitCommit := $(shell git rev-parse HEAD)
DIR := $(shell pwd)
LDFLAGS := "-s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)"

run: static-gen doc-gen build
	./bin/sync

run-dashboard:
	cd dashboard && npm run serve

build:
	go build -ldflags $(LDFLAGS) -o bin/sync *.go

build-release: build-dashboard static-gen
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o build/release/sync-linux main.go

build-dashboard:
	cd dashboard && npm run build

static-gen:
	esc -pkg api -o api/static.go -prefix=dashboard/dist dashboard/dist

protocol-gen:
	protoc --go_out=plugins=grpc:. protocol/*.proto

doc-gen:
	swag init -g api/provider.go

.PHONY: run build protocol-gen doc-gen run-dashboard build-release build-dashboard static-gen

