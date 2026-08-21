# Mobicode Repository Folder Structure

Directory organization and folder responsibilities for the **Mobicode** repository.

---

## Folder Structure Overview

```
mobicode/
├── agent_docs/          # Architecture and agent developer documentation
│   ├── architecture.md
│   └── testing.md
├── build/               # Compiled binary output directory
├── cmd/                 # Application CLI entrypoints and commands
│   └── mobicode/        # Primary executable package & Cobra CLI definitions
├── data/                # Default runtime data directory for SQLite databases
├── db/                  # Database migration files and SQL scripts
│   └── migrations/      # Embedded versioned SQL migration scripts
├── docs/                # Project documentation site (Docusaurus)
├── graphql/             # GraphQL schemas and gqlgen code generation output
│   ├── generated/       # Generated GraphQL execution engine code
│   └── model/           # Generated Go structs for GraphQL schema types
├── internal/            # Core private Go packages
│   ├── auth/            # NIP-98 verification, owner identity, replay protection
│   ├── config/          # Configuration loader (.env and environment variables)
│   ├── database/        # GORM SQLite setup and migration verifier
│   ├── graphql/         # GraphQL HTTP handlers & custom error presenter
│   ├── health/          # Health check endpoint handlers (/health)
│   ├── logger/          # Zap structured logger and Fx logger integration
│   ├── server/          # Chi HTTP router and server lifecycle hooks
│   └── setup/           # templ + HTMX one-time mobile pairing flow
├── mobile/              # Expo / React Native mobile client application
│   ├── assets/          # Mobile app static assets (icons, fonts, images)
│   └── src/             # React Native UI component source code
└── tmp/                 # Temporary directory for Air hot-reload build outputs
```

---

## Detailed Folder Responsibilities

### [`agent_docs/`](file:///Users/soumajit/Developer/projects/mobicode/agent_docs)
Contains architectural documentation, guidelines, and repository context designed for AI agents and developers working on the codebase.

### [`build/`](file:///Users/soumajit/Developer/projects/mobicode/build)
Default output directory for compiled Go binaries (e.g. `build/mobicode`), populated when running `make build`.

### [`cmd/`](file:///Users/soumajit/Developer/projects/mobicode/cmd)
Contains command-line entry points for the Go application.
- [`cmd/mobicode/`](file:///Users/soumajit/Developer/projects/mobicode/cmd/mobicode): Main application binary package ([main.go](file:///Users/soumajit/Developer/projects/mobicode/cmd/mobicode/main.go)).
- [`cmd/mobicode/cmd/`](file:///Users/soumajit/Developer/projects/mobicode/cmd/mobicode/cmd): Cobra CLI command handlers:
  - [`init.go`](file:///Users/soumajit/Developer/projects/mobicode/cmd/mobicode/cmd/init.go): Initializes database schema and runs pending SQL migrations.
  - [`serve.go`](file:///Users/soumajit/Developer/projects/mobicode/cmd/mobicode/cmd/serve.go): Starts the API server using Fx dependency injection.
  - [`version.go`](file:///Users/soumajit/Developer/projects/mobicode/cmd/mobicode/cmd/version.go): Outputs binary version information.
  - `identity.go`: Explicit reset command for a lost or compromised owner key.

### [`data/`](file:///Users/soumajit/Developer/projects/mobicode/data)
Runtime directory used to store local SQLite database files (e.g., `data/app.db`).

### [`db/`](file:///Users/soumajit/Developer/projects/mobicode/db)
Database assets and migration files.
- [`db/migrations/`](file:///Users/soumajit/Developer/projects/mobicode/db/migrations): Contains numbered `.up.sql` and `.down.sql` migration scripts embedded into the application binary via Go `embed.FS` ([migrations.go](file:///Users/soumajit/Developer/projects/mobicode/db/migrations/migrations.go)).

### [`docs/`](file:///Users/soumajit/Developer/projects/mobicode/docs)
User documentation website source built with Docusaurus.

### [`graphql/`](file:///Users/soumajit/Developer/projects/mobicode/graphql)
GraphQL definitions and auto-generated code produced by `gqlgen`.
- [`graphql/generated/`](file:///Users/soumajit/Developer/projects/mobicode/graphql/generated): Contains generated GraphQL execution schema code ([generated.go](file:///Users/soumajit/Developer/projects/mobicode/graphql/generated/generated.go)).
- [`graphql/model/`](file:///Users/soumajit/Developer/projects/mobicode/graphql/model): Contains generated Go models corresponding to GraphQL schema types ([models_gen.go](file:///Users/soumajit/Developer/projects/mobicode/graphql/model/models_gen.go)).

### [`internal/`](file:///Users/soumajit/Developer/projects/mobicode/internal)
Private application code enforceably isolated by Go's `internal` package visibility rules.

- [`internal/auth/`](file:///Users/soumajit/Developer/projects/mobicode/internal/auth): Single-owner NIP-98 verification, BIP-340 signature checks, and replay protection.
- [`internal/config/`](file:///Users/soumajit/Developer/projects/mobicode/internal/config): Parses runtime configuration from environment variables and `.env` files ([config.go](file:///Users/soumajit/Developer/projects/mobicode/internal/config/config.go)).
- [`internal/database/`](file:///Users/soumajit/Developer/projects/mobicode/internal/database): Manages SQLite connections with GORM ([connection.go](file:///Users/soumajit/Developer/projects/mobicode/internal/database/connection.go)) and runs/verifies embedded SQL migrations ([migrator.go](file:///Users/soumajit/Developer/projects/mobicode/internal/database/migrator.go)).
- [`internal/graphql/`](file:///Users/soumajit/Developer/projects/mobicode/internal/graphql): Assembles GraphQL server handlers ([handler.go](file:///Users/soumajit/Developer/projects/mobicode/internal/graphql/handler.go)), custom error presenter, and binds GraphQL resolvers.
- [`internal/health/`](file:///Users/soumajit/Developer/projects/mobicode/internal/health): Implements REST health check handler (`/health`).
- [`internal/logger/`](file:///Users/soumajit/Developer/projects/mobicode/internal/logger): Configures Uber Zap logger and Fx framework event logger.
- [`internal/server/`](file:///Users/soumajit/Developer/projects/mobicode/internal/server): Configures the Chi HTTP router and HTTP server lifecycle hooks.
- [`internal/setup/`](file:///Users/soumajit/Developer/projects/mobicode/internal/setup): Typed `templ` components, HTMX fragments, browser-session setup state, and mobile QR pairing routes.

### [`mobile/`](file:///Users/soumajit/Developer/projects/mobicode/mobile)
Expo React Native mobile client application.
- [`mobile/src/app/`](file:///Users/soumajit/Developer/projects/mobicode/mobile/src/app): Expo Router routes, including the `mobicode://pair` deep-link route.
- [`mobile/src/features/`](file:///Users/soumajit/Developer/projects/mobicode/mobile/src/features): Secure key storage and signed API client implementation.
- [`mobile/assets/`](file:///Users/soumajit/Developer/projects/mobicode/mobile/assets): Static assets such as images, fonts, and icons.

### [`tmp/`](file:///Users/soumajit/Developer/projects/mobicode/tmp)
Temporary directory generated and used by `air` for hot-reloading build outputs during local development (`tmp/server`).
