# MobiCode

Mobile first coding agent for on the go development.

This repository contains the Go server, browser app, mobile app, and documentation site:

| Directory | Purpose |
| --- | --- |
| `cmd/server` | Go server entrypoint |
| `internal/http` | Chi router, middleware, and GraphQL transport |
| `internal/store` | SQLite connection, Goose migrations, and GORM repositories |
| `internal/graphql` | gqlgen schema, generated code, and resolvers |
| `internal/web` | Templ pages, HTMX handlers, and shadcn-templ components |
| `public` | Compiled Tailwind CSS and browser scripts embedded in the server |
| `mobile` | Expo React Native app with gluestack UI |
| `website` | Docusaurus documentation site |

## Run locally

Use Go 1.26.5, Node.js 22 or newer, and pnpm 11.27.1. From a fresh clone, run `make init` to create `.env` (if missing), download Go dependencies, install the locked dependencies for the web app, mobile app, and website, install the Playwright Chromium browser, and build browser assets. You can run it again without replacing an existing `.env`.

`MOBICODE_SERVER_ENV` defaults to `development` for colored console logs. Set it to `production` for structured JSON logs. HTTP completion logs include method, path, status, duration, and request ID in both environments.

Run `make help` to see the daily development commands. The most common are:

- Server: run `make server/dev` after initialization. Air rebuilds on Go and Templ changes, Tailwind watches CSS, and the browser reloads through `http://localhost:7331/`. The API also listens directly on `:8080` (`GET /healthz` returns `ok`). Set `SERVER_PORT` and `RELOAD_PORT` on the make command to use other local ports.
- Database: set `MOBICODE_SERVER_DATA_DIR` in `.env` to choose the data directory (default `tmp` in local dev, `~/.mobicode` otherwise; the database file is placed under `database/mobicode.db`). GORM logging defaults to `warn`. Startup runs embedded Goose migrations before serving requests.
- GraphQL: send POST requests to `/mobile/graphql`; the playground is disabled by default
- Mobile pairing: an authenticated browser session calls `createDevicePairing`, renders the returned `qrPayload`, and the mobile app calls `claimDevicePairing` with its token and device metadata. The mobile app then sends `Authorization: Bearer <accessToken>` for `viewer`. Set `MOBICODE_SERVER_BASE_URL` to an externally reachable HTTP(S) origin for reverse proxies or public deployments. Without it, the server advertises a discovered private IPv4 address and `MOBICODE_SERVER_PORT`; pairing creation returns `SERVER_ADDRESS_UNAVAILABLE` if none is available.
- Schema changes: edit `internal/graphql/schema/*.graphqls`, run `go tool gqlgen generate`, then implement the generated resolver using services from `internal/app`
- Browser app: run `make server/dev` after initialization; open `http://localhost:8080/`
- Browser development: use `make server/dev` and open `http://localhost:7331/` for automatic reloads
- Mobile: `make mobile/start` (or `make mobile/android`, `make mobile/ios`, `make mobile/web`)
- Development account: `make server/seed` creates the first user in the configured local database and prints a random password once. Re-running it leaves existing accounts alone. Run it against the same database as the server before using mobile auto-pair.
- Air uses `tmp/air` for disposable build files so its exit cleanup does not remove the development database under `tmp/database`.
- Mobile architecture and generated GraphQL types: see [mobile/README.md](mobile/README.md). The app uses Expo Router, urql, and a runtime paired server URL.
- Website: `make website/start`

## Go formatting and linting

Install [golangci-lint v2](https://golangci-lint.run/docs/welcome/install/local/) locally (built with Go 1.26.5 or newer), then run `make server/fmt` to format hand-written Go files with gofumpt and goimports. Run `make server/check` for `go vet` plus govet, staticcheck, errcheck, ineffassign, and unused. The formatter skips generated Go files. Run `golangci-lint run --fix` when you want available lint fixes applied automatically.

`make init` installs the Lefthook `commit-msg` hook. Commit subjects use Conventional Commits, for example `feat(server): add hot reload`. To reinstall the hook, run `pnpm exec lefthook install`.

## Run with Docker Compose

`compose.yaml` is a development setup. It mounts the source tree, uses Air to rebuild the server when Go or Templ files change, and watches Tailwind CSS. Run `docker compose up --build` to start it. The reload proxy is available at `http://localhost:7331/` (or `MOBICODE_RELOAD_PORT`), and the server is available directly at `http://localhost:8080/` (or `MOBICODE_SERVER_PORT`).

SQLite needs no separate service. Set `MOBICODE_SERVER_DATA_DIR` in `.env` (default `tmp`); the `tmp/database` directory is backed by a named Compose volume so data persists across restarts. For local `make server/dev`, the same path is created in the repository and ignored by Git. Browser registration and password reset require the administrator token from `.env`; `make init` generates one when creating the file.

The documentation site is published from `master` through GitHub Actions at [soumajitgh.github.io/mobicode](https://soumajitgh.github.io/mobicode/).

Licensed under the Apache License 2.0. See [LICENSE](LICENSE).
