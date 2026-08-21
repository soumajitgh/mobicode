# Contributing to MobiCode

Thanks for contributing. MobiCode is a Go API and Expo mobile client with one
mobile-owned Nostr identity per deployment.

## Quick start

Requirements: Go 1.26+, Node.js 22+, pnpm 11+, and a C compiler for SQLite.

```sh
git clone https://github.com/soumajitgh/mobicode.git
cd mobicode
make dev-setup
make dev-run
```

In another terminal, start the mobile app:

```sh
make mobile-dev
```

`make dev-setup` creates ignored `.env` and `mobile/.env` files with the same
disposable Nostr development key. It refuses to overwrite existing local
configuration. Never use a personal or production key for `DEV_NSEC` or
`EXPO_PUBLIC_DEV_NSEC`.

Useful local commands:

```sh
make dev-query      # Call viewer through the real NIP-98 auth path
make test           # Go test suite
make mobile-check   # Expo TypeScript check
make build          # Build the server binary
```

## Development guidelines

- Keep HTTP handlers thin; business and persistence logic stay outside the
  `templ`/HTMX rendering layer.
- Preserve the single-owner security model. Never add development bypasses that
  can run outside `ENV=development`.
- Keep private keys out of logs, commits, and application UI state.
- Regenerate committed code after schema or UI component changes:

  ```sh
  go run github.com/99designs/gqlgen generate
  go run github.com/a-h/templ/cmd/templ generate
  ```

- Add focused tests for behavior changes and run the relevant checks before
  opening a pull request.

## Pull requests

- Start from `main` on a focused branch and keep each PR scoped to one change.
- Explain user-visible behavior, security implications, and how you tested it.
- Do not commit `.env`, `mobile/.env`, databases, generated build output, or
  real credentials.
- Use GitHub’s private security advisory process for vulnerabilities rather
  than opening a public issue.
