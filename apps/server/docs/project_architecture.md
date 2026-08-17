# Production-Ready Go/Gin Project Architecture

This document defines the architecture for a new production Go API inspired by the directory and layering patterns used by Apache Answer, while intentionally using a smaller stack and a simpler operational model.

The selected stack is:

- **HTTP framework:** Gin
- **Dependency injection / application lifecycle:** Uber Fx
- **ORM:** GORM
- **Database:** SQLite
- **Database migrations:** Goose (versioned SQL migrations)
- **Structured logging:** Uber Zap
- **Architecture style:** layered, Apache Answer-inspired
- **Application code location:** `internal/`

The goal is not to clone Apache Answer. The goal is to reuse its strongest ideas around package boundaries, request flow, service ownership, repositories, schemas, entities, routing, and operational structure while keeping the new service compact.

---

## 1. Architecture goals

The project should make it easy to:

- understand the full path of an HTTP request;
- keep business logic independent of Gin and GORM;
- construct the application through explicit Fx modules;
- test services without starting Gin, Fx, or a database;
- test repositories against a real SQLite database;
- keep API schemas separate from persistence entities;
- replace SQLite later without rewriting business logic;
- use one consistent structured logging system;
- add authentication, background work, caching, tracing, or other infrastructure without collapsing package boundaries;
- shut down cleanly and predictably.

The architecture should start small. New infrastructure should be introduced only when a real requirement appears.

---

## 2. Core dependency rule

Dependencies point toward the business logic.

```text
Client
  |
  v
Router
  |
  v
Middleware
  |
  v
Controller
  |
  v
Service
  |
  v
Repository interface
  ^
  |
Repository implementation
  |
  v
GORM
  |
  v
SQLite
```

The normal request path is:

```text
route -> middleware -> controller -> service -> repository -> GORM -> SQLite
```

### Rules

1. Controllers may depend on Gin and application services.
2. Services use `context.Context`, never `*gin.Context`.
3. Services own business rules, authorization decisions, and transaction boundaries.
4. Repositories own persistence queries and GORM-specific behavior.
5. Services do not import GORM.
6. API request/response schemas are separate from persistence entities.
7. Lower layers must not import controllers, routers, Gin, or HTTP-specific types.
8. Fx is primarily a composition and lifecycle concern, not a business-logic dependency.
9. Zap is the single application logger; packages should not create independent global loggers.
10. Cross-feature imports must be intentional and should not create circular package dependencies.

---

## 3. Recommended directory structure

```text
.
├── cmd/
│   └── api/
│       └── main.go
│
├── configs/
│   └── config.example.yaml
│
├── docs/
│   └── PROJECT_ARCHITECTURE.md
│
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   └── module.go
│   │
│   ├── config/
│   │   ├── config.go
│   │   ├── validate.go
│   │   └── module.go
│   │
│   ├── server/
│   │   ├── http.go
│   │   └── module.go
│   │
│   ├── router/
│   │   ├── router.go
│   │   └── module.go
│   │
│   ├── middleware/
│   │   ├── request_id.go
│   │   ├── recovery.go
│   │   ├── logging.go
│   │   ├── security_headers.go
│   │   ├── auth.go
│   │   └── module.go
│   │
│   ├── handler/
│   │   ├── binding.go
│   │   ├── error.go
│   │   └── response.go
│   │
│   ├── controller/
│   │   ├── task_controller.go
│   │   └── module.go
│   │
│   ├── service/
│   │   └── task/
│   │       ├── service.go
│   │       ├── interface.go
│   │       ├── module.go
│   │       └── service_test.go
│   │
│   ├── repository/
│   │   └── task/
│   │       ├── repository.go
│   │       ├── module.go
│   │       └── repository_test.go
│   │
│   ├── entity/
│   │   └── task.go
│   │
│   ├── schema/
│   │   └── task.go
│   │
│   ├── database/
│   │   ├── database.go
│   │   ├── sqlite.go
│   │   ├── transaction.go
│   │   └── module.go
│   │
│   ├── migrations/
│   │   └── sql/
│   │       ├── 00001_init.sql
│   │       └── 00002_add_task_status.sql
│   │
│   └── observability/
│       ├── logger.go
│       └── module.go
│
├── pkg/
│
├── scripts/
│
├── data/
│   └── .gitkeep
│
├── .env.example
├── .gitignore
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

### Directory guidance

Prefer `internal/` for application implementation.

Use `pkg/` only for code intentionally designed to be imported by other Go modules. Do not use `pkg/` as a generic utilities directory.

For a very small service, a feature-first structure may also be used:

```text
internal/task/
├── controller.go
├── service.go
├── repository.go
├── schema.go
└── module.go
```

However, for this project the default is the **layer-first structure**, because it remains close to Apache Answer and makes transport, service, persistence, and infrastructure concerns easy to locate.

---

## 4. Package responsibilities

## 4.1 `cmd/api`

`main.go` should remain extremely small.

Responsibilities:

- construct the Fx application;
- include the root application module;
- call `Run()`.

It should not contain:

- routes;
- GORM queries;
- business logic;
- logger construction;
- configuration parsing;
- migrations;
- HTTP handlers.

Example:

```go
package main

