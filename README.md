# MobiCode

Mobile first coding agent for on the go development.

This repository contains the Go server, browser app, mobile app, and documentation site:

| Directory | Purpose |
| --- | --- |
| `cmd/server` | Go server entrypoint |
| `internal/http` | Chi router, middleware, and GraphQL transport |
| `internal/graphql` | gqlgen schema, generated code, and resolvers |
| `internal/web` | Templ pages, HTMX handlers, and shadcn-templ components |
| `public` | Compiled Tailwind CSS and browser scripts embedded in the server |
| `mobile` | Expo React Native app with gluestack UI |
| `website` | Docusaurus documentation site |

## Run locally

Run `make help` to see the daily development commands. The most common are:

- Server: copy `.env.example` to `.env`, then run `make server/dev` (listens on `:8080` by default; set `MOBICODE_SERVER_PORT` to change the port; `GET /healthz` returns `ok`)
- GraphQL: send POST requests to `/mobile/graphql`; set `MOBICODE_SERVER_PLAYGROUND=true` to enable `/mobile/graphql/playground` for localhost clients
- Schema changes: edit `internal/graphql/schema/*.graphqls`, run `make gql`, then implement the generated resolver using services from `internal/app`
- Browser app: run `make web/install`, then `make web/build` and `make server/dev`; open `http://localhost:8080/`
- Browser development: run `make web/watch` and open `http://localhost:7331/` for reloads (`WEB_PORT=8090` changes the app port)
- Mobile: `make mobile/install`, then `make mobile/start` (or `make mobile/android`, `make mobile/ios`, `make mobile/web`)
- Website: `make website/install`, then `make website/start`

The documentation site is published from `master` through GitHub Actions at [soumajitgh.github.io/mobicode](https://soumajitgh.github.io/mobicode/).

Licensed under the Apache License 2.0. See [LICENSE](LICENSE).
