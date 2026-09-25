.PHONY: help init server/dev server/build server/fmt server/check gql templ shadcn generate web/install web/build web/watch web/serve mobile/install mobile/start mobile/android mobile/ios mobile/web mobile/lint mobile/typecheck website/install website/start website/build

GO ?= go
PNPM ?= pnpm
WEB_PORT ?= 8080

help:
	@printf '%s\n' \
	  'Setup:' \
	  '  make init               Set up a fresh clone (Go, web, mobile, website)' \
	  'Server:' \
	  '  make server/dev         Run the Go API (MOBICODE_SERVER_PORT=8080 by default)' \
	  '  make server/build       Build bin/mobicode-server' \
	  '  make server/fmt         Format Go source' \
	  '  make server/check       Run go vet' \
	  '  make gql                Generate GraphQL code' \
	  '  make templ              Generate Go code from Templ pages' \
	  '  make shadcn             Bundle shadcn-templ component scripts' \
	  '  make generate           Run all code generators' \
	  'Web app:' \
	  '  make web/install        Install Tailwind and HTMX build tools' \
	  '  make web/build          Generate Templ and compile web assets' \
	  '  make web/watch          Watch Templ/Go and Tailwind (proxy on :7331)' \
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

server/dev: web/build
	$(GO) run ./cmd/server

server/build: web/build
	@mkdir -p bin
	$(GO) build -o bin/mobicode-server ./cmd/server

server/fmt:
	$(GO)fmt -w cmd internal

server/check:
	$(GO) vet ./...

gql:
	$(GO) run github.com/99designs/gqlgen generate

templ:
	$(GO) tool templ generate -path internal/web

shadcn:
	$(GO) tool shadcn-templ bundle

generate: gql templ shadcn

web/install:
	$(PNPM) install --frozen-lockfile

web/build: templ shadcn
	$(PNPM) run build

web/serve:
	MOBICODE_SERVER_DEV_ASSETS=true MOBICODE_SERVER_PORT=$(WEB_PORT) $(GO) run ./cmd/server

web/watch: web/build
	@$(PNPM) run css:watch & css_pid=$$!; \
	$(GO) tool shadcn-templ bundle --watch & scripts_pid=$$!; \
	trap 'kill $$css_pid $$scripts_pid 2>/dev/null || true' EXIT INT TERM; \
	$(GO) tool templ generate -path internal/web -watch -cmd="$(MAKE) -C ../.. web/serve WEB_PORT=$(WEB_PORT)" -proxy="http://localhost:$(WEB_PORT)" -open-browser=false

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
