# M6 — Execution coordinator

GitHub milestone: [M6](https://github.com/soumajitgh/mobicode/milestone/6) · Project: [Cloud Coding Runtime — MVP](https://github.com/users/soumajitgh/projects/2)

## Delivery order

- [ ] [#46](https://github.com/soumajitgh/mobicode/issues/46) Serialize mutating work per project.
- [ ] [#47](https://github.com/soumajitgh/mobicode/issues/47) Add durable exclusive lock and FIFO queue.
- [ ] [#48](https://github.com/soumajitgh/mobicode/issues/48) Submit prompts as non-blocking queued runs.
- [ ] [#74](https://github.com/soumajitgh/mobicode/issues/74) Support queued and active cancellation.
- [ ] [#50](https://github.com/soumajitgh/mobicode/issues/50) Recover interrupted runs/locks after restart.

## Implementation checklist

- [ ] Many sessions may exist, but exactly one code-mutating run may hold a project workspace lock.
- [ ] Persist queue/lock/run transitions atomically enough to recover from crashes; never infer completion from prompt submission.
- [ ] Return GraphQL promptly after queueing, forward active abort through `CodingRuntime`, and make cancellation idempotent.
- [ ] Reconcile orphaned RUNNING work and stale locks after restart without allowing concurrent checkout mutation.

## Exit evidence

- [ ] Competing prompts are ordered and serialized; cancellation/restart tests cover queue and active paths.
