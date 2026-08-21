# M2 — Sandbox provider

GitHub milestone: [M2](https://github.com/soumajitgh/mobicode/milestone/2) · Project: [Cloud Coding Runtime — MVP](https://github.com/users/soumajitgh/projects/2)

## Delivery order

- [ ] [#15](https://github.com/soumajitgh/mobicode/issues/15) Deliver a provider-agnostic sandbox lifecycle; domain code must not call Docker.
- [ ] [#16](https://github.com/soumajitgh/mobicode/issues/16) Define `SandboxProvider`, neutral types, and a deterministic fake.
- [ ] [#17](https://github.com/soumajitgh/mobicode/issues/17) Implement local Docker create/inspect/start/stop/delete/exec/file/endpoint operations idempotently.
- [ ] [#18](https://github.com/soumajitgh/mobicode/issues/18) Bind one durable workspace to one project and preserve it across runtime replacement.
- [ ] [#19](https://github.com/soumajitgh/mobicode/issues/19) Enforce sandbox isolation and CPU/memory/PID limits.

## Implementation checklist

- [ ] Keep Docker client/types in `internal/sandbox/docker`; use the interface elsewhere.
- [ ] Derive resource names from validated project IDs and record only provider metadata in Mobicode.
- [ ] Forbid host Docker socket, broad host mounts, Mobicode DB, master encryption key, and cross-project volumes.
- [ ] Make all lifecycle operations retry-safe and make provider drift visible to M9 reconciliation.
- [ ] Test provider policy by inspecting produced specs, including negative forbidden-mount/port cases.

## Exit evidence

- [ ] A fake-provider suite verifies domain behavior and a Docker-backed suite verifies provider behavior.
- [ ] A project sandbox can run a health command, stop, start, and retain workspace contents.
