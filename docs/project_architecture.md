# Production-Ready Go/Chi GraphQL-First Project Architecture

Architecture for production Go API. Replaces original Gin HTTP layer with **Chi + standard `net/http`**, keeps Uber Fx + GORM + SQLite + Goose + Zap. **GraphQL primary application API**.

Inspired by Apache Answer's package boundaries and layering, but swaps REST/controller transport for **schema-first GraphQL transport using gqlgen**.

---

## 1. Final technology decisions

| Concern | Decision |
| --- | --- |
| Language | Go |
| HTTP router | Chi (`github.com/go-chi/chi/v5`) |
| HTTP abstraction | Standard `net/http` (`http.Handler`, `http.HandlerFunc`, `http.Server`) |
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
| Health endpoints | Standard `net/http` handlers routed by Chi |
| GraphQL subscriptions | Not enabled initially |

Stack:

```text
Client
  |
  v
net/http + Chi
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

### Why Chi is the HTTP boundary

Chi = thin router over stdlib. App treats `http.Handler` as HTTP contract, Chi as routing/composition impl.

Gives three wins:

1. gqlgen mounts direct — server already `http.Handler`;
2. middleware stays ordinary `func(http.Handler) http.Handler`;
3. server, router, GraphQL handler, middleware testable solo with `httptest`.

No app-wide wrapper around `chi.Router`. Business code shouldn't know which router used.

---

## 2. Architecture goals

Should make easy:

- GraphQL as primary public API contract;
- trace full path of GraphQL operation;
- business logic independent of Chi, gqlgen, GORM, SQLite;
- build app through explicit Uber Fx modules;
- keep GraphQL schema/types separate from persistence entities;
- test services w/o starting Chi, gqlgen, Fx, DB;
- test repositories against real SQLite;
- avoid GraphQL N+1;
- consistent auth across GraphQL ops;
- stable GraphQL error codes, no internal-error leaks;
- limit expensive GraphQL queries;
- swap SQLite later w/o business-logic rewrite;
- add jobs/caching/tracing/subscriptions later w/o breaking boundaries;
- clean predictable shutdown.

Stay small. Add GraphQL infra only when explicit responsibility exists.

---

## 3. GraphQL-first core dependency rule

Dependencies point toward business logic.

```text
GraphQL Client
      |
      v
Chi + net/http Middleware
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

Normal request path:

```text
POST /graphql
    -> http.Server
    -> Chi middleware chain
    -> gqlgen operation execution
    -> resolver
    -> service
    -> repository
    -> GORM
    -> SQLite
```

Nested relation fields:

```text
GraphQL field resolver
    -> request-scoped DataLoader
    -> repository/service
    -> GORM
    -> SQLite
```

### Dependency rules

1. Resolvers may depend on services.
2. Resolvers use `context.Context`; no dependency on Chi router types or `*http.Request`.
3. Services use `context.Context`, never depend on gqlgen, Chi, HTTP transport types.
4. Services own business rules, authz decisions, orchestration, transaction boundaries.
5. Repositories own persistence queries + GORM-specific behavior.
6. Services don't import GORM.
7. Persistence entities don't define GraphQL contract.
8. GraphQL generated types must not become persistence model by default.
9. Lower layers must not import Chi, gqlgen handlers, GraphQL resolvers, `net/http`, or transport code.
10. Fx = composition/lifecycle concern, not business dependency.
11. Zap = single app logger.
12. Resource authz must not live only in GraphQL directives or middleware.
13. DataLoader batching = transport optimization, not place for business rules.
14. Router-specific values must not become service args; convert request metadata to typed context values at HTTP boundary.
15. Router constructor returns `http.Handler` unless caller genuinely needs concrete `chi.Router`.

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
│   │   ├── server.go
│   │   ├── lifecycle.go
│   │   └── module.go
│   │
│   ├── router/
│   │   ├── router.go
│   │   ├── health.go
│   │   └── module.go
│   │
│   ├── middleware/
│   │   ├── request_id.go
│   │   ├── recovery.go
│   │   ├── access_log.go
│   │   ├── security_headers.go
│   │   ├── body_limit.go
│   │   ├── client_ip.go
│   │   ├── auth.go
│   │   └── module.go
│   │
│   ├── requestctx/
│   │   ├── actor.go
│   │   ├── request_id.go
│   │   └── client_ip.go
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

