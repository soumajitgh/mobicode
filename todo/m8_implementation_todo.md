# M8 — Mobile coding flow

Canonical GitHub hierarchy: [#57](https://github.com/soumajitgh/mobicode/issues/57), [#58](https://github.com/soumajitgh/mobicode/issues/58), [#59](https://github.com/soumajitgh/mobicode/issues/59), [#60](https://github.com/soumajitgh/mobicode/issues/60), [#61](https://github.com/soumajitgh/mobicode/issues/61), [#62](https://github.com/soumajitgh/mobicode/issues/62), [#63](https://github.com/soumajitgh/mobicode/issues/63) · Project: [Cloud Coding Runtime — MVP](https://github.com/users/soumajitgh/projects/2)

## Delivery order

- [ ] [#63](https://github.com/soumajitgh/mobicode/issues/63) Projects, provisioning state, and lifecycle controls.
- [ ] [#58](https://github.com/soumajitgh/mobicode/issues/58) Sessions, prompt composition, conversation, live runs, and cancellation.
- [ ] [#60](https://github.com/soumajitgh/mobicode/issues/60) Provider/settings screens with metadata-only handling.
- [ ] [#62](https://github.com/soumajitgh/mobicode/issues/62) Permission/question interactions.
- [ ] [#59](https://github.com/soumajitgh/mobicode/issues/59) Sequence persistence, reconnect, resync, and background recovery.
- [ ] [#61](https://github.com/soumajitgh/mobicode/issues/61) Terminology and accessibility review.

## Implementation checklist

- [ ] Bind UI exclusively to Mobicode GraphQL models/events; never display OpenCode, raw config, runtime IDs, or secrets.
- [ ] Refetch authoritative project/session/run/request state after event gaps or app restart.
- [ ] Make provisioning, run, permission, question, and error states explicit and actionable.
- [ ] Store only safe local state, including last project event sequence; use secure storage for existing identity material.
- [ ] Test network loss, backgrounding, server restart, event replay miss, pending interaction restoration, and accessibility labels.

## Exit evidence

- [ ] The owner can connect a provider, create a project, run a session, respond to an interaction, leave, and resume from mobile.