import (
    "go.uber.org/fx"

    "example.com/project/internal/app"
)

func main() {
    fx.New(
        app.Module,
    ).Run()
}
```

---

## 4.2 `internal/app`

This is the application composition root.

It combines the top-level Fx modules.

Example:

```go
var Module = fx.Module(
    "app",
    config.Module,
    observability.Module,
    database.Module,
    repository.Module,
    service.Module,
    controller.Module,
    middleware.Module,
    router.Module,
    server.Module,
)
```

The app package should contain composition, not business logic.

---

## 4.3 `internal/config`

Configuration is typed and validated at startup.

Example:

```go
type Config struct {
    Environment string

    HTTP struct {
        Host            string
        Port            int
        ReadTimeout     time.Duration
        WriteTimeout    time.Duration
        IdleTimeout     time.Duration
        ShutdownTimeout time.Duration
    }

    Database struct {
        Path        string
        BusyTimeout time.Duration
        WAL         bool
    }

    Log struct {
        Level       string
        Development bool
    }
}
```

Recommended precedence:

```text
defaults < config file < environment variables < command-line flags
```

Production rules:

- fail startup if required configuration is invalid;
- never log secrets;
- keep example configuration non-sensitive;
- use absolute or clearly resolved filesystem paths for production SQLite files;
- validate that the database directory is writable before serving traffic.

---

## 4.4 `internal/observability`

Zap is the application logger.

The package owns logger construction and Fx registration.

Default policy:

- production: JSON logs;
- development: human-readable development encoder may be used;
- log level comes from typed configuration;
- logger is injected as `*zap.Logger`;
- prefer typed fields over loosely typed key/value logging;
- call `Sync()` during application shutdown;
- do not use package-global loggers in application code.

Example constructor shape:

```go
func NewLogger(cfg config.Config) (*zap.Logger, error) {
    // build zap.Config from application config
    // return configured *zap.Logger
}
```

Fx module:

```go
var Module = fx.Module(
    "observability",
    fx.Provide(NewLogger),
    fx.Invoke(RegisterLoggerLifecycle),
)
```

Lifecycle cleanup:

```go
func RegisterLoggerLifecycle(
    lc fx.Lifecycle,
    logger *zap.Logger,
) {
    lc.Append(fx.Hook{
        OnStop: func(ctx context.Context) error {
            return logger.Sync()
        },
    })
}
```

Some operating systems may return harmless sync errors for stdout/stderr. If necessary, normalize those errors in the observability package rather than scattering special cases throughout the application.

---

## 4.5 `internal/database`

This package owns SQLite and GORM construction.

It provides:

- `*gorm.DB`;
- access to the underlying `*sql.DB` when required for pool/lifecycle operations;
- SQLite initialization;
- transaction support;
- database shutdown hooks.

The rest of the application should not construct database connections.

Example:

```go
func New(cfg config.Config, logger *zap.Logger) (*gorm.DB, error) {
    db, err := gorm.Open(
        sqlite.Open(cfg.Database.Path),
        &gorm.Config{},
    )
    if err != nil {
        return nil, fmt.Errorf("open sqlite database: %w", err)
    }

    return db, nil
}
```

The concrete implementation should additionally configure SQLite behavior explicitly.

---

## 5. SQLite strategy

SQLite is the primary database for this project.

This is an intentional choice, not a temporary mock database.

### 5.1 Database file

Recommended development path:

```text
./data/app.db
```

Production should use a persistent mounted directory, for example:

```text
/data/app.db
```

The database file must not live in a temporary container filesystem if persistence is required.

### 5.2 WAL mode

Enable Write-Ahead Logging unless deployment constraints require otherwise:

```sql
PRAGMA journal_mode=WAL;
```

WAL allows readers and a writer to operate concurrently more effectively than the default rollback journal model, but SQLite still permits only one writer at a time.

Therefore this architecture is appropriate for workloads where:

- the service has moderate write concurrency;
- the database is local to the application host;
- horizontal replicas do not independently write to the same SQLite file over a network filesystem.

Do not place a WAL-mode SQLite database on a network filesystem and expect normal multi-host database semantics.

### 5.3 Foreign keys

Foreign key enforcement must be enabled explicitly:

```sql
PRAGMA foreign_keys=ON;
```

Do not rely on environment defaults.

### 5.4 Busy handling

Configure a bounded busy timeout so short writer contention does not immediately fail requests.

Example policy:

```text
busy timeout: 5 seconds
```

The exact value should be configurable and tuned from observed workload.

### 5.5 Connection pool

Because SQLite has different concurrency characteristics from PostgreSQL or MySQL, database/sql pool settings should be conservative.

Initial policy:

```text
MaxOpenConns:  1-4, workload dependent
MaxIdleConns:  small
ConnMaxLifetime: bounded if needed
```

Do not copy PostgreSQL-sized pool settings into SQLite.

Measure before increasing concurrency.

### 5.6 Deployment constraint

The official GORM SQLite driver uses the `go-sqlite3` driver and therefore requires CGO.

The build environment and final container strategy must account for that requirement.

If a CGO-free build becomes a hard requirement later, the driver can be reconsidered without changing service interfaces or repository contracts.

---

## 6. GORM rules

GORM belongs in the repository and database layers.

Good dependency direction:

```text
service
  |
  v