`internal/router` intentionally small. Composes routes + route-scoped middleware; no business handlers or app rules.

`internal/server` owns concrete `http.Server`, lifecycle hooks, listening, graceful shutdown. Separate from router → both easier to test, keeps network lifecycle out of route construction.

`internal/middleware` = HTTP-only cross-cutting behavior, standard middleware signature:

```go
type Middleware func(http.Handler) http.Handler
```

`internal/requestctx` = typed request-context accessors shared by HTTP middleware + GraphQL resolvers. Must not import Chi.

`internal/graphql` owns GraphQL transport: SDL schema files, gqlgen-generated exec code, GraphQL API models, resolvers, directives, custom scalars, DataLoader integration, GraphQL error presentation, gqlgen server construction. No app business logic.

`internal/service` = use-case layer. `internal/repository` = persistence boundary. `internal/entity` = GORM persistence model layer.

Skip generic dumping-ground packages (`utils`, `helpers`, `common`). Build packages around concrete responsibility.

---

## 5. Why GraphQL is not placed in `controller/`

REST architecture used:

```text
controller -> service -> repository
```

GraphQL-first architecture uses:

```text
resolver -> service -> repository
```

Resolver plays transport role REST controller played before.

So default project should **remove `internal/controller` for primary API**.

Don't do this:

```text
GraphQL resolver
    -> controller
    -> service
```

Unnecessary transport-to-transport layer.

Use:

```text
GraphQL resolver
    -> service
```

REST controllers/handlers may still exist for transport-specific endpoints genuinely not GraphQL concerns:

```text
GET /health/live
GET /health/ready
```

---

## 6. GraphQL schema is the API contract

GraphQL schema = primary public API contract.

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

1. Organize by domain/feature, not one huge schema file.
2. Prefer explicit input types for mutations.
3. Don't accept persistence entities as GraphQL inputs.
4. Don't expose internal DB fields just cuz they exist.
5. Keep names stable once clients consume them.
6. Deprecate, don't sudden-remove fields.
7. Don't use GraphQL types to bypass service-level validation.
8. Keep pagination conventions consistent across list fields.
9. Don't expose unbounded list fields.
10. Document nullable vs non-null semantics deliberately.

---

## 7. `gqlgen.yml`

Keep gqlgen config at repo root.

Conceptual config:

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

Treat generated files as generated code.

Don't hand-edit generated exec code.

Recommended commands:

```makefile
graphql-generate:
	go run github.com/99designs/gqlgen generate

graphql-check:
	go run github.com/99designs/gqlgen generate
	git diff --exit-code
```

CI should ensure generated GraphQL code current.

---

## 8. Resolver responsibilities

Resolvers = thin GraphQL adapters.

Should normally:

1. accept gqlgen-provided `context.Context`;
2. read already-established actor identity from context;
3. map GraphQL input into service input;
4. call one application service/use case;
5. map service output into GraphQL output when required;
6. return stable app errors for centralized GraphQL error presentation.

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
- depend directly on `*http.Request` or Chi router details;
- return raw GORM errors.

---

## 9. Resolver root

Use one resolver root object to hold service deps needed by generated resolvers.

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

No mutable request state in resolver root. Request-specific state belongs in `context.Context`.

---

## 10. Chi's role in a GraphQL-first application

Chi = HTTP routing + middleware-composition layer. Keep thin.

Chi owns/composes:

- route matching;
- standard `net/http` middleware;
- request ID creation/propagation;
- panic recovery at HTTP boundary;
- security headers;
- CORS;
- trusted client-IP extraction when required;
- broad auth extraction;
- request-body limits;
- access logging;
- health routes;
- mounting gqlgen handler.

Chi does **not** own GraphQL field resolution, business validation, resource authz, transactions, or persistence.

Recommended routes:

```text
POST /graphql       -> GraphQL operations
GET  /playground    -> development-only GraphQL playground
GET  /health/live
GET  /health/ready
```

Prefer separate `/playground` route over overloading `GET /graphql` — keeps prod GraphQL endpoint contract simple, dev UI easy to disable.

### Router constructor

Router should expose stdlib abstraction:

```go
func NewRouter(
    gqlHandler http.Handler,
    authMiddleware *middleware.Auth,
    cfg config.Config,
) http.Handler {
    r := chi.NewRouter()

    // Global middleware. All r.Use calls happen before routes are registered.
    r.Use(chimiddleware.RequestID)
    r.Use(middleware.Recover())
    r.Use(middleware.SecurityHeaders())
    r.Use(middleware.AccessLog())
    r.Use(middleware.BodyLimit(cfg.HTTP.MaxBodyBytes))
    r.Use(middleware.CORS(cfg.CORS))
    r.Use(authMiddleware.Optional())

    r.Method(http.MethodPost, "/graphql", gqlHandler)

    r.Get("/health/live", liveHandler)
    r.Get("/health/ready", readyHandler)

    if cfg.Environment == "development" {
        r.Handle("/playground", playground.Handler(
            "GraphQL Playground",
            "/graphql",
        ))
    }

    return r
}
```

Important boundary:

```text
router implementation: chi.Router
public constructor result: http.Handler
server dependency:      http.Handler
```

Don't pass `chi.Router` into services or repositories.

### Route grouping

Use `Route`, `Group`, `With` only when middleware genuinely applies to route subset. Keep main GraphQL endpoint free of route-level authz — one endpoint may run both public and authenticated ops.

Example for future non-GraphQL admin endpoints:

```go
r.Route("/internal", func(r chi.Router) {
    r.Use(adminOnly)
    r.Get("/metrics", metricsHandler)
})
```

### Standard-library compatibility rule

App-owned middleware should use standard `net/http` signatures even when mounted via Chi:

```go
func SecurityHeaders() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // set headers
            next.ServeHTTP(w, r)
        })
    }
}
```

Minimizes router lock-in, allows direct middleware tests with `httptest`.

---

## 11. Authentication flow

GraphQL uses single primary HTTP endpoint — route-level authz less useful than REST.

Recommended flow:

```text
Authorization header / cookie
        |
        v
net/http authentication middleware
        |
        v
validate credential
        |
        v
actor stored in request context
        |
        v
Chi
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

Auth middleware should normally establish **optional actor**, not reject every anonymous `/graphql` request. Public query + authenticated mutation share same endpoint.

Authentication answers:

```text
Who is calling?
```

Service authz answers:

```text
May this actor perform this use case on this resource?
```

Don't rely only on GraphQL directive like `@auth` for resource-level authz. Directive enforces broad reqs ("must be logged in"); service owns domain authz.

Keep context key + accessor impl outside router package — resolvers shouldn't need Chi import.

---

## 12. Context propagation

Go request context = request-scoped carrier.

Use for:

- cancellation;
- deadlines;
- request ID;
- authenticated actor;
- tracing context;
- DataLoader registry.

Don't stuff arbitrary business params into context that should be explicit func args.

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

Services should receive actor/resource IDs explicitly where it improves testability + clarity.

---

## 13. Services

Services implement app use cases.

Own:

- business invariants;
- authz decisions;
- orchestration across repositories;
- transaction boundaries;
- interaction w/ external services;
- domain/app validation;
- stable app errors.

Prefer service-owned input/output types over passing generated GraphQL models deep into business layer.

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

Preserves:

```text
GraphQL transport model != application model != persistence model
```

Tiny feature → app DTOs may be lightweight, but dependency direction stays same.

---

## 14. Repository layer

GORM belongs in repository + database packages.

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

Repositories don't own:

- GraphQL schema behavior;
- GraphQL error codes;
- permission policy;
- workflow orchestration.

---

## 15. Persistence entities

Entities = persisted/core DB state.

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
- router-specific bindings;
- HTTP status codes;
- transport-only presentation fields.

Don't auto-bind GraphQL types direct to GORM entities just to save code.

That shortcut tightly couples public GraphQL contract to DB design.

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

Not every feature needs separate struct at every arrow.

Architectural rule about **ownership and dependency**, not mechanical DTO proliferation.

Create mapping when:

- GraphQL field shape differs from persistence;
- IDs need conversion;
- GraphQL enum names differ from stored values;
- internal fields must not be exposed;
- derived/computed fields exist;
- mutation inputs don't match entity structure.

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

Recommended app-to-GraphQL mapping:

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

- app errors stay independent from GraphQL;
- map errors at GraphQL boundary;
- never expose raw GORM errors;
- never expose raw SQLite errors;
- never expose stack traces;
- never expose filesystem paths or secrets;
- unexpected errors return safe public message;
- log unexpected errors once w/ request/operation context.

Use centralized gqlgen error presenter.

Conceptually:

```go
srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
    return graphqlerror.Present(ctx, err)
})
```

Use centralized recovery func for unexpected panics.

---

## 18. GraphQL HTTP status behavior

Don't recreate REST response-envelope architecture inside GraphQL.

GraphQL responses naturally use:

```json
{
  "data": {},
  "errors": []
}
```

Successfully parsed/executed GraphQL HTTP request may contain GraphQL errors while HTTP layer stays successful.

So:

```text
HTTP status
    -> transport/protocol condition

