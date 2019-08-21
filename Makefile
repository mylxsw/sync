Version := $(shell date "+%Y%m%d%H%M")
GitCommit := $(shell git rev-parse HEAD)
DIR := $(shell pwd)
LDFLAGS := "-s -w -X main.Version=$(Version) -X main.GitCommit=$(GitCommit)"

run: build
	./bin/sync

build:
	go build -ldflags $(LDFLAGS) -o bin/sync *.go

protocol-gen:
	protoc --go_out=plugins=grpc:. protocol/*.proto

.PHONY: run build protocol-gen
