MIGRATIONS_DIR := ./internal/migrations/sql
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@v3.27.3
GQLGEN := go run github.com/99designs/gqlgen@v0.17.94
AIR ?= air
GOIMPORTS ?= goimports
GOLANGCI_LINT ?= golangci-lint
BUILD_OUTPUT ?= build/mobicode

.PHONY: dev run build clean test vet format lint check graphql-generate graphql-check \
	migrate-create deps-up deps-down deps-logs mobile-install mobile-dev mobile-check

dev: deps-up
	$(AIR)

run: deps-up
	go run ./cmd/mobicode serve

build:
	mkdir -p $$(dirname $(BUILD_OUTPUT))
	go build -o $(BUILD_OUTPUT) ./cmd/mobicode

clean:
	go clean ./...
	rm -f $(BUILD_OUTPUT)

test:
	go test ./...

vet:
	go vet ./...

format:
	$(GOIMPORTS) -w $$(find cmd internal -name '*.go' -not -path './vendor/*')

lint:
	$(GOLANGCI_LINT) run ./...

check: format vet lint test

graphql-generate:
	$(GQLGEN) generate

graphql-check:
	$(GQLGEN) generate
	git diff --exit-code -- internal/graphql

migrate-create:
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(name) sql

deps-up:
	docker compose up -d libsql

deps-down:
	docker compose down

deps-logs:
	docker compose logs -f libsql

mobile-install:
	pnpm --dir mobile install --frozen-lockfile

mobile-dev:
	pnpm --dir mobile start

mobile-check:
	pnpm --dir mobile typecheck
