# M7 — Realtime and subscriptions

GitHub milestone: [M7](https://github.com/soumajitgh/mobicode/milestone/7) · Project: [Cloud Coding Runtime — MVP](https://github.com/users/soumajitgh/projects/2)

## Delivery order

- [ ] [#51](https://github.com/soumajitgh/mobicode/issues/51) Deliver normalized reconnectable events.
- [ ] [#56](https://github.com/soumajitgh/mobicode/issues/56) Define versioned `ProjectEvent` and project sequence numbers.
- [ ] [#55](https://github.com/soumajitgh/mobicode/issues/55) Build event hub, replay, and durable/ephemeral policy.
- [ ] [#54](https://github.com/soumajitgh/mobicode/issues/54) Add owner-authenticated GraphQL WebSocket subscriptions.
- [ ] [#53](https://github.com/soumajitgh/mobicode/issues/53) Persist/resolve permissions and questions.
- [ ] [#52](https://github.com/soumajitgh/mobicode/issues/52) Add reconnect observability/failure handling.

## Implementation checklist

- [ ] Bridge private OpenCode SSE into normalized events; mobile never sees raw frames.
- [ ] Use monotonic project sequence numbers; persist important transitions and explicitly resync when replay is unavailable.
- [ ] Verify ownership during subscription setup and permission/question resolution.
- [ ] Separate ephemeral progress from durable authoritative state; restore pending interactions after restart.
- [ ] Capture reconnects, unknown frames, lag, and failure metrics without secret payloads.

## Exit evidence

- [ ] Subscription/replay/reconnect tests pass, including gap recovery and pending interactions.