repository interface
  ^
  |
repository implementation
  |
  v
*gorm.DB
```

Avoid:

```text
service -> *gorm.DB
```

### 6.1 Repository example

```go
type Repository struct {
    db *gorm.DB
}

func New(db *gorm.DB) *Repository {
    return &Repository{db: db}
}

func (r *Repository) GetByID(
    ctx context.Context,
    id uint,
) (*entity.Task, error) {
    var task entity.Task

    err := r.db.
        WithContext(ctx).
        First(&task, id).
        Error
    if err != nil {
        return nil, mapPersistenceError(err)
    }

    return &task, nil
}
```

### 6.2 Context

Every repository operation must accept `context.Context` and attach it to GORM operations.

```go
r.db.WithContext(ctx)
```

This allows request cancellation and deadlines to flow into database operations.

### 6.3 Persistence errors

Do not return raw GORM or SQLite errors directly to controllers.

Repository code should translate relevant persistence errors into stable application errors, for example:

```text
record not found -> ErrNotFound
unique constraint -> ErrConflict
busy/locked -> ErrDependencyBusy or retryable internal error
unexpected DB failure -> wrapped internal error
```

### 6.4 Query ownership

Repositories may contain:

- GORM queries;
- joins;
- preloads;
- persistence-specific filtering;
- pagination queries;
- database constraints/error translation.

Repositories should not contain:

- permission policy;
- subscription rules;
- workflow orchestration;
- HTTP status decisions;
- request binding.

---

## 7. Entities and schemas

## 7.1 Entities

Entities represent persisted or core domain state.

GORM annotations are allowed here.

Example:

```go
type Task struct {
    ID        uint      `gorm:"primaryKey"`
    OwnerID   uint      `gorm:"not null;index"`
    Title     string    `gorm:"not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

Entities must not contain:

- Gin bindings;
- controller logic;
- HTTP status codes;
- transport-only presentation fields.

## 7.2 Schemas

Schemas define the public API contract.

Example:

```go
type CreateTaskRequest struct {
    Title string `json:"title" binding:"required,min=3,max=200"`
}

type TaskResponse struct {
    ID        uint      `json:"id"`
    Title     string    `json:"title"`
    CreatedAt time.Time `json:"created_at"`
}
```

Never bind HTTP input directly into a GORM entity.

Avoid:

```go
var task entity.Task
_ = ctx.ShouldBindJSON(&task)
```

Use:

```go
var req schema.CreateTaskRequest
_ = ctx.ShouldBindJSON(&req)
```

Then let the service construct or modify domain state.

---

## 8. Services

Services implement use cases.

They own:

- business invariants;
- authorization decisions;
- orchestration across repositories;
- interaction with external services;
- transaction boundaries;
- mapping domain results into response-ready application values where appropriate.

Example interface consumed by a controller:

```go
type TaskService interface {
    Create(
        ctx context.Context,
        actorID uint,
        req schema.CreateTaskRequest,
    ) (*schema.TaskResponse, error)
}
```

Example repository interface consumed by a service:

```go
type TaskRepository interface {
    Create(ctx context.Context, task *entity.Task) error
    GetByID(ctx context.Context, id uint) (*entity.Task, error)
}
```

Interfaces should normally be declared by the package that consumes them.

That means the service package usually owns the repository interface it needs.

---

## 9. Transactions

The service layer owns transaction boundaries because it understands the complete use case.

The repository layer executes persistence operations inside the transaction supplied by the application transaction mechanism.

A transaction manager abstraction is preferred over exposing GORM transactions directly to services.

Example conceptual API:

```go
type TransactionManager interface {
    WithinTransaction(
        ctx context.Context,
        fn func(context.Context) error,
    ) error
}
```

Usage:

```go
err := txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
    if err := accountRepo.Debit(txCtx, fromID, amount); err != nil {
        return err
    }

    return accountRepo.Credit(txCtx, toID, amount)
})
```

This keeps GORM transaction objects out of the business-service API.

---

## 10. Uber Fx dependency injection

Fx is the dependency graph and lifecycle framework.

Use it to construct:

- configuration;
- logger;
- database;
- repositories;
- services;
- controllers;
- middleware;
- router;
- HTTP server;
- long-running workers.

### 10.1 Module pattern

Each substantial package may expose a module.

Example:

```go
var Module = fx.Module(
    "task-repository",
    fx.Provide(New),
)
```

Service:

```go
var Module = fx.Module(
    "task-service",
    fx.Provide(New),
)
```

Controller:

```go
var Module = fx.Module(
    "task-controller",
    fx.Provide(New),
)
```

### 10.2 Constructors remain ordinary Go

Business constructors should still be directly callable in tests.

Prefer:

```go
func New(repo TaskRepository, logger *zap.Logger) *Service {
    return &Service{
        repo:   repo,
        logger: logger,
    }
}
```

Do not require an Fx application just to instantiate a service in a unit test.

### 10.3 `fx.In` and `fx.Out`

Use parameter/result objects when constructor dependency lists become large or when a module naturally benefits from Fx annotations.

Example:

```go
type Params struct {
    fx.In

    DB     *gorm.DB
    Logger *zap.Logger
}
```

Do not introduce `fx.In` mechanically into every constructor.

### 10.4 `fx.Invoke`

Use `fx.Invoke` sparingly.

Good uses include:

- registering routes when registration is side-effect based;
- ensuring the HTTP server root is instantiated;
- starting a worker root;
- wiring lifecycle-owned roots.

Do not use `fx.Invoke` as a substitute for normal dependency construction.

---

## 11. HTTP server and Fx lifecycle

`internal/server` owns `http.Server`.

The server should use Fx lifecycle hooks.

Conceptual pattern:

```go
func RegisterLifecycle(
    lc fx.Lifecycle,
    srv *http.Server,
    logger *zap.Logger,
) {
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            go func() {
                if err := srv.ListenAndServe(); err != nil &&
                    !errors.Is(err, http.ErrServerClosed) {
                    logger.Error("http server failed", zap.Error(err))
                }
            }()
            return nil
        },
        OnStop: func(ctx context.Context) error {
            return srv.Shutdown(ctx)
        },
    })
}
```

Fx lifecycle hooks should start or stop work; they should not block forever.

---

## 12. Router

The router package defines endpoint registration and route groups.

Example:

```go
api := engine.Group("/api/v1")

