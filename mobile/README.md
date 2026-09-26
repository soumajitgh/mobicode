# Mobicode mobile

The app uses the bundle and package identifier `com.soumajitgh.mobicode` on iOS and Android.

Copy `.env.example` to `.env` and set `EXPO_PUBLIC_MOBICODE_SERVER_URL` to the server address reachable by the phone or simulator. Expo embeds this public value in the app bundle; it is not a secret. On the server, enable `MOBICODE_MOBILE_AUTO_PAIR=true` only for trusted local development.

## API clients

- `src/lib/api.ts` provides an Axios client for REST endpoints such as development auto pairing.
- `src/lib/apollo.ts` provides the Apollo GraphQL client with an authorization link and normalized cache. GraphQL operations live in `src/graphql/*.graphql`.
- `codegen.ts` generates typed GraphQL documents from the Go server schema. Run `pnpm run graphql:generate` after changing GraphQL operations or the schema.
- TanStack Query owns non-GraphQL asynchronous state, including session bootstrap. Apollo owns GraphQL query state. Both providers are mounted in `App.tsx`.
- Zod validates responses that cross the REST boundary. The mobile access token is stored in Expo SecureStore.

Run `pnpm exec expo lint` and `pnpm exec tsc --noEmit` before submitting mobile changes. After adding a native module, rebuild the development app with `pnpm exec expo run:ios` or `pnpm exec expo run:android`.