GraphQL errors[]
    -> operation/field/application condition
```

Don't force app errors into REST-style envelopes like:

```json
{
  "data": null,
  "error": {}
}
```

inside GraphQL.

---

## 19. DataLoader and N+1 prevention

GraphQL makes N+1 DB queries easy to create.

Example schema:

```graphql
type Task {
  owner: User!
}
```

Naive owner resolver can fire one user query per task.

Avoid:

```text
tasks query:        1 query
owner field task 1: 1 query
owner field task 2: 1 query
owner field task 3: 1 query
...
```

Use request-scoped DataLoaders for relation fields where batching materially cuts queries.

Desired behavior:

```text
tasks query:          1 query
users by owner IDs:   1 batched query
```

### DataLoader rules

1. DataLoaders are request-scoped.
2. No global DataLoader cache across users/requests.
3. Batch repo methods should accept `context.Context`.
4. DataLoader keys must preserve requested ordering.
5. Authz must not be bypassed by batching.
6. DataLoaders optimize retrieval; don't own business rules.
7. Don't add loader for every field automatically.

Example repository API:

```go
GetByIDs(ctx context.Context, ids []uint) ([]*entity.User, error)
```

---

## 20. Pagination

Never expose unbounded prod list fields.

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

Exact limits should be configurable or centrally defined.

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

Treat cursor internals as opaque to clients.

---

## 21. Query complexity and abuse controls

GraphQL endpoint shouldn't allow arbitrarily expensive ops.

Prod controls should include:

- max request body size;
- GraphQL operation complexity limit;
- pagination limits;
- auth-aware rate limits;
- timeouts/cancellation;
- bounded resolver work;
- DataLoader batching;
- limits on expensive search/filter ops.

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

Tune complexity value from actual schema/workload behavior, not treat as universal.

Prod introspection policy should be explicit product/security decision.

---

## 22. GraphQL directives

Directives fit cross-cutting GraphQL behavior.

Examples:

```graphql
directive @auth on FIELD_DEFINITION
directive @hasRole(role: Role!) on FIELD_DEFINITION
```

Potential uses:

- broad auth requirements;
- broad role requirements;
- field-level metadata;
- declarative transport concerns.

Don't put complicated domain authz into directives.

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

Use custom scalars only when they improve API contract.

Common examples:

```graphql
scalar DateTime
scalar UUID
```

Scalar code owns transport parsing/serialization. No DB logic.

Example DateTime mapping responsibility:

```text
GraphQL string
   <->
time.Time
```

Keep custom scalar impls in:

```text
internal/graphql/scalar
```

---

## 24. Transactions

Service layer owns transaction boundaries — it knows full use case.

Don't start transactions in resolvers.

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

Keeps GORM transaction objects out of resolver/service APIs.

---

## 25. SQLite strategy

SQLite = primary DB.

Recommended dev path:

```text
./data/app.db
```

Recommended prod mounted path:

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

Start w/ bounded configurable value, e.g.:

```text
5 seconds
```

### Connection pool

Use conservative values for SQLite:

```text
MaxOpenConns: small
MaxIdleConns: small
```

Measure before raising write concurrency.

Existing deployment constraint remains:

```text
prefer one application instance
    +
persistent local/attached storage
```

Don't treat shared SQLite file as horizontally-scaled multi-writer DB.

---

## 26. GORM rules

GORM stays confined to database/repository code.

Every repo op should propagate context:

```go
db.WithContext(ctx)
```

Translate persistence errors into stable app errors:

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

Goose = official DB migration system.

```text
GORM models
    -> runtime object mapping and queries