api.GET("/health/live", healthController.Live)
api.GET("/health/ready", healthController.Ready)

public := api.Group("")
public.POST("/sessions", sessionController.Create)

authenticated := api.Group("")
authenticated.Use(authMiddleware.RequireUser())
authenticated.GET("/tasks/:id", taskController.Get)

admin := api.Group("/admin")
admin.Use(
    authMiddleware.RequireUser(),
    authMiddleware.RequireRole("admin"),
)
admin.DELETE("/tasks/:id", taskController.Delete)
```

Route middleware establishes broad access and request identity.

Resource-specific authorization still belongs in services.

---

## 13. Controllers

Controllers translate HTTP into an application operation.

A controller should normally:

1. bind input;
2. validate input;
3. obtain authenticated identity;
4. call one service use case;
5. map the result through centralized response handling.

Example:

```go
func (c *TaskController) Create(ctx *gin.Context) {
    var req schema.CreateTaskRequest
    if !handler.BindAndValidate(ctx, &req) {
        return
    }

    actorID := middleware.ActorID(ctx)

    result, err := c.tasks.Create(
        ctx.Request.Context(),
        actorID,
        req,
    )

    handler.Respond(ctx, err, result)
}
```

Controllers must not:

- issue GORM queries;
- call SQLite directly;
- coordinate complex workflows;
- implement reusable business policy.

---

## 14. HTTP response contract

Use one stable response format.

Success:

```json
{
  "data": {},
  "error": null,
  "request_id": "01J..."
}
```

Error:

```json
{
  "data": null,
  "error": {
    "code": "task_not_found",
    "message": "task was not found",
    "fields": []
  },
  "request_id": "01J..."
}
```

Never expose:

- raw GORM errors;
- raw SQLite errors;
- stack traces;
- internal filesystem paths;
- tokens;
- secrets;
- dependency credentials.

---

## 15. Application error taxonomy

Application errors should be independent from HTTP.

| Category | HTTP status | Retryable |
| --- | ---: | --- |
| validation | 400 / 422 | no |
| unauthenticated | 401 | after auth |
| forbidden | 403 | no |
| not found | 404 | no |
| conflict | 409 | depends |
| rate limited | 429 | yes |
| database busy | 503 or controlled retry | yes |
| dependency unavailable | 503 | yes |
| unexpected | 500 | maybe |

Centralize HTTP error mapping in `internal/handler`.

Unexpected errors should normally be logged once at the HTTP boundary with the request ID and wrapped internal cause.

---

## 16. Middleware order

Install middleware deliberately.

Recommended order:

```text
request ID
-> recovery
-> security headers
-> access logging
-> metrics/tracing when introduced
-> body-size limit
-> CORS when required
-> authentication
-> route handler
```

The logger middleware should enrich logs with fields such as:

- request ID;
- method;
- route template;
- status;
- duration;
- remote IP when appropriate;
- actor/user ID when safe;
- error code.

Use the Gin route template where possible instead of raw resource IDs for metric labels.

---

## 17. Zap logging conventions

Use structured logs.

Prefer:

```go
logger.Info(
    "task created",
    zap.Uint("task_id", task.ID),
    zap.Uint("actor_id", actorID),
)
```

Avoid:

```go
logger.Info(fmt.Sprintf("task %d created by %d", task.ID, actorID))
```

### Standard fields

Where applicable:

```text
service
version
environment
request_id
trace_id
actor_id
method
route
status
duration
error_code
```

### Logging rules

- never log passwords;
- never log auth tokens;
- avoid logging full request bodies by default;
- do not log secrets from configuration;
- do not log the same error at every layer;
- add context as an error travels upward, then log at the boundary responsible for handling it.

---

## 18. GORM logger integration

GORM logging should integrate with the application logging policy rather than becoming an independent noisy logging subsystem.

Desired behavior:

- development: SQL logging may be enabled at an appropriate level;
- production: default to warnings/errors and slow-query visibility;
- attach request context where practical;
- never log sensitive SQL parameters indiscriminately;
- avoid duplicate application-error and SQL-error logs.

A small adapter around GORM's logger interface may be created if required so database logs flow into Zap consistently.

Keep this adapter in `internal/database` or `internal/observability`, not in repositories.

---

## 19. Database migrations with Goose

**Goose is the official migration system for this project.**

GORM is responsible for runtime persistence and object mapping. Goose is responsible for evolving the database schema over time.

```text
GORM models       -> runtime mapping / queries
Goose migrations  -> schema history / deployment changes
SQLite            -> database
```

Do **not** use GORM `AutoMigrate` as the production deployment migration system. `AutoMigrate` may be used for disposable test databases or short-lived prototypes, but production schema changes must be explicit, versioned, reviewable, and reproducible through Goose.

### 19.1 Migration layout

```text
internal/migrations/
└── sql/
    ├── 00001_init.sql
    ├── 00002_add_task_status.sql
    └── 00003_add_task_indexes.sql
