---
sidebar_position: 2
---

# Browser app

The browser app uses Go, Templ, HTMX, Tailwind CSS, and shadcn-templ. It runs on the same Chi server as `/mobile/graphql` and is separate from the Docusaurus site in `website/`.

## Install and run

From the repository root:

```bash
pnpm install --frozen-lockfile
make server/build
make server/dev
```

Open `http://localhost:7331/` for automatic browser reloads, or `http://localhost:8080/` for the direct server. The home page's **Check server** button requests `/partials/status` through HTMX. The endpoint returns a Templ fragment using the same health service as GraphQL.

For development with Go, Templ, and Tailwind reload:

```bash
make server/dev
```

Air provides the reload proxy at `http://localhost:7331/`. The CSS watcher serves files from disk during development; normal runs embed the compiled assets in the Go binary.

## Add pages and components

Place page templates in `internal/web/pages` and handlers in `internal/web/handlers`. Register browser routes in `internal/web/routes.go`. Application services are created in `internal/app` and passed to handlers; templates only render their inputs.

shadcn-templ is pinned as a Go tool and configured in `components.json`. To add another component:

```bash
go tool shadcn-templ add card
make server/build
```

The CLI places component source in `internal/web/components` and shared helpers in `internal/web/utils`. `make server/build` runs Templ generation, the shadcn script bundle, Tailwind compilation, the local HTMX copy, and the Go build. Commit the generated Go and `public` assets with the source changes so a checkout can run the server directly.
