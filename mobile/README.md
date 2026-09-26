# Mobicode mobile

The Expo app uses `app/` for routes, `src/features/` for product features, `src/components/` for reusable UI, `src/shared/` for theme and configuration, `src/api/` for GraphQL operations and transport, and `src/store/` for Zustand stores. `(auth)` contains pairing; `(tabs)` contains the authenticated Home, Projects, Monitor, and Settings pages. The root layout restores the session and protects both route groups.

The server's `createDevicePairing` mutation produces a `mobicode://pair?server=…&token=…` QR payload. Scan that code on the pairing screen to claim a mobile session. The app validates the payload, calls `claimDevicePairing` at the supplied server, and stores the server URL and access token in Expo SecureStore. On startup it checks `viewer` before entering `(tabs)`. The GraphQL endpoint is `{serverBaseURL}/mobile/graphql`; it is runtime session state, not a build environment variable. Older token-only sessions must pair again because they did not store their server address. The app targets iOS and Android.

For local development, set `MOBICODE_MOBILE_AUTO_PAIR=true` on the Go server to enable `/mobile/dev/auto-pair`. On boot, a development build tries this endpoint after checking a saved session and enters Home if it succeeds. When the endpoint is unavailable, normal QR pairing remains available. By default, the app uses the Expo dev server's hostname and port 8080 for the Go server. Set `EXPO_PUBLIC_MOBICODE_SERVER_URL` when the Go server uses a different address, or `EXPO_PUBLIC_MOBICODE_SERVER_PORT` when only its port differs. Auto pairing is disabled in production builds.

GraphQL operations live with their feature or under `src/api/graphql/`. `codegen.ts` reads the Go server schema and generates typed documents in `src/api/graphql/generated/`. Run `pnpm run graphql:generate` after schema or operation changes. Zustand owns app state; session credentials are persisted in Expo SecureStore.

Gluestack UI v5 is initialized with NativeWind v5. The CLI-generated provider and all 58 available components live in `src/components/ui/`; use `pnpm dlx gluestack-ui@latest add <component> --use-pnpm` for updates. The generated v5 alpha templates currently contain strict TypeScript and lint errors, so app checks exclude that generated directory. Imported components are still checked and bundled with the app.

Run `pnpm run typecheck`, `pnpm run lint`, and `pnpm run format:check` before submitting changes. After adding a native dependency, rebuild the development app.