```

Each Goose SQL migration contains both the forward and rollback operations when a safe rollback is possible.

Example:

```sql
-- +goose Up
CREATE TABLE tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX idx_tasks_status ON tasks(status);

-- +goose Down
DROP TABLE tasks;
```

### 19.2 Migration commands

Expose the common Goose operations through the `Makefile` so developers and CI use the same commands.

```makefile
DB_PATH ?= ./data/app.db
MIGRATIONS_DIR := ./internal/migrations/sql

migrate-create:
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

migrate-up:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) status

migrate-validate:
	goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH) validate
```

Typical usage:

```bash
make migrate-create name=create_users
make migrate-up
make migrate-status
```

The migration command must receive the same database path/configuration used by the application for the target environment. Do not hard-code production database locations into migration files or scripts.

### 19.3 Deployment policy

Migrations are an explicit release operation. Do not have every Fx application process automatically execute pending migrations during normal startup.

Preferred deployment flow:

```text
build artifact
    -> back up database when required
    -> run Goose migrations once
    -> start/restart application
    -> readiness succeeds
```

For a single-instance deployment, a dedicated deploy command or release script may run Goose immediately before starting the new application version. If multiple replicas are introduced later, only one release/migration job should apply migrations.

### 19.4 Migration rules

- use one migration per logical schema change;
- never edit a migration already applied in a shared environment;
- create a new migration to correct or extend an existing one;
- test the complete migration chain against SQLite in CI;
- test both a clean install and upgrades from supported existing schemas;
- back up production data before destructive or high-risk migrations;
- use transactions where SQLite and the migration operation safely support them;
- account for SQLite table-rebuild requirements for unsupported `ALTER TABLE` operations;
- keep application code compatible with staged migrations when a schema change cannot be deployed atomically;
- write `Down` migrations only when rollback is actually safe; an unsafe rollback should be documented rather than pretending it is reversible.

### 19.5 `AutoMigrate` policy

Allowed:

```text
repository unit/integration test setup with a disposable database
throwaway local prototypes
```

Not allowed:

```text
production application startup
shared development/staging environments
release deployment schema management
```

The rule is:

> **GORM models describe how application code accesses data. Goose migrations define how the database schema changes.**

---

## 20. Database backup policy

Because SQLite stores application state in files, backup policy is part of database architecture.

Minimum expectations:

- database lives on persistent storage;
- automated backups exist for production;
- backup and restore procedures are tested;
- WAL-aware backup procedures are used;
- destructive migrations require a recent verified backup;
- database path and backup location are configurable.

Never treat copying an actively changing database file blindly as a guaranteed consistent backup strategy.

---

## 21. Authentication and security baseline

When authentication is introduced:

- validate token signature, issuer, audience, expiry, and algorithm;
- use secure cookie flags for cookie-based auth;
- use CSRF protection for cookie-authenticated state changes;
- restrict CORS to known origins;
- configure Gin trusted proxies explicitly;
- limit request body size;
- rate-limit login, reset, registration, and expensive endpoints;
- sanitize uploaded filenames and validate content;
- store uploads outside the executable directory;
- prevent personal data and secrets from entering logs;
- use parameterized persistence operations through GORM;
- keep dependencies and the Go toolchain patched.

Authorization belongs at two levels:

```text
middleware -> broad route access
service    -> resource/domain authorization
```

---

## 22. Health endpoints

Expose separate health signals.

```text
GET /health/live
GET /health/ready
```

### Liveness

Answers:

> Is the process alive?

Do not perform expensive dependency checks.

### Readiness

Answers:

> Can this process currently serve requests?

Readiness may verify:

- database initialization completed;
- required application startup state exists;
- the SQLite database is reachable and operational.

---

## 23. Graceful shutdown

Fx owns lifecycle orchestration.

On shutdown:

1. stop accepting new HTTP requests;
2. allow in-flight requests to complete within the configured timeout;
3. stop background workers;
4. flush telemetry when present;
5. sync Zap;
6. close the underlying database connection.

Long-running operations must accept cancellation via `context.Context`.

---

## 24. Background work

Controllers must not launch unmanaged goroutines for durable work.

For small non-durable asynchronous tasks, lifecycle-aware workers may be introduced carefully.

For durable jobs later:

- use a durable queue;
- assign idempotency keys;
- define retry limits;
- use exponential backoff with jitter;
- distinguish retryable and permanent failures;
- provide operational visibility;
- make handlers safe to execute more than once.

The queue implementation is intentionally not selected yet.

---

## 25. Testing strategy

Use the narrowest test that proves behavior.

| Layer | Test style | Main assertions |
| --- | --- | --- |
| service | unit with fakes/mocks | business rules and error paths |
| repository | integration with SQLite | queries, constraints, transactions |
| controller | `httptest` | binding, status, response contract |
| middleware | `httptest` | context, headers, rejection behavior |
| migrations | integration | clean install and upgrade path |
| app | small end-to-end suite | critical user journeys |

### 25.1 Service tests

Service tests should not require:

- Fx;
- Gin;
- GORM;
- SQLite.

Example:

```go
repo := &FakeTaskRepository{}
service := taskservice.New(repo, zap.NewNop())
```

### 25.2 Repository tests

Use real SQLite.

Prefer a temporary file database when behavior involving WAL, file locking, or persistence matters.

Use an in-memory database when the behavior being tested does not depend on filesystem semantics.

### 25.3 Determinism

Inject where useful:

- clock;
- ID generator;
- external clients.

Do not mock everything. Repository tests should prove actual database behavior.

### 25.4 CI test commands

At minimum:

```text
go test ./...
go test -race ./...
go vet ./...
```

Add linting and vulnerability scanning as the project matures.

---

## 26. Fx testing rule

Fx should not be part of most unit tests.

Use Fx tests only to prove the application graph is valid or to exercise lifecycle composition.

A useful architecture test should ensure:

- all required dependencies can be resolved;
- constructors do not conflict;
- modules compose successfully;
- lifecycle hooks can start/stop in a controlled test environment.

Business rules remain ordinary Go tests.

---

## 27. Build and deployment baseline

CI should eventually run:

```text
go mod verify
go vet ./...
static analysis / lint
go test -race ./...
integration tests
goose migration validation
go build ./cmd/api
vulnerability scan
container build
container scan
```

Because the official GORM SQLite driver requires CGO, the build pipeline must include the required C toolchain.

A multi-stage container build is recommended.

Runtime requirements:

- non-root user where possible;
- writable mounted directory for SQLite;
- stdout/stderr logging;
- minimal runtime image compatible with the CGO-linked binary;
- explicit health checks;
- graceful termination support.

---

## 28. Production SQLite deployment model

Preferred initial deployment:

```text
               ┌───────────────────┐
