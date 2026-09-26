.PHONY: help init server/dev server/seed server/build server/test server/test-race server/test-e2e server/fmt server/check mobile/install mobile/start mobile/android mobile/ios mobile/web mobile/lint mobile/typecheck website/install website/start website/build

GO ?= go
GOLANGCI_LINT ?= golangci-lint
PNPM ?= pnpm
SERVER_PORT ?= 8080
RELOAD_PORT ?= 7331

help:
	@printf '%s\n' \
	  'Setup:' \
	  '  make init               Set up a fresh clone (Go, web, mobile, website)' \
	  'Server:' \
	  '  make server/dev         Run Air with Go, Templ, and CSS reload' \
	  '  make server/seed        Create the initial development account' \
	  '  make server/build       Build bin/mobicode-server' \
	  '  make server/test        Run package and integration tests' \
	  '  make server/test-race   Run package and integration tests with the race detector' \
	  '  make server/test-e2e    Run Playwright browser tests' \
	  '  make server/fmt         Format Go source with gofumpt and goimports' \
	  '  make server/check       Run go vet and Go linters' \
	  'Mobile:' \
	  '  make mobile/install     Install locked dependencies' \
	  '  make mobile/start       Start Expo' \
	  '  make mobile/android     Open Expo on Android' \
	  '  make mobile/ios         Open Expo on iOS' \
	  '  make mobile/web         Open Expo in a browser' \
	  '  make mobile/lint        Run Expo lint' \
	  '  make mobile/typecheck   Check TypeScript' \
	  'Website:' \
	  '  make website/install    Install locked dependencies' \
	  '  make website/start      Start Docusaurus' \
	  '  make website/build      Build the static site'

init:
	@GO="$(GO)" PNPM="$(PNPM)" sh scripts/init.sh

server/dev:
	$(PNPM) run htmx:copy
	@$(PNPM) run css:watch & css_pid=$$!; \
	trap 'kill $$css_pid 2>/dev/null || true' EXIT INT TERM; \
	MOBICODE_SERVER_DEV_ASSETS=true MOBICODE_SERVER_PORT=$(SERVER_PORT) $(GO) tool air -proxy.app_port=$(SERVER_PORT) -proxy.proxy_port=$(RELOAD_PORT) -c .air.toml

server/seed:
	$(GO) run ./cmd/seed

server/build:
	$(GO) tool templ generate -path internal/web
	$(GO) tool shadcn-templ bundle
	$(PNPM) run build
	@mkdir -p bin
	$(GO) build -o bin/mobicode-server ./cmd/server

server/test:
	$(GO) test ./cmd/... ./internal/... ./tests/integration/...

server/test-race:
	$(GO) test -race ./cmd/... ./internal/... ./tests/integration/...

server/test-e2e:
	$(PNPM) run test:e2e

server/fmt:
	$(GOLANGCI_LINT) fmt

server/check:
	$(GO) vet ./...
	$(GOLANGCI_LINT) run

mobile/install:
	cd mobile && $(PNPM) install --frozen-lockfile

mobile/start:
	cd mobile && $(PNPM) start

mobile/android:
	cd mobile && $(PNPM) run android

mobile/ios:
	cd mobile && $(PNPM) run ios

mobile/web:
	cd mobile && $(PNPM) run web

mobile/lint:
	cd mobile && $(PNPM) exec expo lint

mobile/typecheck:
	cd mobile && $(PNPM) exec tsc --noEmit

website/install:
	cd website && $(PNPM) install --frozen-lockfile

website/start:
	cd website && $(PNPM) start

website/build:
	cd website && $(PNPM) run build
