# ADR-001: pinned stack and runtime boundaries

- Status: Accepted
- Date: 2026-08-14

## Decision

Sprint 1 pins Nuxt 4.5, Go 1.26.5, Python 3.12.13, TimescaleDB 2.29 on PostgreSQL
17 and Redis 8.8. Nuxt renders confirmed resources, Go/Gin owns commands and durable state, and a
separate Python process executes the allowlisted Gridworld adapter.

## Rationale

The separation prevents request handlers and the browser from becoming training runtimes. Redis
Streams transport bounded JSON messages; PostgreSQL/TimescaleDB remains the source of truth.

## Consequences

- A run command is accepted before execution is confirmed.
- Pause, resume and cancel remain requested states until runner events confirm them.
- WebSockets notify clients of durable projections; reconnecting clients resync over REST.
- Binary artifacts never travel through PostgreSQL or Redis Streams.
