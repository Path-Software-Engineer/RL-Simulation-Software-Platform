# Sprint 1 Technical Stories

## TS-P7-S1-001 — Versioned environment and policy registry

**Need.** Reject arbitrary adapters while retaining reproducible RL semantics.

**Acceptance.** UUID/version/space compatibility, SHA-256 and provenance are validated. **State:**
Implemented. **Evidence:** `artifacts/`, database seeds, `PolicyRegistry`.

**User Stories:** US-P7-S1-001, US-P7-S1-003.

## TS-P7-S1-002 — Durable lifecycle and idempotent control API

**Need.** Keep state changes transactional and auditable.

**Acceptance.** Domain transitions, request hashes, audit rows and problem details are stable.
**State:** Implemented. **Evidence:** Go domain/application/Postgres packages and OpenAPI.

**User Stories:** US-P7-S1-001, US-P7-S1-004, US-P7-S1-006.

## TS-P7-S1-003 — Real episode runner outside HTTP

**Need.** Execute the controlled RL workload without tying it to an API request.

**Acceptance.** A hash-verified Q-Learning/SARSA policy runs in Python and returns real transitions.
**State:** Implemented. **Evidence:** `workers/rl-runner/src`, pytest suite, direct runner check.

**User Stories:** US-P7-S1-001, US-P7-S1-002, US-P7-S1-005.

## TS-P7-S1-004 — Reliable Redis Streams boundary

**Need.** Tolerate duplicate delivery and malformed messages.

**Acceptance.** Versioned envelopes, consumer groups, outbox, inbox, retries and bounded DLQ exist.
**State:** Implemented. **Evidence:** AsyncAPI/schemas, dispatcher, consumer and projector.

**User Stories:** US-P7-S1-001, US-P7-S1-004.

## TS-P7-S1-005 — Temporal episode persistence

**Need.** Preserve ordered trajectory and sampled metrics without unbounded reads.

**Acceptance.** Episode/transition constraints, Timescale hypertable, aggregate and retention policy
are migrated; query limits are enforced. **State:** Implemented. **Evidence:** migration and query API.

**User Stories:** US-P7-S1-002, US-P7-S1-005.

## TS-P7-S1-006 — Confirmed-state WebSocket projection

**Need.** Notify the browser without making an ephemeral socket authoritative.

**Acceptance.** Bearer-protected reads/commands, authenticated subprotocol, origin check, heartbeat,
bounded channel, reconnect/backoff and REST resync exist.
**State:** Implemented. **Evidence:** `hub.go`, router and `useRunStream.ts`.

**User Stories:** US-P7-S1-004.

## TS-P7-S1-007 — Accessible episode visualizer

**Need.** Explain policy and trajectory across desktop, tablet and mobile.

**Acceptance.** Semantic grid cells, textual legend, keyboard focus, reduced motion, responsive
layout and loading/error states exist. **State:** Implemented; visual review remains an acceptance
gate. **Evidence:** Nuxt components, CSS and component tests.

**User Stories:** US-P7-S1-002, US-P7-S1-003, US-P7-S1-005.

## TS-P7-S1-008 — Reproducible local integration and quality gate

**Need.** Prove the cross-layer behavior rather than isolated compilation.

**Acceptance.** Pinned images, health checks, static/formal validators, unit suites and exact live
episode smoke run in one gate. **State:** Implemented; execution evidence must be recorded before
release. **Evidence:** `docker-compose.yml`, `infra/docker`, `scripts/run-quality-gate.ps1`.

**User Stories:** all Sprint 1 stories.

| Technical Story | Related User Stories | Primary evidence |
|---|---|---|
| TS-P7-S1-001 | 001, 003 | Artifacts and registry |
| TS-P7-S1-002 | 001, 004, 006 | Go API and migration |
| TS-P7-S1-003 | 001, 002, 005 | Python runner |
| TS-P7-S1-004 | 001, 004 | AsyncAPI and Redis consumers |
| TS-P7-S1-005 | 002, 005 | Timescale migration |
| TS-P7-S1-006 | 004 | WebSocket resync flow |
| TS-P7-S1-007 | 002, 003, 005 | Nuxt visualizer |
| TS-P7-S1-008 | all | Compose and quality gate |
