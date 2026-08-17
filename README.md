# Mobicode

Mobicode is a self-hosted, mobile-first coding-agent API. It is a single Go
binary with a GraphQL API, SQLite persistence, and email/password
authentication.

## Features

- GraphQL API at `/query`
- SQLite with embedded, versioned migrations
- Email/password authentication with short-lived JWT access tokens and rotated
  per-device refresh tokens
- `mobicode init` and `mobicode serve` operational workflow

## Requirements

- Go 1.26 or later
- A C compiler with CGO enabled (required by SQLite)

## Quick start

```sh
git clone https://github.com/soumajitgh/mobicode.git
cd mobicode
cp .env.example .env
```

Set a secure `JWT_SECRET` in `.env`. It must be at least 32 bytes:

```sh
openssl rand -base64 48
```

Initialize the database and start the API:

```sh
go run ./cmd/mobicode init --config .env
go run ./cmd/mobicode serve --config .env
```

The API is available at `http://localhost:8080/query`. Health checks are
available at `http://localhost:8080/health`.

## Configuration

Configuration comes from environment variables or an optional `.env` file
passed with `--config`.

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | HTTP server port |
| `ENV` | `development` | Runtime environment; use `production` for JSON logs |
| `DATABASE_PATH` | `data/app.db` | SQLite database path |
| `JWT_SECRET` | — | Required secret for signing access tokens; 32 bytes minimum |

`mobicode init` is idempotent and applies pending migrations. `mobicode serve`
does not run migrations; it exits with an actionable error until the schema is
current.

## GraphQL

All application operations use GraphQL. For example, register an account:

```sh
curl http://localhost:8080/query \
  -H 'Content-Type: application/json' \
  --data '{"query":"mutation { register(name: \\"Ada\\", email: \\"ada@example.com\\", password: \\"correct-horse-battery-staple\\") { accessToken refreshToken user { id name email } } }"}'
```

Store refresh tokens in platform-secure storage such as iOS Keychain or Android
Keystore. Do not store them in plain local storage.

## Development

```sh
go test ./...
go run github.com/99designs/gqlgen generate
go build ./cmd/mobicode
```

The Expo/React Native client, when present, lives independently in
[`mobile/`](mobile/).

## Contributing

Issues and pull requests are welcome. Keep changes focused, add tests for
behavior changes, and run `go test ./...` before opening a pull request.

For security vulnerabilities, use
[GitHub's private security advisory form](https://github.com/soumajitgh/mobicode/security/advisories/new).

## License

[MIT](LICENSE.md)
