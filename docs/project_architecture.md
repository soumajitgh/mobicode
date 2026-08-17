# Production-Ready Go/Gin GraphQL-First Project Architecture

This document defines the architecture for a production Go API that keeps the existing
Gin + Uber Fx + GORM + SQLite + Goose + Zap foundation, but makes **GraphQL the primary
application API**.

The architecture remains inspired by Apache Answer's strong package boundaries and
layering, while replacing the REST/controller-centric transport model with a
**schema-first GraphQL transport using gqlgen**.

---

## 1. Final technology decisions

| Concern | Decision |
| --- | --- |
| Language | Go |
| HTTP framework | Gin |
| Primary API | GraphQL |
| GraphQL implementation | gqlgen |
| GraphQL style | Schema-first |
| Dependency injection | Uber Fx |
| Application lifecycle | Uber Fx |
| ORM | GORM |
| Database | SQLite |
| Database migrations | Goose with versioned SQL migrations |
| Structured logging | Uber Zap |
| Application architecture | GraphQL resolver -> service -> repository |
| GraphQL schema | `.graphqls` files |
| GraphQL generated models | gqlgen-generated models where appropriate |
| Persistence models | `internal/entity` |
| Application code | `internal/` |
| Public reusable code | `pkg/` only when genuinely reusable |
| Health endpoints | REST-style Gin endpoints |
| GraphQL subscriptions | Not enabled initially |

The selected stack is therefore:

```text
Client
  |
  v
Gin
  |
  v
gqlgen GraphQL handler
  |
  v
Resolvers
  |
  v
Services
  |
  v
Repository interfaces
  ^
  |
Repository implementations
  |
  v
GORM
  |
  v
SQLite
```

---

## 2. Architecture goals

The project should make it easy to:

- treat GraphQL as the primary public API contract;
- understand the full path of a GraphQL operation;
- keep business logic independent of Gin, gqlgen, GORM, and SQLite;
- construct the application through explicit Uber Fx modules;
- keep GraphQL schema/types separate from persistence entities;
- test services without starting Gin, gqlgen, Fx, or a database;
- test repositories against a real SQLite database;
- avoid GraphQL N+1 query problems;
- apply authentication consistently to GraphQL operations;
- expose stable GraphQL error codes without leaking internal errors;
- limit expensive GraphQL queries;
- replace SQLite later without rewriting business logic;
- add background jobs, caching, tracing, or subscriptions later without collapsing package boundaries;
- shut down cleanly and predictably.

The architecture should remain small. GraphQL infrastructure should be introduced only
when it has an explicit responsibility.

---

## 3. GraphQL-first core dependency rule

Dependencies point toward business logic.

```text
GraphQL Client
      |
      v
Gin HTTP Middleware
      |
      v
gqlgen Handler
      |
      v
Resolver
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

The normal application request path is:

```text
POST /graphql
    -> Gin middleware
    -> gqlgen operation execution
    -> resolver
    -> service
    -> repository
    -> GORM
    -> SQLite
```

For nested relation fields:

```text
GraphQL field resolver
    -> DataLoader
    -> repository/service
    -> GORM
    -> SQLite
```

### Dependency rules

1. GraphQL resolvers may depend on application services.
2. Resolvers use `context.Context`, not `*gin.Context`.
3. Services use `context.Context` and never depend on gqlgen or Gin.
4. Services own business rules, authorization decisions, orchestration, and transaction boundaries.
5. Repositories own persistence queries and GORM-specific behavior.
6. Services do not import GORM.
7. Persistence entities do not define the GraphQL contract.
8. GraphQL generated types must not become the persistence model by default.
9. Lower layers must not import Gin, gqlgen handlers, GraphQL resolvers, or transport-specific code.
10. Fx is a composition/lifecycle concern, not a business-logic dependency.
11. Zap is the single application logger.
12. Resource authorization must not be implemented only in GraphQL directives or middleware.
13. DataLoader batching is a transport optimization, not a place for business rules.

---

## 4. Recommended directory structure

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
├── gqlgen.yml
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
│   ├── graphql/
│   │   ├── schema/
│   │   │   ├── schema.graphqls
│   │   │   ├── task.graphqls
│   │   │   └── common.graphqls
│   │   │
│   │   ├── generated/
│   │   │   └── generated.go
│   │   │
│   │   ├── model/
│   │   │   └── models_gen.go
│   │   │
│   │   ├── resolver/
│   │   │   ├── resolver.go
│   │   │   ├── query.resolvers.go
│   │   │   ├── mutation.resolvers.go
│   │   │   ├── task.resolvers.go
│   │   │   └── module.go
│   │   │
│   │   ├── directive/
│   │   │   └── auth.go
│   │   │
│   │   ├── scalar/
│   │   │   └── datetime.go
│   │   │
│   │   ├── dataloader/
│   │   │   ├── loader.go
│   │   │   ├── middleware.go
│   │   │   └── module.go
│   │   │
│   │   ├── error/
│   │   │   ├── presenter.go
│   │   │   └── recover.go
│   │   │
│   │   ├── handler.go
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
├── .gitignore
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

### Directory guidance

`internal/graphql` owns the GraphQL transport.

It contains:

- SDL schema files;
- gqlgen-generated execution code;
- GraphQL generated API models;
- resolvers;
- GraphQL directives;
- custom scalar mappings;
- DataLoader integration;
- GraphQL error presentation;
- gqlgen server construction.

It must not contain application business logic.

`internal/service` remains the use-case layer.

`internal/repository` remains the persistence boundary.

`internal/entity` remains the GORM persistence model layer.

---

## 5. Why GraphQL is not placed in `controller/`

The REST architecture used:

```text
controller -> service -> repository
```

The GraphQL-first architecture uses:

```text
resolver -> service -> repository
```

A GraphQL resolver performs the transport role that a REST controller previously
performed.

Therefore the default project should **remove `internal/controller` for the primary API**.

Do not create this:

```text
GraphQL resolver
    -> controller
    -> service
