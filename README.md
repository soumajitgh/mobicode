# MobiCode

**A mobile-first coding agent for on-the-go development.**

MobiCode pairs an Expo mobile app with a Go API in an Nx monorepo.

## Stack

- Expo + React Native
- Go + Gin
- Nx + pnpm
- Mise-pinned Go toolchain

## Get started

```sh
corepack enable
pnpm setup
pnpm dev
```

Nx starts the mobile app, libSQL, and Go API together. The server applies its
migrations before serving requests; it waits up to 30 seconds for a fresh local
database. The API health check is available at
`http://localhost:8080/health`.

Copy `.env.example` to `.env` to override the default `http://localhost:9081`
database URL or other settings.

`LIBSQL_PORT` controls Docker's host port. If you change it, update
`SERVER_DATABASE_URL` to the same port.

## Commands

```sh
pnpm dev                 # Start the mobile app, libSQL, and Go API
pnpm mobile:dev          # Start Expo only
pnpm server:dev          # Start libSQL and the Go API with live reload
pnpm server:run          # Start libSQL and run the API once
pnpm deps:up             # Start only local dependencies
pnpm deps:down           # Stop local dependencies (data is retained)
pnpm deps:logs           # Follow dependency logs
pnpm migration:create -- --name add_users # Create a new migration
pnpm nx run server:test  # Test the Go API
pnpm nx run server:build # Build the Go API
```

## License

[MIT](LICENSE)
