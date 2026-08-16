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

## Deployment profiles

The three sprint checkpoints use local Docker Compose and create no public cloud resource. The
release profile maps the same boundaries to one HTTP-activated Azure Container App and ACR, with
Neon PostgreSQL/TimescaleDB as durable truth. Caddy, Nuxt, Go, Python and a transient Redis sidecar
share one replica that scales completely to zero; migrations execute as a manual,
checksum-protected Container Apps Job.

The Go API uses Neon's pooled connection while migrations use the direct endpoint. Redis remains a
stream transport, never the durable source of truth. See
`ADR-005-zero-cost-azure-container-apps-and-neon.md` and `docs/azure-neon-release.md`.

## Sprint 2 DQN flow

The registered DQN artifact adds training configuration, not executable code. The Python worker
constructs the bounded neural Q-network internally, samples its replay buffer and synchronizes the
target network at the registered interval. Each episode is projected before its metrics so the
dashboard can observe confirmed progress while the run remains `running`.

```text
Nuxt command → Go outbox → Redis command → bounded DQN worker
                                              ↓
                         episode + metric events per episode
                                              ↓
                    Go inbox/projector → TimescaleDB → Nuxt charts
```

See `ADR-003-bounded-dqn-training-and-observability.md` for the execution and evidence boundary.

## Sprint 3 world-model flow

The world-model artifact contains transition examples and a fixed action plan, never executable
code. The worker fits an average displacement for each allowlisted action, generates an
autoregressive prediction and executes the same action against the real Gridworld. Each projected
transition stores both paths and their Manhattan error.

```text
versioned examples → empirical action deltas → predicted next state ─┐
registered actions → real Gridworld step → expected next state ─────┼→ durable comparison
                                                                    └→ error/risk metrics
```

Existing episode and transition resources remain authoritative. Migration 0003 adds nullable
world-model fields so Sprint 1/2 evidence retains its original shape. See
`ADR-004-empirical-world-model-and-rollout-error.md`.
