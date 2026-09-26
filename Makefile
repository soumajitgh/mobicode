.PHONY: help init server/install server/gql server/templ server/shadcn server/generate server/assets server/dev server/build server/serve server/watch server/fmt server/lint server/check mobile/install mobile/start mobile/android mobile/ios mobile/web mobile/lint mobile/typecheck website/install website/start website/build

GO ?= go
GOLANGCI_LINT ?= golangci-lint
PNPM ?= pnpm
WEB_PORT ?= 8080

help:
	@printf '%s\n' \
	  'Setup:' \
	  '  make init               Set up a fresh clone (Go, web, mobile, website)' \
	  'Server:' \
	  '  make server/install     Install browser asset dependencies' \
	  '  make server/generate    Generate GraphQL, Templ, and component code' \
	  '  make server/assets      Build browser assets' \
	  '  make server/dev         Build assets and run the Go server' \
	  '  make server/build       Build bin/mobicode-server' \
	  '  make server/watch       Watch Templ and CSS (proxy on :7331)' \
	  '  make server/fmt         Format Go source with gofumpt and goimports' \
	  '  make server/lint        Run Go linters' \
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

server/install:
	$(PNPM) install --frozen-lockfile

server/gql:
	$(GO) run github.com/99designs/gqlgen generate

server/templ:
	$(GO) tool templ generate -path internal/web

server/shadcn:
	$(GO) tool shadcn-templ bundle

server/generate: server/gql server/templ server/shadcn

server/assets: server/templ server/shadcn
	$(PNPM) run build

server/dev: server/assets
	$(GO) run ./cmd/server

server/build: server/assets
	@mkdir -p bin
	$(GO) build -o bin/mobicode-server ./cmd/server

server/serve:
	MOBICODE_SERVER_DEV_ASSETS=true MOBICODE_SERVER_PORT=$(WEB_PORT) $(GO) run ./cmd/server

server/watch: server/assets
	@$(PNPM) run css:watch & css_pid=$$!; \
	$(GO) tool shadcn-templ bundle --watch & scripts_pid=$$!; \
	trap 'kill $$css_pid $$scripts_pid 2>/dev/null || true' EXIT INT TERM; \
	$(GO) tool templ generate -path internal/web -watch -cmd="$(MAKE) -C ../.. server/serve WEB_PORT=$(WEB_PORT)" -proxy="http://localhost:$(WEB_PORT)" -open-browser=false

server/fmt:
	$(GOLANGCI_LINT) fmt

server/lint:
	$(GOLANGCI_LINT) run

server/check:
	$(GO) vet ./...
	$(MAKE) server/lint

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
