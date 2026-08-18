# Backend Testing & TDD Guidelines

Guidelines for writing fast, reliable, and maintainable tests for the Mobicode Go backend.

> **Core Rule**: Test business logic heavily at the service layer, test repositories against real SQLite, test GraphQL at the external API boundary, and mock only where isolation provides clear value.

---

## 1. Stack & Architecture

### Default Stack
- **Standard Library**: `testing`, `net/http/httptest`
- **Assertions**: `github.com/stretchr/testify/require` (fatal), `github.com/stretchr/testify/assert` (non-fatal)
- **Optional Mocking**: `go.uber.org/mock` (only when handwritten fakes become repetitive/complex)

### Testing Pyramid
```text
      ┌──────────────┐
      │     E2E      │  Small / Minimal
      └──────┬───────┘
   ┌─────────┴─────────┐
   │  GraphQL / HTTP   │  API Boundary (`httptest`)
   └─────────┬─────────┘
 ┌───────────┴───────────┐
 │   Repository / DB     │  Real SQLite + Production Migrations
 └───────────┬───────────┘
┌────────────┴────────────┐
│    Service / Domain     │  Majority of Test Suite (Fast, Pure, Isolated)
└─────────────────────────┘
```

---

## 2. File Organization

Co-locate test files alongside code. Put shared helpers in `internal/testutil/`.

```text
internal/
├── auth/
│   ├── jwt.go
│   ├── jwt_test.go
│   ├── service.go
│   └── service_test.go
├── user/
│   ├── repository.go
│   ├── repository_test.go
│   ├── resolver.go
│   └── service_test.go
└── testutil/            # Cross-cutting test setup & fixtures
    ├── database.go
    ├── graphql.go
    └── fixtures.go
```

*Note: Exclude generated code (`graphql/generated/`, `graphql/model/`) from manual test writing and coverage targets.*

---

## 3. Testing Layers

### Service & Domain (Primary TDD Boundary)
- **Scope**: Business logic, validation, authorization, state transitions, token rotation, domain errors.
- **Rules**: Must be fast, deterministic, and pure—no DB, HTTP, GraphQL, GORM, or Uber Fx dependencies.
- **Dependencies**: Use small, consumer-focused interfaces. Prefer handwritten fakes over generated mocks.

```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByEmail(ctx context.Context, email string) (*User, error)
}

type fakeRepository struct {
    created *User
    err     error
}

func (f *fakeRepository) Create(ctx context.Context, user *User) error {
    if f.err != nil { return f.err }
    f.created = user
    return nil
}
```

### Repositories (Real SQLite)
- **Scope**: Queries, uniqueness, foreign keys, soft deletes, transactions, database error mapping.
- **Rules**: Never mock GORM or SQLite. Never use dev database (`data/app.db`).
- **Isolation**: Use `t.TempDir()` with real embedded migrations (`NewMigratedTestDB(t)`).

```go
func NewTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    dbPath := filepath.Join(t.TempDir(), "test.db")
    db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
    require.NoError(t, err)
    // Run embedded migrations here
    return db
}
```

### GraphQL & HTTP Boundary
- **Scope**: Schema mapping, mutation/query execution, auth context passing, HTTP status codes, GraphQL error formats.
- **Rules**: Use `net/http/httptest`. Test API boundary behavior only—do **not** duplicate detailed service business rule variations.

### Auth & Middleware
- Test password hashing/verification, JWT parsing/expiration, refresh token rotation, and Chi middleware using `httptest`. Test Mobicode logic, not standard lib/bcrypt internals.

### Uber Fx
- Do **not** test `fx.Provide` calls or container wiring directly. Instantiate constructors directly (`NewService(repo)`). Optional: single app-level dependency graph start/stop integration test.

---

## 4. Isolation & Mocking Rules

| Mock / Fake Candidates | Do NOT Mock |
| :--- | :--- |
| Repositories (in service unit tests) | GORM & SQLite internals |
| Clock / Time abstraction (`Clock` interface) | Zap logger |
| Random token generators | Chi router & `httptest` internals |
| External HTTP API clients / Storage / Email | `gqlgen` generated code |
| Remote agent infrastructure | Standard library & bcrypt |

### Time & Environment
- **Time**: Inject a `Clock` interface (`RealClock` vs `FakeClock`) only when domain logic explicitly depends on time.
- **Env**: Use `t.Setenv("KEY", "val")` for isolated environment variable overrides.
- **State**: Avoid global mutable state. Use constructor dependency injection.

---

## 5. TDD & Feature Workflow

### Cycle: RED → GREEN → REFACTOR

```text
1. Define behavior → Write failing test (RED)
2. Minimal implementation → Make test pass (GREEN)
3. Clean up code without changing behavior (REFACTOR)
```

### Feature Layer Sequence (e.g. `createWorkspace`)
1. **Service**: Write service test (RED) → Interface/Service impl (GREEN) → Refactor.
2. **Repository**: Write DB integration test with SQLite (RED) → GORM query impl (GREEN) → Refactor.
3. **GraphQL**: Write API `httptest` (RED) → Resolver/schema impl (GREEN) → Refactor.
4. **Verification**: Run `go test ./...` and `go test -race ./...`.

---

## 6. Patterns & Conventions

### Table-Driven Tests & Naming
Use `Test<Type>_<Method>_<Scenario>` and subtests (`t.Run`):

```go
func TestService_Register(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {name: "valid user", email: "john@example.com", wantErr: false},
        {name: "invalid email", email: "invalid", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            err := ValidateEmail(tt.email)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
        })
    }
}
```

### Assertions & Errors
- Use `require.NoError(t, err)` for prerequisite checks; use `assert.Equal` for non-fatal checks.
- Assert sentinel/typed errors with `require.ErrorIs(t, err, ErrUserNotFound)` instead of matching string messages.

### Concurrency & Independence
- Call `t.Parallel()` on pure, thread-safe unit tests. Avoid parallelism when sharing DB or global state.
- Every test must set up its own state and clean up after itself. Never chain test execution order.

---

## 7. Commands Cheat Sheet

```bash
# Run all tests
go test ./...

# Run single package / single test
go test ./internal/user -run TestService_Register

# Run with race detection (crucial for concurrency)
go test -race ./...

# Bypass test cache
go test -count=1 ./...

# Coverage report
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
```

---

## 8. Definition of Done Checklist

- [ ] Core business rules covered by service unit tests.
- [ ] Edge cases and failure paths tested.
- [ ] Database interactions tested against real SQLite with migrations.
- [ ] API endpoints verified at GraphQL / HTTP boundary.
- [ ] `go test ./...` and `go test -race ./...` pass cleanly.
- [ ] No dependency on dev database or global mutable state.
