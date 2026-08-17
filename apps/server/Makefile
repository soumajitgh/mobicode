MIGRATIONS_DIR := ./internal/migrations/sql
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@v3.27.3
AIR ?= air
GOIMPORTS ?= goimports
GOLANGCI_LINT ?= golangci-lint
BUILD_OUTPUT ?= ../../dist/apps/server/server

.PHONY: dev run build test vet format lint check graphql-generate graphql-check migrate-create

dev:
	$(AIR)

run:
	go run ./cmd/api

build:
	mkdir -p $$(dirname $(BUILD_OUTPUT))
	go build -o $(BUILD_OUTPUT) ./cmd/api

test:
	go test ./...

vet:
	go vet ./...

format:
	$(GOIMPORTS) -w $$(find . -name '*.go' -not -path './vendor/*')

lint:
	$(GOLANGCI_LINT) run ./...

check: format vet lint test

graphql-generate:
	go run github.com/99designs/gqlgen generate

graphql-check:
	go run github.com/99designs/gqlgen generate
	git diff --exit-code -- internal/graphql

migrate-create:
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(name) sql
