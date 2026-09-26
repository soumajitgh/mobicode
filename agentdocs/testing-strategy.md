# Testing Strategy

This document defines the testing landscape for Mobicode during the prototype phase.

The goal is **high-signal tests**, not broad coverage. Tests should protect important behavior, critical flows, and integration boundaries without slowing development.

## Principles

- Test behavior, not implementation details.
- Prefer a few meaningful tests over many trivial ones.
- Do not chase coverage percentages.
- Add a test when a regression would meaningfully affect development or users.
- Prefer real application components over mocks when practical.
- Mock only true external dependencies.

## Test Layers

### Package Tests

Keep focused Go tests next to the package they validate.

```text
internal/auth/service_test.go
internal/onboarding/service_test.go
```

Use for:
- business rules
- validation
- authorization decisions
- state transitions
- non-trivial pure logic

Prefer external test packages (`package foo_test`) when the public API is sufficient.

### Integration Tests

Integration tests are the **primary backend test layer** during the prototype phase.

```text
tests/
└── integration/
    ├── auth_test.go
    ├── onboarding_test.go
    ├── graphql_test.go
    └── migrations_test.go
```

Use Go's standard testing tools and exercise real application boundaries:

```text
Chi router
→ handlers / GraphQL
→ services
→ repositories
→ database
```

Use them for:
- authentication flows
- protected routes
- onboarding behavior
- GraphQL/API behavior
- database constraints
- migrations

Use a temporary SQLite database and real migrations where possible.

Use Testcontainers only when testing behavior that depends on a real external service or database such as PostgreSQL, MySQL, or Redis. Do not introduce containers where SQLite or an in-process dependency is sufficient.

### End-to-End Tests

Use **Playwright** as the standard E2E framework.

```text
tests/
└── e2e/
    ├── onboarding.spec.ts
    └── auth.spec.ts
```

Playwright tests should run against the real application and treat it as a black box:

```text
Browser
→ web application
→ backend API / GraphQL
→ database
```

Keep E2E coverage small and focused on critical user journeys.

Examples:
- first boot redirects to onboarding
- create the first user and complete onboarding
- login and logout
- unauthenticated users cannot access protected pages
- authenticated users can access critical application flows
- future mobile pairing and GitHub flows once functional

As the suite grows, keep a small smoke suite for fast CI checks and reserve broader E2E coverage for critical workflows.

## Test Ownership

```text
Package tests   → Go testing
Integration     → Go testing + real application components
E2E             → Playwright + real running application
```

Do not duplicate the same behavior across every layer. Test each behavior at the lowest layer that gives sufficient confidence.

## What Not to Test Yet

Avoid adding tests for:

- trivial getters/setters
- framework or library behavior
- CSS/layout details
- snapshot-heavy UI testing
- placeholder screens
- exhaustive edge cases without practical value
- coverage targets for their own sake

## Decision Rule

Before adding a test, ask:

> If this behavior breaks, should CI stop the change?

If the answer is no, the test is probably unnecessary during the prototype phase.
