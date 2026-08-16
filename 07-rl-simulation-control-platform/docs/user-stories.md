# Sprint 1 User Stories

## US-P7-S1-001 — Start a controlled Gridworld episode

**Story.** As an RL learner, I want to choose a registered environment and tabular policy so that I
can execute an understandable episode without running arbitrary code.

**Acceptance criteria.** Only compatible registered IDs are offered; creation returns a durable
queued run; the same idempotency key cannot create a second run.

**State.** Implemented. **Evidence:** `apps/web/app/pages/index.vue`, `services/control-api`,
`database/migrations/0001_sprint_01_gridworld.up.sql`.

## US-P7-S1-002 — Observe the agent's decisions

**Story.** As an RL learner, I want to see each state, action, reward and next state so that I can
explain the trajectory.

**Acceptance criteria.** The player uses ordered persisted transitions; terminality and truncation
are distinct; no client-side transition is invented.

**State.** Implemented. **Evidence:** `GridworldGrid.vue`, `EpisodePlayer.vue`, transition query API.

## US-P7-S1-003 — Understand the registered policy

**Story.** As a reviewer, I want policy version, provenance, hash and action overlay so that the
behavior is auditable.

**Acceptance criteria.** Overlay values come from the registered policy; SHA-256 is visible; an
unknown or modified artifact is rejected.

**State.** Implemented. **Evidence:** `artifacts/`, `PolicyRegistry`, `GET /api/v1/policies/{id}`.

## US-P7-S1-004 — Follow confirmed run state

**Story.** As an operator, I want pause, resume and cancel to reflect runner-confirmed state so that
the interface does not overstate control.

**Acceptance criteria.** Controls appear only in valid states; WebSocket reconnects with backoff;
every socket message triggers REST resync.

**State.** Implemented. **Evidence:** run domain state machine, cooperative runner control,
`useRunStream.ts`.

## US-P7-S1-005 — Review bounded episode evidence

**Story.** As a technical reviewer, I want reward, steps, collisions and terminal reason so that I
can evaluate the controlled run within its limitations.

**Acceptance criteria.** Metrics are persisted and bounded; the deterministic fixture records 10
steps, 0 collisions, 9.64 reward and goal reached.

**State.** Implemented; live acceptance requires the containerized gate. **Evidence:** runner tests,
smoke test and `reports/sprint-01/controlled-episode-summary.md`.

## US-P7-S1-006 — Annotate an observed episode

**Story.** As a reviewer, I want to attach bounded feedback so that human observations remain
traceable to an episode.

**Acceptance criteria.** Category is allowlisted, note is limited to 500 characters, command is
idempotent and audited.

**State.** Implemented. **Evidence:** feedback form, HTTP contract and persistence migration.
