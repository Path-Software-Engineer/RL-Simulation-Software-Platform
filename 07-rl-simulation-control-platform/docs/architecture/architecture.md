# Sprint 1 architecture

## Boundaries

- **Experience:** Nuxt renders confirmed data, issues bounded commands and performs REST resync.
- **Control:** Go/Gin owns lifecycle invariants, idempotency, audit and projections. It does not run
  episodes inside HTTP requests.
- **Execution:** Python loads only registered, hash-verified tabular policies and the single
  compatible Gridworld adapter.
- **Persistence:** PostgreSQL stores product state; TimescaleDB stores bounded metric samples.
- **Messaging:** Redis Streams carries versioned commands/events with consumer groups and DLQ.

## Run sequence

1. The operator unlocks the local workspace; Nuxt authenticates REST with Bearer and WebSocket
   through the registered `rl-run-v1` subprotocol.
2. Nuxt posts an environment ID, policy ID, seed and maximum steps with an idempotency key.
3. Go validates registered records and commits the queued run, audit and outbox row atomically.
4. The dispatcher publishes `rl.run.requested.v1`; the Python consumer deduplicates it.
5. The runner verifies policy path containment, identity and SHA-256, then executes the episode.
6. Versioned events are projected through an inbox transaction into run, episode, transition and
   metric tables.
7. The WebSocket hub emits a small `run.updated` notification with a REST resync path.
8. Nuxt rereads durable resources and renders the confirmed episode.

## Consistency and failure behavior

- Commands are at-least-once; outbox/inbox/idempotency make side effects repeat-safe.
- A malformed or unsupported message goes to a bounded dead-letter stream and is acknowledged.
- Unknown environments, policies, actions and incompatible spaces fail closed.
- Pause/cancel are cooperative. The UI shows intermediate states until the runner confirms them.
- Query limits prevent unbounded telemetry responses; WebSocket never streams every transition.

## Deployment profile

Sprint 1 is a local Docker Compose release. No public cloud resource is created by this sprint.
The topology is intentionally portable and has no secret or provider-specific dependency.
