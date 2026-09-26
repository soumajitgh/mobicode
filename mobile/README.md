# Mobicode mobile

The Expo app uses `app/` for routes, `src/features/` for product features, `src/shared/` for reusable UI and configuration, `src/api/` for GraphQL operations and transport, and `src/store/` for Zustand stores. `(auth)` contains pairing; `(app)` contains the authenticated shell. The root layout restores the session and protects both route groups.

The server's `createDevicePairing` mutation produces a `mobicode://pair?server=…&token=…` QR payload. Paste that link into the pairing screen to claim a mobile session. The app validates the payload, calls `claimDevicePairing` at the supplied server, and stores the server URL and access token in Expo SecureStore. On startup it checks `viewer` before entering `(app)`. The GraphQL endpoint is `{serverBaseURL}/mobile/graphql`; it is runtime session state, not a build environment variable. Older token-only sessions must pair again because they did not store their server address. Pairing on web is unavailable because SecureStore is native only.

GraphQL operations live with their feature or under `src/api/graphql/`. `codegen.ts` reads the Go server schema and generates typed documents in `src/api/graphql/generated/`. Run `pnpm run graphql:generate` after schema or operation changes. Zustand owns app state; session credentials are persisted in Expo SecureStore.

Gluestack UI v5 is initialized with NativeWind v5. The CLI-generated provider and all 58 available components live in `src/shared/components/ui/`; use `pnpm dlx gluestack-ui@latest add <component> --use-pnpm` for updates. The generated v5 alpha templates currently contain strict TypeScript and lint errors, so app checks exclude that generated directory. Imported components are still checked and bundled with the app.

Run `pnpm run typecheck`, `pnpm run lint`, and `pnpm run format:check` before submitting changes. After adding a native dependency, rebuild the development app.