Internet ----> │ Go API process    │
               │ Gin + Fx          │
               │ GORM              │
               └─────────┬─────────┘
                         │
                         v
               ┌───────────────────┐
               │ local persistent  │
               │ SQLite database   │
               │ /data/app.db      │
               └───────────────────┘
```

Prefer one application instance with durable local or attached block storage unless the architecture is deliberately extended for replicated/multi-writer operation.

Do not assume a shared SQLite file is a substitute for a client/server database in horizontally scaled multi-host deployments.

If the product eventually requires high concurrent writes or multiple independent application replicas writing concurrently, the repository boundary should make migration to PostgreSQL substantially easier.

---

## 29. Initial Fx module graph

```text
app.Module
│
├── config.Module
├── observability.Module
├── database.Module
│
├── repository.Module
│   └── taskrepo.Module
│
├── service.Module
│   └── taskservice.Module
│
├── controller.Module
│   └── taskcontroller.Module
│
├── middleware.Module
├── router.Module
└── server.Module
```

Dependency graph:

```text
Config
 ├───────────────> Zap Logger
 ├───────────────> GORM / SQLite
 └───────────────> HTTP Server

GORM / SQLite
      |
      v
Repositories
      |
      v
Services
      |
      v
Controllers
      |
      v
Router / Gin
      |
      v
