---
sidebar_position: 1
---

# Getting started

MobiCode is a mobile first coding agent for on the go development. This repository currently provides the foundation for its Go server, Expo mobile app, and documentation website.

## Repository layout

| Path | Role |
| --- | --- |
| `cmd/server` | Runnable Go server entrypoint |
| `internal/http` | Chi router, middleware, and GraphQL transport |
| `internal/graphql` | gqlgen schema, generated code, and resolvers |
| `internal/web` | Templ pages, HTMX handlers, and shadcn-templ components |
| `public` | Compiled CSS and scripts embedded by the server |
| `mobile` | Expo and React Native application using gluestack UI |
| `website` | This Docusaurus site |

## Run the server

From the repository root:

```bash
cp .env.example .env
pnpm install --frozen-lockfile
make server/dev
```

The server listens on port 8080 by default. Set `MOBICODE_SERVER_PORT` in `.env` or your shell to change it. Server variables use the `MOBICODE_SERVER_*` prefix; mobile variables use `MOBICODE_MOBILE_*`. `GET /healthz` returns `ok`.

Send GraphQL queries as JSON to `POST /mobile/graphql`. For example:

```json
{"query":"{ health { status } }"}
```

Set `MOBICODE_SERVER_PLAYGROUND=true` to enable the playground at `/mobile/graphql/playground` for localhost clients. To add a GraphQL field, edit a feature schema in `internal/graphql/schema`, run `go tool gqlgen generate`, and implement the generated resolver. Application services are assembled in `internal/app` and injected into the resolver.

## Run the mobile app

```bash
cd mobile
pnpm install
pnpm start
```

Follow the Expo CLI instructions to open the app on a device or simulator.

## Run this site

```bash
cd website
pnpm install
pnpm start
```

The site deploys to GitHub Pages when `master` changes.
