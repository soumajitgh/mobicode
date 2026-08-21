# M9 — Recovery, hardening, and MVP release

Canonical GitHub hierarchy: [#64](https://github.com/soumajitgh/mobicode/issues/64), [#65](https://github.com/soumajitgh/mobicode/issues/65), [#66](https://github.com/soumajitgh/mobicode/issues/66), [#68](https://github.com/soumajitgh/mobicode/issues/68), [#69](https://github.com/soumajitgh/mobicode/issues/69) · Project: [Cloud Coding Runtime — MVP](https://github.com/users/soumajitgh/projects/2)

## Delivery order

- [ ] [#69](https://github.com/soumajitgh/mobicode/issues/69) Reconcile sandbox desired/actual state on startup.
- [ ] [#79](https://github.com/soumajitgh/mobicode/issues/79) Reconcile runtime health, event bridges, and credential/settings drift.
- [ ] [#65](https://github.com/soumajitgh/mobicode/issues/65) Add E2E acceptance tests.
- [ ] [#66](https://github.com/soumajitgh/mobicode/issues/66) Complete security/resource/audit/observability hardening.
- [ ] [#68](https://github.com/soumajitgh/mobicode/issues/68) Write operator runbook and release checklist.

## Implementation checklist

- [ ] On control-plane restart, compare desired state with provider inspection, repair safe drift, health-check runtime, and persist retryable error state.
- [ ] Reconnect runtime bridges and converge only generated config/credential replicas based on canonical revisions.
- [ ] Exercise project creation, prompt completion, serialized concurrency, stop/restart, provider rotation/revocation, subscription reconnect, and pending interactions end to end.
- [ ] Re-verify no host Docker socket, control-plane database/key, cross-project secrets, public runtime endpoint, or sensitive log payload enters a sandbox.
- [ ] Ship lifecycle/runtime/queue/event/reconciliation metrics, audit records, upgrade/rollback procedure, and incident runbook.

## Release gate

- [ ] Every acceptance item in [#7](https://github.com/soumajitgh/mobicode/issues/7) has linked test/PR evidence.
- [ ] The pinned OpenCode contract suite, Go tests/race checks, mobile checks, and E2E suite pass.
- [ ] Security review approves the final sandbox policy and operator checklist.
