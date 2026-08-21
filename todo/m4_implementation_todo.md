# M4 — OpenCode runtime adapter

GitHub milestone: [M4](https://github.com/soumajitgh/mobicode/milestone/4) · Project: [Cloud Coding Runtime — MVP](https://github.com/users/soumajitgh/projects/2)

## Delivery order

- [ ] [#34](https://github.com/soumajitgh/mobicode/issues/34) Isolate the pinned runtime behind Mobicode’s adapter boundary.
- [ ] [#35](https://github.com/soumajitgh/mobicode/issues/35) Define `CodingRuntime`, neutral types, and fake runtime.
- [ ] [#36](https://github.com/soumajitgh/mobicode/issues/36) Add private health/session operations.
- [ ] [#37](https://github.com/soumajitgh/mobicode/issues/37) Add async prompt, abort, permission, and question operations.
- [ ] [#38](https://github.com/soumajitgh/mobicode/issues/38) Read SSE and map raw runtime events to Mobicode signals.
- [ ] [#39](https://github.com/soumajitgh/mobicode/issues/39) Add fixtures and pinned-image contract tests.

## Implementation checklist

- [ ] Keep every OpenCode type, route, config format, credential format, and raw event in `internal/runtime/opencode`.
- [ ] Use sandbox-private HTTP/SSE with deadlines, cancellation, backoff, and redacted errors; never shell-drive the TUI.
- [ ] Return async prompt submission promptly; terminal run state comes from coordinator/events.
- [ ] Treat SSE as reconnectable but lossy; unknown frames are observable and harmless.
- [ ] Contract-test health, sessions, prompt, events, permissions/questions, credentials, and settings before version upgrades.

## Exit evidence

- [ ] No public schema/package outside the adapter imports OpenCode types.
- [ ] Fixture and real pinned-runtime contract suites pass.
