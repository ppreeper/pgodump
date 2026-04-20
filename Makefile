BINARY     := pgodump
CMD        := ./cmd/pgodump
PREFIX     ?= $(shell go env GOPATH)
BINDIR     := $(PREFIX)/bin

.PHONY: all build install clean test test-unit test-integration vet lint

all: build

## build: compile the binary (version from debug.ReadBuildInfo)
build:
	go build -o $(BINARY) $(CMD)

## install: build and install to GOPATH/bin
install:
	go install $(CMD)

## clean: remove build artefacts
clean:
	rm -f $(BINARY)

## vet: run go vet
vet:
	go vet ./...

## lint: run golangci-lint (requires golangci-lint in PATH)
lint:
	golangci-lint run ./...

## test: run all tests (unit + integration)
test:
	go test -v -race -timeout 120s ./...

## test-unit: run only unit tests (no testcontainers)
test-unit:
	go test -v -race -run "^Test[^I]" -timeout 30s ./...

## test-integration: run only integration/container tests
test-integration:
	go test -v -race -run "Integration|Migration|Inheritance" -timeout 120s ./...

## help: print this help
help:
	@grep -E '^##' Makefile | sed 's/## /  /'
