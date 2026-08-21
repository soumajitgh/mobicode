# M3 — Sandbox image and provisioning

GitHub milestone: [M3](https://github.com/soumajitgh/mobicode/milestone/3) · Project: [Cloud Coding Runtime — MVP](https://github.com/users/soumajitgh/projects/2)

## Delivery order

- [ ] [#28](https://github.com/soumajitgh/mobicode/issues/28) Establish the durable pinned runtime contract.
- [ ] [#29](https://github.com/soumajitgh/mobicode/issues/29) Build/version the dedicated image with pinned OpenCode, least-privilege paths, entrypoint, and health check.
- [ ] [#30](https://github.com/soumajitgh/mobicode/issues/30) Start private headless OpenCode only on sandbox-local networking.
- [ ] [#31](https://github.com/soumajitgh/mobicode/issues/31) Clone public repositories idempotently into durable workspace.
- [ ] [#32](https://github.com/soumajitgh/mobicode/issues/32) Define separate, minimum-scope Git credential materialization.

## Implementation checklist

- [ ] Add `sandbox/Dockerfile`, entrypoint, bootstrap, and health-check assets; pin base image and OpenCode version.
- [ ] Provide `/mobicode/workspace`, runtime, config, secrets, state, and log boundaries with safe ownership.
- [ ] Never publish OpenCode publicly or use `latest`; version bumps require M4 contract tests.
- [ ] Record clone/default-branch metadata and recover safely from partial clones.
- [ ] Keep Git credentials separate from LLM credentials and out of checkout/logs.

## Exit evidence

- [ ] Image build and private health checks pass.
- [ ] Clone -> stop -> restart/recreate preserves Git workspace state.
