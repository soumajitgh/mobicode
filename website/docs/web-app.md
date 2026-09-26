---
sidebar_position: 2
---

# Browser app

The browser app uses Go, Templ, HTMX, Tailwind CSS, and shadcn-templ. It runs on the same Chi server as `/mobile/graphql` and is separate from the Docusaurus site in `website/`.

## Install and run

From the repository root:

```bash
make server/install
make server/assets
make server/dev
```

Open `http://localhost:8080/`. The home page's **Check server** button requests `/partials/status` through HTMX. The endpoint returns a Templ fragment using the same health service as GraphQL.

For development with Templ reload and Tailwind watching:

```bash
make server/watch
```

Open the reload proxy at `http://localhost:7331/`. Set `WEB_PORT=8090` on the make command to choose another app port. The CSS watcher serves files from disk during development; normal runs embed the compiled assets in the Go binary.

## Add pages and components

Place page templates in `internal/web/pages` and handlers in `internal/web/handlers`. Register browser routes in `internal/web/routes.go`. Application services are created in `internal/app` and passed to handlers; templates only render their inputs.

shadcn-templ is pinned as a Go tool and configured in `components.json`. To add another component:

```bash
go tool shadcn-templ add card
make server/assets
```

The CLI places component source in `internal/web/components` and shared helpers in `internal/web/utils`. `make server/assets` runs Templ generation, the shadcn script bundle, Tailwind compilation, and the local HTMX copy. Commit the generated Go and `public` assets with the source changes so a checkout can run the server directly.
