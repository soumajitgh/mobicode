# Website

This website is built using [Docusaurus](https://docusaurus.io/), a modern static website generator.

## Installation

```sh
pnpm install
```

## Local Development

```sh
pnpm start
```

This command starts a local development server and opens up a browser window. Most changes are reflected live without having to restart the server.

## Build

```sh
pnpm build
```

This command generates static content into the `build` directory and can be served using any static contents hosting service.

## Deployment

GitHub Actions builds and deploys the site to GitHub Pages when changes under
`docs/` are pushed to `main`. The published site is available at
`https://soumajitgh.github.io/mobicode/`.
