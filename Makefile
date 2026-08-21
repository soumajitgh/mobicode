MIGRATIONS_DIR := ./internal/migrations/sql
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@v3.27.3
GQLGEN := go run github.com/99designs/gqlgen@v0.17.94
AIR ?= air
GOIMPORTS ?= goimports
GOLANGCI_LINT ?= golangci-lint
BUILD_OUTPUT ?= build/mobicode
CONFIG ?= .env
DEV_API_URL ?= http://localhost:8085
DEV_GRAPHQL_BODY ?= {"query":"query Viewer { viewer { publicKey } }"}

.PHONY: dev run init build clean test vet format lint check graphql-generate graphql-check \
	migrate-create deps-up deps-down deps-logs mobile-install mobile-dev mobile-check \
	dev-setup dev-init dev-run dev-sign dev-query

init:
	go run ./cmd/mobicode init

dev: deps-up init
	$(AIR)

run: deps-up init
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

# Creates ignored, disposable local development identities. It refuses to overwrite
# existing local configuration so a real key cannot be replaced accidentally.
dev-setup:
	@test ! -e $(CONFIG) || (echo "$(CONFIG) already exists; update it manually or remove it before rerunning make dev-setup." && exit 1)
	@test ! -e mobile/.env || (echo "mobile/.env already exists; update it manually or remove it before rerunning make dev-setup." && exit 1)
	@NSEC="$$(go run ./cmd/mobicode dev generate-nsec --value-only)"; \
	printf 'PORT=8085\nENV=development\nDATABASE_PATH=data/app.db\nPUBLIC_BASE_URL=$(DEV_API_URL)\nDEV_NSEC=%s\n' "$$NSEC" > $(CONFIG); \
	printf 'EXPO_PUBLIC_API_BASE_URL=$(DEV_API_URL)\nEXPO_PUBLIC_DEV_NSEC=%s\n' "$$NSEC" > mobile/.env; \
	echo "Created $(CONFIG) and mobile/.env with a disposable development identity."

dev-init:
	go run ./cmd/mobicode init --config $(CONFIG)

dev-run: dev-init
	go run ./cmd/mobicode serve --config $(CONFIG)

dev-sign:
	go run ./cmd/mobicode dev sign-request --config $(CONFIG) --url $(DEV_API_URL)/graphql --body '$(DEV_GRAPHQL_BODY)'

dev-query:
	@AUTH="$$(go run ./cmd/mobicode dev sign-request --config $(CONFIG) --url $(DEV_API_URL)/graphql --body '$(DEV_GRAPHQL_BODY)')"; \
	curl --fail-with-body $(DEV_API_URL)/graphql -H "$$AUTH" -H 'Content-Type: application/json' --data '$(DEV_GRAPHQL_BODY)'
