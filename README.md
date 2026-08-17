# MobiCode

**A mobile-first coding agent for on-the-go development.**

MobiCode is a Go-first application. The Go API owns the repository root, while
the Expo/React Native client is an independent package in [`mobile/`](mobile/).
The layout follows the same broad ownership model as
[Apache Answer](https://github.com/apache/answer): one root Go module with the
client kept in a dedicated subdirectory.

## Repository structure

```text
.
├── cmd/api/                 # Go executable
├── internal/                # Application code
├── docs/                    # Architecture documentation
├── mobile/                  # Standalone Expo/React Native app
├── go.mod                   # Root Go module
├── Makefile                 # Go-first development commands
├── Dockerfile               # API image
└── compose.yml              # Local libSQL dependency
```

There is no JavaScript workspace or task orchestrator at the root. JavaScript
dependencies and commands belong exclusively to `mobile/`.

## Run the API

Install the toolchain pinned by Mise, copy the optional environment file, and
start the server:

```sh
mise install
cp .env.example .env
make run
```

`make run` starts libSQL with Docker Compose, applies migrations, and serves the
API. GraphQL is available at `http://localhost:8080/graphql`; development mode
also exposes the GraphQL Playground at `http://localhost:8080/playground`.
Health endpoints are `/health/live` and `/health/ready`.

For live reload, install [Air](https://github.com/air-verse/air) and use
`make dev`.

## Run the mobile app

The mobile app manages its own dependencies and lockfile:

```sh
corepack enable
pnpm --dir mobile install --frozen-lockfile
pnpm --dir mobile start
```

## Commands

```sh
make dev                         # Start libSQL and the API with live reload
make run                         # Start libSQL and run the API once
make build                       # Build build/mobicode
make test                        # Run Go tests
make vet                         # Run go vet
make lint                        # Run golangci-lint
make check                       # Format, vet, lint, and test the Go code
make deps-up                     # Start local dependencies
make deps-down                   # Stop local dependencies; retain their data
make deps-logs                   # Follow libSQL logs
make migrate-create name=users  # Create a SQL migration
make mobile-install              # Install locked mobile dependencies
make mobile-dev                  # Start Expo
make mobile-check                # Type-check the mobile app
```

## License

[MIT](LICENSE)