HTTP Server
```

---

## 30. Example feature flow

For a `task` feature:

```text
POST /api/v1/tasks
        |
        v
request ID middleware
        |
        v
auth middleware
        |
        v
TaskController.Create
        |
        v
TaskService.Create
        |
        v
TaskRepository.Create
        |
        v
GORM
        |
        v
SQLite
```

Responsibilities:

```text
Controller
  bind JSON
  validate transport input
  get actor ID
  call service

Service
  validate business rules
  authorize actor
  construct entity
  own transaction decision

Repository
  execute GORM operation
  translate persistence errors

Handler
  translate application error to HTTP
```

---

## 31. Implementation order

Build one vertical slice before adding optional infrastructure.

### Phase 1 — application shell

1. initialize Go module;
2. create `cmd/api`;
3. add typed configuration;
4. add Uber Fx root application;
5. add Zap logger;
6. add Gin server;
7. add graceful start/stop lifecycle.

### Phase 2 — persistence

1. add GORM;
2. add SQLite driver;
3. configure database path;
4. enable required SQLite settings;
5. create transaction manager;
6. add Goose and create the first versioned SQL migration;
7. add Makefile commands for Goose (`create`, `up`, `down`, `status`, `validate`);
8. add readiness database check.

### Phase 3 — first vertical feature

Implement one resource end-to-end:

```text
schema
-> entity
-> repository
-> service
-> controller
-> route
```

Suggested first feature:

```text
Task
```

Implement:

```text
POST   /api/v1/tasks
GET    /api/v1/tasks/:id
GET    /api/v1/tasks
PATCH  /api/v1/tasks/:id
DELETE /api/v1/tasks/:id
```

### Phase 4 — shared HTTP behavior

Add:

- centralized response envelope;
- application error model;
- validation handling;
- request ID;
- recovery;
- structured access logs;
- body size limits;
- security headers.

### Phase 5 — testing baseline

Add:

- service unit tests;
- repository SQLite integration tests;
- controller tests;
- middleware tests;
- migration tests;
- Fx graph smoke test.

### Phase 6 — authentication

Add only after the core vertical slice is sound:

- identity model;
- authentication middleware;
- service-level authorization;
- auth rate limits;
- security tests.

### Phase 7 — operational maturity

Add as required:

- metrics;
- OpenTelemetry tracing;
- background queues;
- caching;
- external integrations;
- richer health checks.

---

## 32. Definition of done for a new endpoint

An endpoint is complete when:

- its route is registered in the correct group;
- request and response schemas are explicit;
- input is validated;
- body size constraints are appropriate;
- controller contains only transport logic;
- business logic lives in a service;
- service accepts `context.Context`;
- resource-level authorization is enforced where required;
- persistence is behind a repository interface;
- repository uses `WithContext` for GORM operations;
- persistence errors are translated;
- no raw GORM or SQLite error reaches the client;
- logs contain useful structured context;
- success and expected failure paths are tested;
- required migration is included;
- timeout/retry/idempotency behavior is defined when relevant.

---

## 33. Architectural anti-patterns

Avoid these patterns.

### Fat controller

```go
func Create(ctx *gin.Context) {
    // bind
    // query DB
    // check permissions
    // update multiple tables
    // send external request
    // build response
}
```

Move orchestration into the service.

### GORM in service

```go
type Service struct {
    db *gorm.DB
}
```

Prefer repository interfaces.

### Gin context in service

```go
func (s *Service) Create(ctx *gin.Context) error
```

Use:

```go
func (s *Service) Create(ctx context.Context, ...) error
```

### Binding into entity

```go
ctx.ShouldBindJSON(&entity.Task{})
```

Bind into a schema DTO.

### Global logger

```go
var Log = zap.NewNop()
```

Inject the application logger.

### Fx everywhere

Do not make every business type depend on `fx.In`, `fx.Out`, or `fx.Lifecycle`.

Fx is primarily the composition framework.

### Automatic schema mutation in production

Do not let every application process run unconstrained `AutoMigrate` at startup.

Use Goose-managed explicit versioned SQL migrations.

---

## 34. Final technology decisions

| Concern | Decision |
| --- | --- |
| Language | Go |
| HTTP | Gin |
| DI | Uber Fx |
| Application lifecycle | Uber Fx |
| ORM | GORM |
| Database | SQLite |
| SQLite driver | official GORM SQLite driver initially |
| Logging | Uber Zap |
| API structure | controller -> service -> repository |
| Request DTOs | `internal/schema` |
| Persistence models | `internal/entity` |
| Configuration | typed startup configuration |
| Migrations | Goose with explicit versioned SQL migrations |
| Tests | unit + real SQLite integration + httptest |
| Application code | `internal/` |
| Public reusable code | `pkg/` only when genuinely reusable |

---

## 35. Patterns retained from Apache Answer

This project intentionally retains the following ideas from Apache Answer:

- small executable entry point under `cmd/`;
- centralized HTTP server construction;
- route groups separated by authentication level;
- controller, service, repository, entity, and schema boundaries;
- shared binding and response handling;
- constructor-based dependency injection;
- versioned database migrations;
- `internal/` for application implementation;
- `pkg/` only for genuinely reusable packages.

The implementation differs intentionally in these areas:

```text
Apache Answer-inspired baseline      This project
----------------------------------------------------------
manual/Wire-style composition   ->   Uber Fx
Answer persistence stack        ->   GORM
server SQL database patterns    ->   SQLite
migration mechanism             ->   Goose + versioned SQL
existing logging setup          ->   Uber Zap
large application subsystems    ->   only what this project needs
```

The architecture should remain smaller than Apache Answer until scale or product requirements justify additional infrastructure.

---

## 36. Core rule to remember

For almost every feature, the expected flow is:

```text
router
  -> middleware
  -> controller
  -> service
  -> repository interface
  -> GORM repository implementation
  -> SQLite
```

And application construction is:

```text
Uber Fx
  -> config
  -> Zap
  -> GORM/SQLite
  -> repositories
  -> services
  -> controllers
  -> middleware/router
  -> Gin HTTP server
```

If a future design decision breaks these boundaries, it should have a concrete reason rather than being done for convenience.
