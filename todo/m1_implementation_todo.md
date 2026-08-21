# M1 — Domain and persistence

GitHub milestone: [M1](https://github.com/soumajitgh/mobicode/milestone/1) · Project: [Cloud Coding Runtime — MVP](https://github.com/users/soumajitgh/projects/2)

## Delivery order

- [ ] [#10](https://github.com/soumajitgh/mobicode/issues/10) Define the Mobicode-owned project/runtime domain. Keep OpenCode and Docker out of domain models; the control plane is canonical.
- [ ] [#11](https://github.com/soumajitgh/mobicode/issues/11) Add `projects`/`sandboxes` migrations and services. Model desired and actual lifecycle state, provider/image metadata, last operation/error, and guarded transitions. Test valid/invalid transitions with real SQLite.
- [ ] [#12](https://github.com/soumajitgh/mobicode/issues/12) Add session/run migrations. Public IDs are Mobicode UUIDs; runtime session IDs are private metadata. Model queued/running/completed/failed/aborted transitions.
- [ ] [#13](https://github.com/soumajitgh/mobicode/issues/13) Persist replayable project events, pending permission/question requests, and redacted audit records. Define retention and durable-versus-ephemeral policy.
- [ ] [#14](https://github.com/soumajitgh/mobicode/issues/14) Add authenticated GraphQL project lifecycle operations. Resolvers stay thin, regenerate gqlgen, and never expose runtime IDs/endpoints/configuration.

## Implementation checklist

- [ ] Use numbered SQL migrations with tested down migrations; do not use GORM AutoMigrate.
- [ ] Add package-local domain tests plus real-SQLite repository tests under existing testing conventions.
- [ ] Add authenticated GraphQL HTTP tests for each new public mutation/query path.
- [ ] Record lifecycle failure in persistent state rather than returning an unstructured error.
- [ ] Confirm stop is non-destructive and deletion is the only durable-workspace destruction intent.

## Exit evidence

- [ ] `make test`, `make test-race`, and relevant GraphQL generation pass.
- [ ] A restart can reconstruct project/sandbox/session/run state without consulting OpenCode.
- [ ] PRs close their linked issues and are marked Done in the Project.
