.PHONY: help server/dev server/build server/fmt server/check mobile/install mobile/start mobile/android mobile/ios mobile/web mobile/lint mobile/typecheck website/install website/start website/build

GO ?= go
NPM ?= npm

help:
	@printf '%s\n' \
	  'Server:' \
	  '  make server/dev         Run the Go API (MOBICODE_SERVER_PORT=8080 by default)' \
	  '  make server/build       Build bin/mobicode-server' \
	  '  make server/fmt         Format Go source' \
	  '  make server/check       Run go vet' \
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

server/dev:
	$(GO) run ./cmd/server

server/build:
	@mkdir -p bin
	$(GO) build -o bin/mobicode-server ./cmd/server

server/fmt:
	$(GO)fmt -w cmd internal

server/check:
	$(GO) vet ./...

mobile/install:
	$(NPM) --prefix mobile ci

mobile/start:
	$(NPM) --prefix mobile start

mobile/android:
	$(NPM) --prefix mobile run android

mobile/ios:
	$(NPM) --prefix mobile run ios

mobile/web:
	$(NPM) --prefix mobile run web

mobile/lint:
	cd mobile && npx expo lint

mobile/typecheck:
	cd mobile && npx tsc --noEmit

website/install:
	$(NPM) --prefix website ci

website/start:
	$(NPM) --prefix website start

website/build:
	$(NPM) --prefix website run build
