# Backend testing guide

Keep tests small, deterministic, and focused on observable behavior. Use Go's
standard `testing` package and `net/http/httptest`, with Testify for assertions
and `cmp.Diff` for complex values.

## Test categories

- **Unit:** package-local, fast, and without external infrastructure. Use for
  validation, configuration parsing, and domain logic.
- **Integration:** use real SQLite, production SQL migrations, HTTP handlers,
  and GraphQL handlers where practical.
- **E2E:** reserve `tests/e2e` for a small number of critical user workflows.
  Exercise handlers in process; do not require a network listener or external
  services.

Production-package tests belong alongside production code. Shared infrastructure
belongs in `internal/testutil`; package-specific builders and fakes do not.

## Assertions and naming

Use standard Go names such as `TestVerifier_Verify` and use table-driven subtests
for behavior variants.

Use `require` when the test cannot continue after a failure:

```go
require.NoError(t, err)
require.Error(t, err)
require.Equal(t, want, got)
require.NotNil(t, value)
```

Use `assert` only for independent checks that should all run. For nested structs,
slices, and maps, use `cmp.Diff`:

```go
if diff := cmp.Diff(want, got); diff != "" {
	t.Fatalf("result mismatch (-want +got):\n%s", diff)
}
```

Use `ErrorIs` with stable sentinel errors. Avoid asserting error strings unless
they are an explicit external API contract.

## Data, state, and dependencies

- Use real SQLite and the exact production migrations for persistence behavior.
  Never mock GORM or replace schema migrations with `AutoMigrate`.
- Prefer a pure value, a small handwritten fake, or a lightweight real dependency
  before a mock. Do not introduce a mocking framework by default.
- Do not test generated gqlgen code, Zap, GORM, SQLite, or HTTP framework internals.
- Use `t.TempDir()` and `t.Cleanup()` for resources; use `t.Setenv()` for
  environment values.
- Inject time only when it is necessary to make behavior deterministic.
- Add `t.Parallel()` only when state isolation is proven.

Keep GraphQL resolvers thin. Test important GraphQL behavior through the HTTP
boundary, with real authentication and persistence where applicable.

## Required workflow for backend changes

1. Define and test business behavior first.
2. Test persistence with real SQLite and migrations.
3. Add an HTTP GraphQL integration test for the important API path.
4. Add a regression test before fixing reproducible bugs.
5. Include regression coverage for every security fix.
6. Run `make test` before considering work complete.
7. Run `make test-race` before opening or completing a PR.

Coverage is informational. Prioritize authentication, replay protection,
persistence, state transitions, configuration, and API contracts over trivial
constructors or generated code.

## Commands

```bash
make test
make test-race
make test-cover
go vet ./...
```