Goose SQL migrations
    -> schema history and deployment changes
```

Don't use GORM `AutoMigrate` as prod migration mechanism.

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
└── GraphQL handler (http.Handler)
```

Application graph:

```text
Config
  ├──────────────> Zap
  ├──────────────> GORM / SQLite
  ├──────────────> HTTP middleware configuration
  └──────────────> http.Server

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
gqlgen handler (http.Handler)
      |
      v
Chi router (http.Handler)
      |
      v
http.Server
```

Fx builds the graph; packages stay usable w/o Fx. Constructors = ordinary Go funcs, lifecycle hooks limited to resources that actually start/stop work.

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
        generated.Config{Resolvers: resolver},
    )

    srv := handler.NewDefaultServer(schema)
    return srv
}
```

Router receives already-constructed gqlgen handler, returns standard abstraction:

```go
func NewRouter(
    gql http.Handler,
    auth *middleware.Auth,
    cfg config.Config,
) http.Handler {
    r := chi.NewRouter()

    r.Use(chimiddleware.RequestID)
    r.Use(middleware.Recover())
    r.Use(middleware.AccessLog())
    r.Use(auth.Optional())

    r.Method(http.MethodPost, "/graphql", gql)
    r.Get("/health/live", liveHandler)
    r.Get("/health/ready", readyHandler)

    return r
}
```

Server receives only `http.Handler`:

```go
func NewServer(
    cfg config.Config,
    router http.Handler,
) *http.Server {
    return &http.Server{
        Addr:              cfg.HTTP.Address,
        Handler:           router,
        ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
        ReadTimeout:       cfg.HTTP.ReadTimeout,
        WriteTimeout:      cfg.HTTP.WriteTimeout,
        IdleTimeout:       cfg.HTTP.IdleTimeout,
        MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
    }
}
```

Constructors stay ordinary Go funcs, callable w/o Fx in tests.

---

## 30. HTTP server lifecycle

`internal/server` owns concrete `http.Server`; Chi owns only route composition.

Recommended constructor:

```go
func New(
    cfg config.Config,
    handler http.Handler,
) *http.Server {
    return &http.Server{
        Addr:              cfg.HTTP.Address,
        Handler:           handler,
        ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
        ReadTimeout:       cfg.HTTP.ReadTimeout,
        WriteTimeout:      cfg.HTTP.WriteTimeout,
        IdleTimeout:       cfg.HTTP.IdleTimeout,
        MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
    }
}
```

Always set `ReadHeaderTimeout` explicitly. Keep request-body limits in middleware — `http.Server` has no global request-body-size limit.

### Fx lifecycle

Use Fx lifecycle hooks only for start/stop behavior:

```go
func RegisterLifecycle(
    lc fx.Lifecycle,
    srv *http.Server,
    logger *zap.Logger,
) {
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            ln, err := net.Listen("tcp", srv.Addr)
            if err != nil {
                return err
            }

            go func() {
                if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

Binding listener during `OnStart` beats starting `ListenAndServe` blindly in goroutine — startup fails synchronously when address can't bind.

On shutdown:

1. stop accepting new HTTP requests;
2. let in-flight GraphQL ops finish within shutdown deadline;
3. cancel remaining request contexts when deadline expires;
4. stop background workers;
5. flush telemetry;
6. sync Zap where appropriate;
7. close DB resources.

Shutdown deadline belongs in app config, not hard-coded in router.

---

## 31. Middleware order

Recommended global middleware order:

```text
request ID
-> panic recovery
-> trusted client-IP extraction (only when deployment topology is known)
-> security headers
-> access logging / tracing
-> request body-size limit
-> CORS
-> authentication extraction
-> request-scoped DataLoader injection
-> gqlgen handler
```

Middleware creating context values must run before middleware/handlers consuming those values.

### Baseline middleware choices

Prefer Chi's maintained middleware when it exactly matches responsibility, e.g. `middleware.RequestID`, `middleware.Recoverer`. Keep app-specific behavior (Zap access logging, auth, security policy, request-context conventions) in app-owned middleware.

Do **not** install every Chi middleware by default. Specifically:

- skip path rewriting/cleaning unless API contract requires it;
- don't enable compression blindly for every response;
- don't add global request timeouts conflicting w/ GraphQL subscriptions or long-running ops if added later;
- don't trust `X-Forwarded-For` or similar headers w/o known proxy trust model.

### Client IP

Client-IP extraction is deployment-specific. Configure exactly one strategy based on actual proxy/CDN topology. Never accept arbitrary forwarded headers direct from public internet and use as security/rate-limit identity.

### DataLoaders

DataLoader state must be request-scoped. Middleware may construct loader registry, attach to `r.Context()`, but cache must never share globally across users.

Useful access-log fields:

```text
request_id
method
route_pattern
status
duration
client_ip
actor_id
graphql_operation_name
graphql_operation_type
error_code
```

Don't log full GraphQL variables by default — may contain secrets/personal data. Don't use entire GraphQL document as metric label.

---

## 32. GraphQL operation logging

HTTP logging alone sees mostly:

```text
POST /graphql
```

Not enough operational context.

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

Don't log full query documents in prod by default.

---

## 33. Security baseline

When auth enabled:

- validate credential signature, issuer, audience, expiry, algorithm where applicable;
- use secure cookie flags for cookie-based auth;
- use CSRF protections when cookie auth needs them;
- restrict CORS to known origins;
- treat proxy/client-IP headers as untrusted unless deployment explicitly defines trusted proxies or trusted ingress header;
- enforce request-header and request-body limits;
- set `ReadHeaderTimeout` on `http.Server`;
- rate-limit auth and expensive ops where required;
- enforce GraphQL query complexity limits;
- enforce pagination bounds;
- avoid logging GraphQL variables by default;
- prevent secrets/personal data entering logs;
- use service-level authz;
- disable or explicitly control dev-only playground/introspection behavior;
- keep deps and Go toolchain patched.

Don't assume GraphQL auto-protects against expensive queries.

### CORS

CORS = HTTP policy, not GraphQL resolver concern. Configure allowed origins, methods, headers, credentials explicit from app config. Avoid permissive wildcard origins when credentials enabled.

### Rate limiting

Rate limiting may be standard `net/http` middleware. Rate-limit key must come from trustworthy identity source (authenticated actor ID or correctly-established client IP). If app ever horizontally scales, move shared rate-limit state out of process.

---

## 34. Health endpoints stay outside GraphQL

Keep health probes as ordinary HTTP endpoints.

```text
GET /health/live
GET /health/ready
```

Don't model Kubernetes/container health as:

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

Dev may expose GraphQL playground on dedicated route:

```text
GET /playground
```

Conceptual router setup:

```go
if cfg.Environment == "development" {
    r.Handle("/playground", playground.Handler(
        "GraphQL Playground",
        "/graphql",
    ))
}
```

Prod exposure must be explicit config decision. Don't accidentally couple prod API availability to playground availability.

Keep primary endpoint focused on GraphQL ops:

```text
POST /graphql
```

If GraphQL-over-HTTP GET ops or persisted-query GET requests get supported later, add deliberately + test caching/security semantics rather than enabling GET as side effect of playground.

---

## 36. Testing strategy

Use narrowest test proving behavior.

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

Service tests shouldn't require:

- Fx;
- Chi;
- gqlgen;
- GORM;
- SQLite.

Example:

```go
repo := &FakeTaskRepository{}
svc := taskservice.New(repo, zap.NewNop())
```

### Resolver tests

Resolvers may test w/ fake service:

```go
resolver := resolver.New(fakeTaskService, zap.NewNop())
```

Test transport mapping w/o real DB.

### GraphQL integration tests

Use real executable schema for important operation-level behavior.

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

Batch methods used by DataLoaders need integration tests.

---

## 37. GraphQL schema tests

Schema changes = API changes.

CI should catch unintended generated/schema drift.

Recommended checks:

```text
go test ./...
go test -race ./...
go vet ./...
gqlgen generate
git diff --exit-code
```

As project matures, consider schema compat checks for breaking changes.

At minimum, review changes that:

- remove fields;
- change nullable to non-null unsafely;
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

Runtime reqs remain:

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
http.Server
     |
     v
Chi middleware chain
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
http.Server
  listener + protocol lifecycle
  header/read/write/idle timeouts
  graceful shutdown

Chi + net/http middleware
  route matching
  request ID
  broad authentication extraction
  request limits
  security headers / CORS

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

Build one vertical GraphQL slice before adding optional infra.

### Phase 1 — application shell

1. initialize Go module;
2. create `cmd/api`;
3. add typed configuration;
4. add Uber Fx;
5. add Zap;
6. add Chi;
7. create router constructor returning `http.Handler`;
8. create concrete `http.Server` w/ explicit timeouts;
9. add Fx lifecycle and graceful shutdown.

### Phase 2 — GraphQL shell

1. add gqlgen;
2. create `gqlgen.yml`;
3. create `internal/graphql/schema`;
4. create root `Query`;
5. create root `Mutation`;
6. generate gqlgen code;
7. construct gqlgen handler through Fx;
8. mount `POST /graphql` direct with Chi;
9. add dev-only `/playground` route.

At this point:

```graphql
query {
  ping
}
```

should work end-to-end.

### Phase 3 — HTTP production baseline

Add before business surface grows:

1. request IDs;
2. panic recovery;
3. Zap access logging;
4. security headers;
5. body-size limits;
6. explicit CORS policy;
7. trusted client-IP config if required;
8. liveness/readiness endpoints;
9. `ReadHeaderTimeout`, read/write/idle timeouts, header-size limits;
10. router and middleware `httptest` coverage.

### Phase 4 — persistence

1. add GORM;
2. add SQLite driver;
3. configure SQLite;
4. add transaction manager;
5. add Goose;
6. create first migration;
7. add migration Makefile commands;
8. add DB readiness check.

### Phase 5 — first GraphQL vertical slice

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

### Phase 6 — GraphQL production baseline

Add:

- centralized error presenter;
- gqlgen panic recovery;
- operation logging;
- pagination limits;
- query complexity limit;
- GraphQL integration tests.

### Phase 7 — authentication

Add:

- identity model;
- HTTP auth extraction middleware;
- typed actor request context;
- optional broad `@auth` directive if useful;
- service-level authorization;
- authentication/security tests.

### Phase 8 — DataLoaders

Add DataLoaders only for relations showing N+1 behavior. Start w/ actual measured/query-count evidence.

### Phase 9 — operational maturity

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

New query/mutation done when:

- schema explicit;
- nullable/non-null behavior intentional;
- input types explicit;
- list fields bounded/paginated;
- generated gqlgen code current;
- resolver contains only transport mapping;
- service contains business behavior;
- service accepts `context.Context`;
- authorization enforced where required;
- persistence behind repository interfaces;
- repository uses `WithContext`;
- persistence errors translated;
- application errors map to stable GraphQL error codes;
- raw GORM/SQLite errors can't reach clients;
- logging contains request/operation context;
- expected success/error paths tested;
- required migration included;
- N+1 behavior considered;
- expensive-operation behavior bounded.

---

## 43. Architectural anti-patterns

### Fat resolver

Avoid resolvers doing auth parsing, GORM queries, permission policy, multi-table writes, external calls, response construction all in one func. Move business orchestration to service.

### Resolver calling GORM

Avoid:

```text
resolver -> *gorm.DB
```

Use:

```text
resolver -> service -> repository
```

### GraphQL models used as persistence entities

Avoid repo methods accepting gqlgen-generated transport models just to save mapping code. Prefer persistence/domain-owned types.

### GraphQL types leaking into services

Avoid service APIs accepting/returning `internal/graphql/model` types. Prefer service-owned input/result types.

### HTTP/router types inside resolvers or services

Avoid:

```go
func (s *Service) Create(r *http.Request, ...)
```

and avoid passing `chi.Router`, `chi.RouteContext`, or URL params into business logic.

Use:

```go
func (s *Service) Create(ctx context.Context, input CreateInput) (...)
```

### Exposing concrete Chi router unnecessarily

Prefer:

```go
func NewRouter(...) http.Handler
```

over making every consumer depend on `chi.Router`. Return concrete router only when caller truly needs Chi-specific route registration/inspection.

### Middleware registered after routes

Register global `r.Use(...)` middleware before routes. Use `Group`, `With`, `Route` for scoped middleware instead of mutating parent chain after route registration.

### Blindly trusting forwarded IP headers

Don't treat `X-Forwarded-For`, `X-Real-IP`, or CDN headers as trustworthy w/o documented proxy topology. Spoofable client IP must not become basis for authz or rate limiting.

### Global DataLoader cache

Default DataLoaders are request-scoped. No process-global loader cache across users.

### Repository from DataLoader bypasses policy

Don't use DataLoaders as shortcut around authz. Loader results must stay safe for requesting actor/use case.

### Unbounded lists

Avoid potentially large GraphQL list fields w/o pagination.

### Business errors encoded as ad-hoc strings

Clients should depend on stable `extensions.code` values, not exact human error messages.

### REST envelope inside GraphQL

Don't wrap every GraphQL operation in REST-style `{data,error}` payload just to imitate REST. Use GraphQL `data` + `errors` unless domain genuinely needs payload object.

### GraphQL directives as the whole authorization layer

Directive may implement broad access guard; service stays authoritative for resource and domain authz.

### Automatic schema migration

Don't run unconstrained GORM `AutoMigrate` in prod startup. Use Goose.

---

## 44. What remains from the previous REST-oriented architecture

Unchanged from original architecture direction:

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

Transport refinement:

```text
EARLIER REST-ORIENTED SHAPE

router
  -> middleware
  -> controller
  -> service
  -> repository
  -> GORM
  -> SQLite

GRAPHQL-FIRST SHAPE

http.Server
  -> Chi / net-http middleware
  -> gqlgen
  -> resolver
  -> service
  -> repository
  -> GORM
  -> SQLite
```

Switch from Gin to Chi intentionally limited to HTTP boundary. Resolver, service, repository, database, migration, app-lifecycle boundaries unchanged.

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
HTTP framework                 ->   Chi + standard net/http
manual/Wire-style composition  ->   Uber Fx
persistence stack              ->   GORM
server SQL DB patterns         ->   SQLite
migration mechanism            ->   Goose
logging                        ->   Zap
```

Borrow **principles and boundaries**, not another repo's package tree mechanically.

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

HTTP construction:

```text
http.Server
  -> http.Handler
      -> Chi router
          -> net/http middleware
          -> gqlgen handler
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
  -> gqlgen handler (http.Handler)
  -> Chi router (http.Handler)
  -> http.Server
```

For related-object fields:

```text
field resolver
  -> request-scoped DataLoader
  -> repository batch method
```

For authorization:

```text
HTTP middleware/directive
  -> broad authentication/access requirement

Service
  -> authoritative resource/domain authorization
```

Most important boundaries:

> **GraphQL is the transport contract. Services are the application contract. Repositories are the persistence contract. `http.Handler` is the HTTP composition contract.**

Future design decision mixing these responsibilities needs concrete reason, not convenience.

---

## 47. Production OSS and upstream reference patterns

Architecture intentionally informed by active upstream/production OSS codebases, tailored to this project's smaller scope.

### go-chi/chi

Reference: <https://github.com/go-chi/chi>

Patterns adopted:

- standard `net/http` handlers and middleware as fundamental abstraction;
- small composable router surface;
- middleware-driven request context;
- route grouping only when it improves ownership or policy clarity;
- explicit deployment-aware client-IP handling rather than blindly trusting forwarded headers.

### gqlgen

Reference: <https://github.com/99designs/gqlgen>

Patterns adopted:

- schema-first GraphQL;
- executable GraphQL server exposed as `http.Handler`;
- authentication data propagated through `context.Context`;
- centralized GraphQL error/recovery behavior;
- transport-specific concerns kept outside business services.

### Stash

Reference: <https://github.com/stashapp/stash>

Stash = substantial Go app using gqlgen + Chi in same HTTP stack. Lesson: Chi can stay outer HTTP/middleware boundary while gqlgen owns GraphQL execution + request-scoped loader behavior.

### GUAC

Reference: <https://github.com/guacsec/guac>

GUAC = large Go OSS codebase with both gqlgen + Chi. Its scale reinforces keeping API, transport, app concerns separated — don't let GraphQL-generated types or router details become the domain model.

### Probo

Reference: <https://github.com/getprobo/probo>

Probo = active production-oriented Go OSS app exposing GraphQL API, currently using gqlgen + Chi. Repo documents explicit rule: unexpected internal errors must not leak to API clients. Reinforces this doc's centralized GraphQL error mapping + safe-error policy.

### Reference rule

Don't copy package names or directory trees just cuz large project uses them. Adopt pattern only when it solves a responsibility that exists here. Project should stay smaller than its references until own complexity demands more layers.