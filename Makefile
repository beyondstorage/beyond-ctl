SHELL := /bin/bash

.PHONY: all check format vet build test generate tidy release

-include Makefile.env

VERSION := v0.0.1

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GO_BUILD := CGO_ENABLED=0 go build -ldflags "-X main.Version=${VERSION}"

help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo "  check               to do static check"
	@echo "  build               to create bin directory and build"
	@echo "  test                to run test"

format:
	go fmt ./...

vet:
	go vet ./...

build: tidy format vet
	${GO_BUILD} -o bin/byctl ./cmd/byctl

release:
	mkdir -p ./release

	GOOS=${GOOS} GOARCH=${GOARCH} ${GO_BUILD} -o ./bin/${GOOS}_${GOARCH}/byctl_${VERSION}_${GOOS}_${GOARCH} ./cmd/byctl
	tar -C ./bin/${GOOS}_${GOARCH}/ -czf ./release/byctl_${VERSION}_${GOOS}_${GOARCH}.tar.gz byctl_${VERSION}_${GOOS}_${GOARCH}

release-linux-amd64: GOOS := linux
release-linux-amd64: GOARCH := amd64
release-linux-amd64: release

release-darwin-amd64: GOOS := darwin
release-darwin-amd64: GOARCH := amd64
release-darwin-amd64: release

release-windows-amd64: GOOS := windows
release-windows-amd64: GOARCH := amd64
release-windows-amd64: release

release-all: release-linux-amd64 release-darwin-amd64 release-windows-amd64

test:
	BEYOND_CTL_INTEGRATION_TEST=off go test -race -coverprofile=coverage.txt -covermode=atomic -v ./...
	go tool cover -html="coverage.txt" -o "coverage.html"

integration_test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -v ./cmd/byctl

tidy:
	go mod tidy
	go mod verify