```

That introduces an unnecessary transport-to-transport layer.

Use:

```text
GraphQL resolver
    -> service
```

REST controllers/handlers may still exist later for transport-specific endpoints that
are genuinely not GraphQL concerns, such as:

```text
GET /health/live
GET /health/ready
```

---

## 6. GraphQL schema is the API contract

The GraphQL schema is the primary public API contract.

Example:

```graphql
scalar DateTime

type Task {
  id: ID!
  title: String!
  status: TaskStatus!
  createdAt: DateTime!
  updatedAt: DateTime!
}

enum TaskStatus {
  PENDING
  COMPLETED
}

input CreateTaskInput {
  title: String!
}

input UpdateTaskInput {
  title: String
  status: TaskStatus
}

type TaskConnection {
  nodes: [Task!]!
  pageInfo: PageInfo!
}

type PageInfo {
  hasNextPage: Boolean!
  endCursor: String
}

type Query {
  task(id: ID!): Task
  tasks(first: Int = 20, after: String): TaskConnection!
}

type Mutation {
  createTask(input: CreateTaskInput!): Task!
  updateTask(id: ID!, input: UpdateTaskInput!): Task!
  deleteTask(id: ID!): Boolean!
}
```

### Schema rules

1. Organize schema by domain/feature instead of keeping a single huge schema file.
2. Prefer explicit input types for mutations.
3. Avoid accepting persistence entities as GraphQL inputs.
4. Avoid exposing internal database fields just because they exist.
5. Keep names stable once consumed by clients.
6. Prefer deprecation over sudden field removal.
7. Do not use GraphQL types as an excuse to bypass service-level validation.
8. Keep pagination conventions consistent across list fields.
9. Avoid exposing unbounded list fields.
10. Document nullable vs non-null semantics deliberately.

---

## 7. `gqlgen.yml`

Keep gqlgen configuration at the repository root.

Conceptual configuration:

```yaml
schema:
  - internal/graphql/schema/*.graphqls

exec:
  filename: internal/graphql/generated/generated.go
  package: generated

model:
  filename: internal/graphql/model/models_gen.go
  package: model

resolver:
  layout: follow-schema
  dir: internal/graphql/resolver
  package: resolver
  filename_template: "{name}.resolvers.go"

models:
  ID:
    model:
      - github.com/99designs/gqlgen/graphql.ID
```

Generated files should be treated as generated code.

Do not hand-edit generated execution code.

Recommended commands:

```makefile
graphql-generate:
	go run github.com/99designs/gqlgen generate

graphql-check:
	go run github.com/99designs/gqlgen generate
	git diff --exit-code
```

CI should ensure generated GraphQL code is up to date.

---

## 8. Resolver responsibilities

Resolvers are thin GraphQL adapters.

A resolver should normally:

1. accept gqlgen-provided `context.Context`;
2. read already-established actor identity from context;
3. map GraphQL input into service input;
4. call one application service/use case;
5. map service output into GraphQL output when required;
6. return stable application errors for centralized GraphQL error presentation.

Example:

```go
func (r *mutationResolver) CreateTask(
    ctx context.Context,
    input model.CreateTaskInput,
) (*model.Task, error) {
    actor, err := auth.ActorFromContext(ctx)
    if err != nil {
        return nil, err
    }

    result, err := r.taskService.Create(ctx, task.CreateInput{
        ActorID: actor.ID,
        Title:   input.Title,
    })
    if err != nil {
        return nil, err
    }

    return mapTask(result), nil
}
```

Resolvers must not:

- issue GORM queries;
- open transactions;
- implement reusable permission policy;
- contain workflow orchestration;
- know SQLite details;
- log every error independently;
- depend directly on `*gin.Context`;
- return raw GORM errors.

---

## 9. Resolver root

Use one resolver root object to hold service dependencies required by generated
resolvers.

Example:

```go
type Resolver struct {
    TaskService task.Service
    Logger      *zap.Logger
}
```

Constructor:

```go
func New(
    taskService task.Service,
    logger *zap.Logger,
) *Resolver {
    return &Resolver{
        TaskService: taskService,
        Logger:      logger,
    }
}
```

Do not place mutable request state in the resolver root.

Request-specific state belongs in `context.Context`.

---

## 10. Gin's role in a GraphQL-first application

Gin remains the HTTP server/router layer.

Gin owns:

- HTTP server integration;
- request ID middleware;
- recovery middleware;
- security headers;
- CORS;
- broad authentication extraction;
- body-size limits;
- access logging;
- health routes;
- mounting the gqlgen handler.

Gin does **not** own GraphQL field resolution.

Recommended routes:

```text
POST /graphql      -> GraphQL operations
GET  /graphql      -> GraphQL playground in development only
GET  /health/live
GET  /health/ready
```

A production router should conceptually look like:

```go
func NewRouter(
    gqlHandler http.Handler,
    authMiddleware *middleware.Auth,
) *gin.Engine {
    engine := gin.New()

    engine.Use(
        middleware.RequestID(),
        middleware.Recovery(),
        middleware.SecurityHeaders(),
        middleware.AccessLog(),
        middleware.BodyLimit(),
        middleware.CORS(),
        authMiddleware.Optional(),
    )

    engine.POST("/graphql", gin.WrapH(gqlHandler))

    engine.GET("/health/live", liveHandler)
    engine.GET("/health/ready", readyHandler)

    return engine
}
```

Authentication middleware should usually be able to establish an optional actor because
the same GraphQL endpoint may contain both public and authenticated operations.

Field/use-case authorization still occurs deeper in the application.

---

## 11. Authentication flow

GraphQL uses a single primary HTTP endpoint, so route-level authorization is less useful
than in REST.

Recommended flow:

```text
Authorization header / cookie
        |
        v
Gin auth middleware
        |
        v
validate credential
        |
        v
actor stored in request context
        |
        v
gqlgen
        |
        v
resolver
        |
        v
service-level authorization
```

### Rule

Authentication answers:

```text
Who is calling?
```

Service authorization answers:

```text
May this actor perform this use case on this resource?
```

Do not rely only on a GraphQL directive such as:

```graphql
@auth
```

for resource-level authorization.

A directive may enforce broad requirements such as "must be logged in", while the
service remains responsible for domain authorization.

---

## 12. Context propagation

The Go request context is the request-scoped carrier.

Use it for:

- cancellation;
- deadlines;
- request ID;
- authenticated actor;
- tracing context;
- DataLoader registry.

Do not use context for arbitrary business parameters that should be explicit function
arguments.

Example actor accessor:

```go
type actorKey struct{}

func WithActor(ctx context.Context, actor Actor) context.Context {
    return context.WithValue(ctx, actorKey{}, actor)
}

func ActorFromContext(ctx context.Context) (Actor, bool) {
    actor, ok := ctx.Value(actorKey{}).(Actor)
    return actor, ok
}
```

Services should receive actor/resource identifiers explicitly where that improves
testability and business clarity.

---

## 13. Services

Services implement application use cases.

They own:

- business invariants;
- authorization decisions;
- orchestration across repositories;
- transaction boundaries;
- interaction with external services;
- domain/application validation;
- stable application errors.

Prefer service-owned input/output types rather than passing generated GraphQL models deep
into the business layer.

Example:

```go
type CreateInput struct {
    ActorID uint
    Title   string
}

type Task struct {
    ID        uint
    Title     string
    Status    string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Service interface {
    Create(ctx context.Context, input CreateInput) (*Task, error)
    GetByID(ctx context.Context, actorID, taskID uint) (*Task, error)
}
```

This preserves:

```text
GraphQL transport model != application model != persistence model
```

For a tiny feature, application DTOs may be lightweight, but the dependency direction
must remain the same.

---

## 14. Repository layer

GORM belongs in repository and database packages.

Good dependency direction:

```text
resolver
   |
   v
service
   |
   v
repository interface
   ^
   |
repository implementation
   |
   v
GORM
```

Avoid:

```text
resolver -> GORM
service  -> GORM
```

Example:

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

Repositories own:

- GORM queries;
- joins;
- preloads;
- persistence filtering;
- pagination queries;
- batch queries used by DataLoaders;
- constraint/error translation.

Repositories do not own:

- GraphQL schema behavior;
- GraphQL error codes;
- permission policy;
- workflow orchestration.

---

## 15. Persistence entities

Entities represent persisted/core database state.

Example:

```go
type Task struct {
    ID        uint      `gorm:"primaryKey"`
    OwnerID   uint      `gorm:"not null;index"`
    Title     string    `gorm:"not null"`
    Status    string    `gorm:"not null;index"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

Entities must not contain:

- gqlgen resolver logic;
- GraphQL directives;
- GraphQL descriptions;
- Gin bindings;
- HTTP status codes;
- transport-only presentation fields.

Do not automatically bind GraphQL types directly to GORM entities merely to reduce code.

That shortcut strongly couples the public GraphQL contract to database design.

---

## 16. Mapping between layers

Recommended mapping:

```text
GraphQL Input
    |
    v
Service Input
    |
    v
Entity / Repository
    |
    v
Service Result
    |
    v
GraphQL Model
```

Not every feature requires a separate struct at every arrow.

The architectural rule is about **ownership and dependency**, not mechanical DTO
proliferation.

Create a mapping when:

- the GraphQL field shape differs from persistence;
- IDs require conversion;
- GraphQL enum names differ from stored values;
- internal fields must not be exposed;
- derived/computed fields exist;
- mutation inputs do not match entity structure.

---

## 17. GraphQL error contract

GraphQL errors should use stable machine-readable error codes in `extensions`.

Example response:

```json
{
  "data": {
    "task": null
  },
  "errors": [
    {
      "message": "task was not found",
      "path": ["task"],
      "extensions": {
        "code": "NOT_FOUND",
        "request_id": "01J..."
      }
    }
  ]
}
```

Recommended application-to-GraphQL mapping:

| Application error | GraphQL `extensions.code` |
| --- | --- |
| validation | `BAD_USER_INPUT` |
| unauthenticated | `UNAUTHENTICATED` |
| forbidden | `FORBIDDEN` |
| not found | `NOT_FOUND` |
| conflict | `CONFLICT` |
| rate limited | `RATE_LIMITED` |
| dependency busy | `DEPENDENCY_BUSY` |
| unexpected | `INTERNAL_SERVER_ERROR` |

### Error rules

- application errors remain independent from GraphQL;
- map errors at the GraphQL boundary;
- never expose raw GORM errors;
- never expose raw SQLite errors;
- never expose stack traces;
- never expose filesystem paths or secrets;
- unexpected errors should return a safe public message;
- log unexpected errors once with request/operation context.

Use a centralized gqlgen error presenter.

Conceptually:

```go
srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
    return graphqlerror.Present(ctx, err)
})
```

Use a centralized recovery function for unexpected panics.

---

## 18. GraphQL HTTP status behavior

Do not recreate the REST response-envelope architecture inside GraphQL.

GraphQL responses naturally use:

```json
{
  "data": {},
  "errors": []
}
```

A successfully parsed/executed GraphQL HTTP request may contain GraphQL errors while the
HTTP layer itself remains successful.

Therefore:

```text
HTTP status
    -> transport/protocol condition

GraphQL errors[]
    -> operation/field/application condition
```

Do not force application errors into REST-style envelopes such as:

```json
{
  "data": null,
  "error": {}
}
```

inside GraphQL.

---

## 19. DataLoader and N+1 prevention

GraphQL makes it easy to create N+1 database queries.

Example schema:

```graphql
type Task {
  owner: User!
}
```

A naive owner resolver can execute one user query per task.

Avoid:

```text
tasks query:        1 query
owner field task 1: 1 query
owner field task 2: 1 query
owner field task 3: 1 query
...
```

Use request-scoped DataLoaders for relation fields where batching materially reduces
queries.

Desired behavior:

```text
tasks query:          1 query
users by owner IDs:   1 batched query
```

### DataLoader rules

1. DataLoaders are request-scoped.
2. Do not keep a global DataLoader cache across users/requests.
3. Batch repository methods should accept `context.Context`.
4. DataLoader keys must preserve requested ordering.
5. Authorization must not be bypassed by batching.
6. DataLoaders optimize retrieval; they do not own business rules.
7. Do not introduce a loader for every field automatically.

Example repository API:

```go
GetByIDs(ctx context.Context, ids []uint) ([]*entity.User, error)
```

---

## 20. Pagination

Never expose unbounded production list fields.

Avoid:

```graphql
type Query {
  tasks: [Task!]!
}
```

Prefer:

```graphql
type Query {
  tasks(first: Int = 20, after: String): TaskConnection!
}
```

Initial rules:

```text
default page size: 20
maximum page size: 100
```

The exact limits should be configurable or centrally defined.

Prefer cursor pagination for public GraphQL APIs where list stability matters.

Connection shape:

```graphql
type TaskConnection {
  nodes: [Task!]!
  pageInfo: PageInfo!
}

type PageInfo {
  hasNextPage: Boolean!
  endCursor: String
}
```

Cursor internals should be treated as opaque by clients.

---

## 21. Query complexity and abuse controls

A GraphQL endpoint should not allow arbitrarily expensive operations.

Production controls should include:

- maximum request body size;
- GraphQL operation complexity limit;
- pagination limits;
- authentication-aware rate limits;
- timeouts/cancellation;
- bounded resolver work;
- DataLoader batching;
- limits on expensive search/filter operations.

Example conceptual server setup:

```go
srv := handler.New(
    generated.NewExecutableSchema(
        generated.Config{
            Resolvers: resolver,
        },
    ),
)

srv.Use(extension.Introspection{})
srv.Use(extension.FixedComplexityLimit(250))
```

The complexity value must be tuned from actual schema/workload behavior rather than
treated as universal.

For production, introspection policy should be an explicit product/security decision.

---

## 22. GraphQL directives

Directives are appropriate for cross-cutting GraphQL behavior.

Examples:

```graphql
directive @auth on FIELD_DEFINITION
directive @hasRole(role: Role!) on FIELD_DEFINITION
```

Potential uses:

- broad authentication requirements;
- broad role requirements;
- field-level metadata;
- declarative transport concerns.

Do not place complicated domain authorization into directives.

Avoid:

```text
directive
    -> query database
    -> evaluate complete resource ownership policy
    -> mutate domain state
```

Prefer:

```text
directive
    -> broad access guard

service
    -> complete domain authorization
```

---

## 23. Custom scalars

Use custom scalars only when they improve the API contract.

Common examples:

```graphql
scalar DateTime
scalar UUID
```

Scalar code owns transport parsing/serialization.

It must not contain database logic.

Example DateTime mapping responsibility:

```text
GraphQL string
   <->
time.Time
```

Keep custom scalar implementations in:

```text
internal/graphql/scalar
```

---

## 24. Transactions

The service layer owns transaction boundaries because it understands the complete use
case.

Do not start transactions in resolvers.

Preferred abstraction:

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

This keeps GORM transaction objects out of resolver and service APIs.

---

## 25. SQLite strategy

SQLite remains the primary database.

Recommended development path:

```text
./data/app.db
```

Recommended production mounted path:

```text
/data/app.db
```

### WAL

Enable:

```sql
PRAGMA journal_mode=WAL;
```

### Foreign keys

Enable:

```sql
PRAGMA foreign_keys=ON;
```

### Busy timeout

Start with a bounded configurable value, for example:

```text
5 seconds
```

### Connection pool

Use conservative values appropriate for SQLite:

```text
MaxOpenConns: small
MaxIdleConns: small
```

Measure before increasing write concurrency.

The existing deployment constraint remains:

```text
prefer one application instance
    +
persistent local/attached storage
```

Do not treat a shared SQLite file as a horizontally scaled multi-writer database.

---

## 26. GORM rules

GORM remains confined to database/repository code.

Every repository operation should propagate context:

```go
db.WithContext(ctx)
```

Persistence errors should be translated into stable application errors:

```text
record not found
    -> ErrNotFound

unique constraint
    -> ErrConflict

busy/locked
    -> ErrDependencyBusy

unexpected DB failure
    -> wrapped internal error
```

Raw GORM/SQLite errors must never reach GraphQL clients.

---

## 27. Goose migrations

Goose remains the official database migration system.

```text
GORM models
    -> runtime object mapping and queries

Goose SQL migrations
    -> schema history and deployment changes
```

Do not use GORM `AutoMigrate` as the production migration mechanism.

Migration layout:

```text
internal/migrations/
└── sql/
    ├── 00001_init.sql
    ├── 00002_add_task_status.sql
    └── 00003_add_task_indexes.sql
```

Typical Makefile targets:

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

Deployment:

```text
build
  -> backup when required
  -> run Goose once
  -> start/restart app
  -> readiness succeeds
```

---

## 28. Uber Fx module graph

Recommended top-level composition:

```go
var Module = fx.Module(
    "app",
    config.Module,
    observability.Module,
    database.Module,
    repository.Module,
    service.Module,
    graphql.Module,
    middleware.Module,
    router.Module,
    server.Module,
)
```

GraphQL module:

```text
graphql.Module
│
├── resolver.Module
├── dataloader.Module
├── directive providers
├── GraphQL executable schema
└── GraphQL handler
```

Application graph:

```text
Config
  ├──────────────> Zap
  ├──────────────> GORM / SQLite
  └──────────────> HTTP Server

GORM / SQLite
      |
      v
Repositories
      |
      v
Services
      |
      v
Resolver root
      |
      v
gqlgen executable schema
      |
      v
GraphQL handler
      |
      v
Gin router
      |
      v
HTTP Server
```

---

## 29. Example Fx constructors

Resolver module:

```go
var Module = fx.Module(
    "graphql-resolver",
    fx.Provide(New),
)
```

GraphQL handler:

```go
func NewHandler(
    resolver *resolver.Resolver,
) http.Handler {
    schema := generated.NewExecutableSchema(
        generated.Config{
            Resolvers: resolver,
        },
    )

    srv := handler.NewDefaultServer(schema)

    return srv
}
```

Router receives the already-constructed handler:

```go
func NewRouter(
    gql http.Handler,
    logger *zap.Logger,
) *gin.Engine {
    engine := gin.New()

    // middleware...

    engine.POST("/graphql", gin.WrapH(gql))

    return engine
}
```

Constructors should remain ordinary Go functions that can be called without Fx in tests.

---

## 30. HTTP server lifecycle

`internal/server` continues to own `http.Server`.

Conceptual constructor:

```go
func New(
    cfg config.Config,
    router *gin.Engine,
) *http.Server {
    return &http.Server{
        Addr:         cfg.HTTP.Address,
        Handler:      router,
        ReadTimeout:  cfg.HTTP.ReadTimeout,
        WriteTimeout: cfg.HTTP.WriteTimeout,
        IdleTimeout:  cfg.HTTP.IdleTimeout,
    }
}
```

Use Fx lifecycle hooks for start/stop.

On shutdown:

1. stop accepting new HTTP requests;
2. allow in-flight GraphQL operations to finish within timeout;
3. cancel remaining request contexts;
4. stop background workers;
5. flush telemetry;
6. sync Zap;
7. close the database.

---

## 31. Middleware order

Recommended Gin middleware order:

```text
request ID
-> recovery
-> security headers
-> access logging
-> tracing/metrics when introduced
-> body-size limit
-> CORS
-> authentication extraction
-> DataLoader injection when implemented at HTTP boundary
-> gqlgen handler
```

Prefer DataLoader setup that is clearly request scoped.

Useful access-log fields:

```text
request_id
method
route
status
duration
actor_id
graphql_operation_name
graphql_operation_type
error_code
```

Do not log full GraphQL variables by default because they may contain secrets or
personal data.

---

## 32. GraphQL operation logging

HTTP logging alone sees mostly:

```text
POST /graphql
```

That is not enough operational context.

Where practical, enrich request context/logging with:

```text
operation_name
operation_type
```

Example:

```text
graphql.operation_name = CreateTask
graphql.operation_type = mutation
```

Avoid high-cardinality labels based on entire query documents.

Do not log full query documents in production by default.

---

## 33. Security baseline

When authentication is enabled:

- validate credential signature, issuer, audience, expiry, and algorithm where applicable;
- use secure cookie flags for cookie-based authentication;
- use CSRF protections when cookie authentication makes them necessary;
- restrict CORS to known origins;
- configure Gin trusted proxies explicitly;
- enforce request body limits;
- rate-limit authentication and expensive operations;
- enforce GraphQL query complexity limits;
- enforce pagination bounds;
- avoid logging GraphQL variables by default;
- prevent secrets and personal data from entering logs;
- use service-level authorization;
- keep dependencies and the Go toolchain patched.

Do not assume GraphQL automatically protects against expensive queries.

---

## 34. Health endpoints stay outside GraphQL

Keep health probes as ordinary HTTP endpoints.

```text
GET /health/live
GET /health/ready
```

Do not model Kubernetes/container health as:

```graphql
query {
  health
}
```

Liveness answers:

```text
Is the process alive?
```

Readiness answers:

```text
Can this process currently serve requests?
```

Readiness may verify:

- database initialization;
- SQLite connectivity;
- required startup state.

---

## 35. GraphQL playground

Development may expose a GraphQL playground on:

```text
GET /graphql
```

or:

```text
GET /playground
```

Production exposure should be explicitly configured.

Conceptual:

```go
if cfg.Environment == "development" {
    engine.GET("/graphql", gin.WrapH(playground.Handler(
        "GraphQL Playground",
        "/graphql",
    )))
}
```

Do not accidentally couple production API availability to playground availability.

---

## 36. Testing strategy

Use the narrowest test that proves behavior.

| Layer | Test style | Main assertions |
| --- | --- | --- |
| service | unit with fakes/mocks | business rules, authorization, error paths |
| repository | integration with SQLite | queries, constraints, batch methods, transactions |
| resolver | unit/integration | input mapping, service call, output mapping |
| GraphQL operation | HTTP/GraphQL integration | schema execution, errors, auth, nullability |
| DataLoader | unit/integration | batching, ordering, request isolation |
| middleware | `httptest` | identity/context/headers/rejection |
| migrations | SQLite integration | clean install and upgrade path |
| app | small end-to-end suite | critical GraphQL journeys |

### Service tests

Service tests should not require:

- Fx;
- Gin;
- gqlgen;
- GORM;
- SQLite.

Example:

```go
repo := &FakeTaskRepository{}
svc := taskservice.New(repo, zap.NewNop())
```

### Resolver tests

Resolvers may be tested with a fake service:

```go
resolver := resolver.New(fakeTaskService, zap.NewNop())
```

Test transport mapping without a real DB.

### GraphQL integration tests

Use the real executable schema for important operation-level behavior.

Example targets:

```graphql
mutation CreateTask($input: CreateTaskInput!) {
  createTask(input: $input) {
    id
    title
  }
}
```

Assert:

- returned `data`;
- `errors`;
- `extensions.code`;
- authorization;
- nullability;
- pagination behavior.

### Repository tests

Use real SQLite.

Batch methods used by DataLoaders require integration tests.

---

## 37. GraphQL schema tests

Schema changes are API changes.

CI should catch unintended generated/schema drift.

Recommended checks:

```text
go test ./...
go test -race ./...
go vet ./...
gqlgen generate
git diff --exit-code
```

As the project matures, consider schema compatibility checks for breaking changes.

At minimum, review changes that:

- remove fields;
- change nullable to non-null in unsafe ways;
- change argument types;
- remove enum values;
- change input requirements;
- alter pagination conventions.

---

## 38. Build and deployment baseline

CI should eventually run:

```text
go mod verify
gqlgen generate
generated-code drift check
go vet ./...
static analysis / lint
go test -race ./...
repository integration tests
GraphQL integration tests
Goose migration validation
go build ./cmd/api
vulnerability scan
container build
container scan
```

Runtime requirements remain:

- non-root user where possible;
- writable mounted SQLite directory;
- stdout/stderr structured logging;
- explicit health checks;
- graceful termination;
- production request limits.

---

## 39. Example Task feature flow

GraphQL operation:

```graphql
mutation CreateTask($input: CreateTaskInput!) {
  createTask(input: $input) {
    id
    title
    status
  }
}
```

Execution:

```text
POST /graphql
     |
     v
Gin request ID
     |
     v
Gin auth extraction
     |
     v
gqlgen
     |
     v
Mutation.createTask resolver
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
Gin
  establish HTTP/request context
  broad authentication extraction
  request limits

gqlgen
  parse/validate GraphQL document
  execute operation
  resolve fields

Resolver
  translate GraphQL input
  get actor/request context
  call service
  translate output

Service
  business validation
  authorization
  orchestration
  transaction decision

Repository
  execute GORM operations
  translate persistence errors

GraphQL error presenter
  map application errors
  attach stable extensions.code
```

---

## 40. Example Task relation flow with DataLoader

Query:

```graphql
query Tasks {
  tasks(first: 20) {
    nodes {
      id
      title
      owner {
        id
        name
      }
    }
  }
}
```

Desired flow:

```text
Query.tasks
   |
   v
TaskService.List
   |
   v
TaskRepository.List
   |
   v
20 tasks
   |
   +----------------------------+
                                |
                                v
                         Task.owner resolver
                                |
                                v
                         User DataLoader
                                |
                                v
                    UserRepository.GetByIDs
                                |
                                v
                          one batch query
```

Avoid one owner query per task.

---

## 41. Implementation order

Build one vertical GraphQL slice before adding optional infrastructure.

### Phase 1 — application shell

1. initialize Go module;
2. create `cmd/api`;
3. add typed configuration;
4. add Uber Fx;
5. add Zap;
6. add Gin;
7. add `http.Server`;
8. add graceful lifecycle.

### Phase 2 — GraphQL shell

1. add gqlgen;
2. create `gqlgen.yml`;
3. create `internal/graphql/schema`;
4. create root `Query`;
5. create root `Mutation`;
6. generate gqlgen code;
7. construct gqlgen handler through Fx;
8. mount `POST /graphql` in Gin;
9. add dev-only playground.

At this point:

```graphql
query {
  ping
}
```

should work end-to-end.

### Phase 3 — persistence

1. add GORM;
2. add SQLite driver;
3. configure SQLite;
4. add transaction manager;
5. add Goose;
6. create first migration;
7. add migration Makefile commands;
8. add DB readiness check.

### Phase 4 — first GraphQL vertical slice

Implement:

```text
Task GraphQL schema
-> Task resolver
-> Task service
-> Task repository
-> GORM entity
-> SQLite migration
```

Suggested API:

```graphql
type Query {
  task(id: ID!): Task
  tasks(first: Int = 20, after: String): TaskConnection!
}

type Mutation {
  createTask(input: CreateTaskInput!): Task!
  updateTask(id: ID!, input: UpdateTaskInput!): Task!
  deleteTask(id: ID!): Boolean!
}
```

### Phase 5 — GraphQL production baseline

Add:

- centralized error presenter;
- panic recovery;
- request IDs;
- operation logging;
- pagination limits;
- query complexity limit;
- request body limit;
- security headers;
- CORS;
- GraphQL integration tests.

### Phase 6 — authentication

Add:

- identity model;
- Gin authentication extraction;
- actor request context;
- optional broad `@auth` directive if useful;
- service-level authorization;
- authentication/security tests.

### Phase 7 — DataLoaders

Add DataLoaders only for relations that exhibit N+1 behavior.

Start with actual measured/query-count evidence.

### Phase 8 — operational maturity

Add as required:

- metrics;
- OpenTelemetry;
- tracing;
- persistent query strategy;
- caching;
- background queues;
- subscriptions;
- external integrations.

---

## 42. Definition of done for a new GraphQL operation

A new query or mutation is complete when:

- its schema is explicit;
- nullable/non-null behavior is intentional;
- input types are explicit;
- list fields are bounded/paginated;
- generated gqlgen code is current;
- resolver contains only transport mapping;
- service contains business behavior;
- service accepts `context.Context`;
- authorization is enforced where required;
- persistence is behind repository interfaces;
- repository uses `WithContext`;
- persistence errors are translated;
- application errors map to stable GraphQL error codes;
- raw GORM/SQLite errors cannot reach clients;
- logging contains request/operation context;
- expected success/error paths are tested;
- required migration is included;
- N+1 behavior has been considered;
- expensive-operation behavior is bounded.

---

## 43. Architectural anti-patterns

### Fat resolver

Avoid:

```go
func (r *mutationResolver) CreateTask(
    ctx context.Context,
    input model.CreateTaskInput,
) (*model.Task, error) {
    // inspect auth
    // query GORM
    // enforce permissions
    // write multiple tables
    // call external API
    // construct response
}
```

Move business orchestration into a service.

---

### Resolver calling GORM

Avoid:

```text
resolver -> *gorm.DB
```

Use:

```text
resolver -> service -> repository
```

---

### GraphQL models used as persistence entities

Avoid:

```go
func (r *Repository) Create(ctx context.Context, task *model.Task) error
```

Prefer persistence/domain-owned types.

---

### GraphQL types leaking into services

Avoid:

```go
func (s *Service) Create(
    ctx context.Context,
    input model.CreateTaskInput,
) (*model.Task, error)
```

when `model` is gqlgen-generated transport code.

Prefer:

```go
func (s *Service) Create(
    ctx context.Context,
    input task.CreateInput,
) (*task.Task, error)
```

---

### `gin.Context` inside resolvers/services

Avoid:

```go
func (s *Service) Create(ctx *gin.Context, ...)
```

Use:

```go
func (s *Service) Create(ctx context.Context, ...)
```

---

### Repository from DataLoader bypasses policy

Avoid using DataLoaders as a shortcut around authorization.

DataLoader results must still be safe for the requesting actor/use case.

---

### Unbounded lists

Avoid:

```graphql
tasks: [Task!]!
```

for potentially large tables.

Use pagination.

---

### Business errors encoded as ad-hoc strings

Avoid clients depending on:

```text
"task not found"
```

Use stable:

```json
{
  "extensions": {
    "code": "NOT_FOUND"
  }
}
```

---

### REST envelope inside GraphQL

Avoid returning custom wrappers like:

```graphql
type CreateTaskPayload {
  data: Task
  error: APIError
}
```

for every operation solely to imitate REST.

Use GraphQL `data` + `errors` unless the domain genuinely requires a payload object.

---

### GraphQL directives as the whole authorization layer

Avoid treating `@auth` as sufficient domain authorization.

The service remains authoritative.

---

### Global DataLoader cache

Avoid request-independent DataLoader caches unless explicitly designed with safe cache
semantics.

Default DataLoaders are request-scoped.

---

### Automatic schema migration

Do not run unconstrained GORM `AutoMigrate` in production startup.

Use Goose.

---

## 44. What remains from the previous REST-oriented architecture

The following decisions remain unchanged from the existing architecture:

- small `cmd/api/main.go`;
- Uber Fx composition;
- typed startup configuration;
- Zap logging;
- GORM repository boundary;
- SQLite;
- Goose migrations;
- service-owned business logic;
- service-owned transactions;
- persistence entities separated from public API types;
- health/readiness endpoints;
- graceful shutdown;
- repository integration tests;
- production database backup rules.

The existing document already established Gin, Fx, GORM, SQLite, Goose, Zap, layered
boundaries, and `internal/` as the application-code root.

The primary transport change is:

```text
OLD

router
  -> middleware
  -> controller
  -> service
  -> repository
  -> GORM
  -> SQLite


NEW

Gin
  -> middleware
  -> gqlgen
  -> resolver
  -> service
  -> repository
  -> GORM
  -> SQLite
```

---

## 45. Patterns retained from Apache Answer

Retain:

- small executable entry point;
- centralized server construction;
- service/repository boundaries;
- persistence model separation;
- constructor-based dependency injection;
- versioned migrations;
- `internal/` application code;
- `pkg/` only for intentionally reusable code.

Adapt:

```text
Apache Answer-inspired concept      GraphQL-first project
---------------------------------------------------------------
HTTP controller boundary       ->   gqlgen resolver boundary
REST route-per-use-case        ->   GraphQL schema/operation
request/response DTOs          ->   GraphQL inputs/models + service DTOs
manual/Wire-style composition  ->   Uber Fx
persistence stack              ->   GORM
server SQL DB patterns         ->   SQLite
migration mechanism            ->   Goose
logging                        ->   Zap
```

---

## 46. Core rules to remember

For almost every GraphQL feature:

```text
schema
  -> resolver
  -> service
  -> repository interface
  -> GORM repository implementation
  -> SQLite
```

Application construction:

```text
Uber Fx
  -> config
  -> Zap
  -> GORM/SQLite
  -> repositories
  -> services
  -> GraphQL resolvers
  -> gqlgen executable schema
  -> gqlgen handler
  -> Gin router
  -> HTTP server
```

For related-object fields:

```text
field resolver
  -> request-scoped DataLoader
  -> repository batch method
```

For authorization:

```text
Gin middleware/directive
  -> broad authentication/access requirement

Service
  -> authoritative resource/domain authorization
```

The most important boundary is:

> **GraphQL is the transport contract. Services are the application contract. Repositories are the persistence contract.**

If a future design decision mixes those responsibilities, it should have a concrete
reason rather than being done for convenience.
