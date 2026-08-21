# Mobicode

Mobicode is a self-hosted, mobile-first coding-agent API. It is a single Go
binary with a GraphQL API, SQLite persistence, and a single mobile-owned
Nostr identity.

## Features

- GraphQL API at `/graphql`
- SQLite with embedded, versioned migrations
- NIP-98 signed HTTP authentication using a secp256k1/BIP-340 Nostr key
- One-time browser pairing at `/setup`, rendered with `templ` + HTMX + DaisyUI
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

Set the externally visible API origin in `.env`:

```sh
PUBLIC_BASE_URL=https://example.com
```

Initialize the database and start the API:

```sh
go run ./cmd/mobicode init --config .env
go run ./cmd/mobicode serve --config .env
```

Open `https://example.com/setup` once the server is running. The first browser
visitor receives a QR code; scan it in the MobiCode mobile app and approve the
displayed public-key fingerprint. That key becomes the deployment's only
authorized identity. Health checks are available at `/health`.

The initial setup flow intentionally follows an open-first-visitor model. Do
not expose `/setup` on a public server before the owner is ready to pair it.

## Configuration

Configuration comes from environment variables or an optional `.env` file
passed with `--config`.

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | HTTP server port |
| `ENV` | `development` | Runtime environment; use `production` for JSON logs |
| `DATABASE_PATH` | `data/app.db` | SQLite database path |
| `PUBLIC_BASE_URL` | `http://localhost:<PORT>` | Canonical public API origin used in NIP-98 request binding; must be HTTPS in production |

`mobicode init` is idempotent and applies pending migrations. `mobicode serve`
does not run migrations; it exits with an actionable error until the schema is
current.

## GraphQL

All GraphQL requests require `Authorization: Nostr <base64-signed-event>`. The
mobile signs the exact `POST` body, URL, and method; the server rejects invalid,
expired, unauthorized, or replayed proofs. `viewer` returns the configured
public key.

Set `EXPO_PUBLIC_API_BASE_URL` to the exact same origin as `PUBLIC_BASE_URL`
when running the mobile app. The private `nsec` is kept in Expo SecureStore and
unlocked with the device's configured biometric/device authentication.

To replace a lost or compromised mobile key, stop the server and run:

```sh
go run ./cmd/mobicode identity reset --confirm --config .env
```

## Development identity and API testing

For a fresh local checkout, one command creates ignored server/mobile `.env`
files with the same disposable identity:

```sh
make dev-setup
```

Then run the API with `make dev-run`, start Expo with `make mobile-dev`, and
call the authenticated viewer query with `make dev-query`. `make dev-setup`
refuses to overwrite existing local config files. To create a key manually,
use `go run ./cmd/mobicode dev generate-nsec`.

`DEV_NSEC` and `EXPO_PUBLIC_DEV_NSEC` must contain the same value. On a fresh development database,
`mobicode serve` automatically configures that key as the owner; the Expo app
uses it without a pairing QR or biometric prompt. Both settings are rejected
outside `ENV=development` and must never use a real Nostr identity.

To exercise the real NIP-98 server path from curl or a script, sign the exact
body first:

```sh
AUTH=$(go run ./cmd/mobicode dev sign-request --config .env \
  --body '{"query":"query Viewer { viewer { publicKey } }"}')
curl http://localhost:8080/graphql -H "$AUTH" -H 'Content-Type: application/json' \
  --data '{"query":"query Viewer { viewer { publicKey } }"}'
```

## Development

```sh
go test ./...
go run github.com/99designs/gqlgen generate
go run github.com/a-h/templ/cmd/templ generate
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
