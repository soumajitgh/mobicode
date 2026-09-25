---
sidebar_position: 1
---

# Getting started

MobiCode is a mobile first coding agent for on the go development. This repository currently provides the foundation for its Go server, Expo mobile app, and documentation website.

## Repository layout

| Path | Role |
| --- | --- |
| `cmd/server` | Runnable Go server entrypoint |
| `internal/server` | Internal HTTP handlers |
| `mobile` | Expo and React Native application using gluestack UI |
| `website` | This Docusaurus site |

## Run the server

From the repository root:

```bash
go run ./cmd/server
```

The server listens on port 8080 by default. Set `ADDR` to change its listen address. `GET /healthz` returns `ok`.

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
