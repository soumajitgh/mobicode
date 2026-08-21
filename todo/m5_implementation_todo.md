# M5 — Credentials and settings

GitHub milestone: [M5](https://github.com/soumajitgh/mobicode/milestone/5) · Project: [Cloud Coding Runtime — MVP](https://github.com/users/soumajitgh/projects/2)

## Delivery order

- [ ] [#40](https://github.com/soumajitgh/mobicode/issues/40) Establish control-plane ownership of credentials/settings.
- [ ] [#41](https://github.com/soumajitgh/mobicode/issues/41) Encrypt canonical provider credentials with revisions and metadata-only reads.
- [ ] [#42](https://github.com/soumajitgh/mobicode/issues/42) Reconcile stale credential/settings replicas.
- [ ] [#43](https://github.com/soumajitgh/mobicode/issues/43) Add provider connect/disconnect GraphQL operations.
- [ ] [#44](https://github.com/soumajitgh/mobicode/issues/44) Render/apply versioned credential snapshots in sandboxes.
- [ ] [#45](https://github.com/soumajitgh/mobicode/issues/45) Add normalized global/project coding settings and merge/render rules.

## Implementation checklist

- [ ] Use server-side authenticated encryption; secrets never appear in logs, events, GraphQL reads, or mobile state.
- [ ] Make control-plane records canonical; sandbox auth/config files are generated replicas only.
- [ ] Track desired/applied revisions and fan out safely on create, start, update, revoke, and reconciliation.
- [ ] Preserve repository-owned project config and surface overrides; do not expose a raw OpenCode configuration editor.
- [ ] Emit redacted audit events for provider/settings changes.

## Exit evidence

- [ ] Credential rotation/revocation and settings updates converge across running/new sandboxes.
- [ ] Unit/integration coverage proves encryption, revision, merge, rendering, and redaction behavior.
