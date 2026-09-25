# MobiCode

Mobile first coding agent for on the go development.

This repository contains the initial framework for three components:

| Directory | Purpose |
| --- | --- |
| `cmd/server` | Go server entrypoint |
| `internal/http` | Chi router, middleware, and handlers |
| `mobile` | Expo React Native app with gluestack UI |
| `website` | Docusaurus documentation site |

## Run locally

Run `make help` to see the daily development commands. The most common are:

- Server: copy `.env.example` to `.env`, then run `make server/dev` (listens on `:8080` by default; set `MOBICODE_SERVER_PORT` to change the port; `GET /healthz` returns `ok`)
- Mobile: `make mobile/install`, then `make mobile/start` (or `make mobile/android`, `make mobile/ios`, `make mobile/web`)
- Website: `make website/install`, then `make website/start`

The documentation site is published from `master` through GitHub Actions at [soumajitgh.github.io/mobicode](https://soumajitgh.github.io/mobicode/).

Licensed under the Apache License 2.0. See [LICENSE](LICENSE).
