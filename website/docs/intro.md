---
sidebar_position: 1
---

# Getting started

MobiCode is a mobile first coding agent for on the go development. This repository currently provides the foundation for its Go server, Expo mobile app, and documentation website.

## Repository layout

| Path | Role |
| --- | --- |
| `cmd/server` | Runnable Go server entrypoint |
| `internal/http` | Chi router, middleware, and handlers |
| `mobile` | Expo and React Native application using gluestack UI |
| `website` | This Docusaurus site |

## Run the server

From the repository root:

```bash
cp .env.example .env
go run ./cmd/server
```

The server listens on port 8080 by default. Set `MOBICODE_SERVER_PORT` in `.env` or your shell to change it. Server variables use the `MOBICODE_SERVER_*` prefix; mobile variables use `MOBICODE_MOBILE_*`. `GET /healthz` returns `ok`.

## Run the mobile app

```bash
cd mobile
npm install
npm start
```

Follow the Expo CLI instructions to open the app on a device or simulator.

## Run this site

```bash
cd website
npm install
npm start
```

The site deploys to GitHub Pages when `master` changes.
