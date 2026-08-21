# MobiCode

**A small, self-hosted home for your coding-agent workflow.**

MobiCode is a mobile-first app foundation you run yourself: one Go service, a
local SQLite database, and one phone that owns the deployment.

## Get started

You need Go 1.26+, Node.js, pnpm, and a C compiler for SQLite.

```sh
git clone https://github.com/soumajitgh/mobicode.git
cd mobicode
make dev-setup
make dev-run
```

`make dev-setup` creates local, ignored configuration files and a disposable
development identity. It will not replace files that already exist.

In another terminal, start the mobile app:

```sh
make mobile-dev
```

Your local API is now running at `http://localhost:8085`.

## What is included

- A Go server with a GraphQL API
- A local SQLite database
- One mobile-owned identity for each deployment
- An Expo mobile app in [`mobile/`](mobile/)
- A browser setup page for pairing a phone outside development

## Pair a phone

For a normal deployment, set `PUBLIC_BASE_URL` in `.env`, start the server,
then open `/setup` in your browser. Scan the QR code with the MobiCode app and
confirm the displayed fingerprint.

The first phone you pair becomes the owner. Set it up privately: do not expose
`/setup` to the public internet before you are ready to pair.

## For developers

```sh
make test          # Run Go tests
make mobile-check  # Check the mobile app
make build         # Build the server
make dev-query     # Test the authenticated API locally
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the contribution workflow.

## Security

Your phone keeps the private key; the server only receives signed requests.
Keep `.env` files private and never use a personal or production key for local
development. To report a vulnerability, use [GitHub's private security
advisory form](https://github.com/soumajitgh/mobicode/security/advisories/new).

## License

[MIT](LICENSE.md)
