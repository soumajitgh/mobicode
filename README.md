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
pnpm mobile
```

In a second terminal, start the API:

```sh
pnpm server
```

The API health check is available at `http://localhost:8080/health`.

## Commands

```sh
pnpm mobile              # Start Expo
pnpm server              # Start the Go API
pnpm nx run server:test  # Test the Go API
pnpm nx run server:build # Build the Go API
```

## License

[MIT](LICENSE)
