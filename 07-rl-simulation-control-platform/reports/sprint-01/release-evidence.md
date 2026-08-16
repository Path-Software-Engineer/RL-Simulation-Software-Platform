# Sprint 1 release acceptance

Status: **TECHNICAL DELIVERY CHECKPOINT — automated acceptance passed; visual certification not
archived**.

## Verified in the repository and external PowerShell on 2026-08-15

- OpenAPI, AsyncAPI, local references, schemas and examples passed structural validation.
- Two policy artifacts and the environment manifest passed SHA-256 verification.
- The direct registered Q-Learning episode reached the goal in 10 transitions with 0 collisions
  and 9.64 total reward.
- Python Ruff and pytest passed with 8 tests.
- Go formatting and test validation passed in the pinned container.
- Nuxt typecheck, unit tests and production image build passed.
- TimescaleDB, Redis, migrations, Go API, Python runner and Nuxt became healthy under Compose.
- The real cross-layer smoke persisted one episode with 10 ordered transitions, 9.64 reward,
  0 collisions and `goal-reached` terminality.
- Repository UTF-8, final-newline, trailing-space, broken-text, secret and Git-whitespace checks
  passed for 85 text files.

## Evidence boundary at tag time

The user-provided terminal transcript records a single uninterrupted `1/8` through `8/8` quality
gate pass and clean Compose shutdown. Browser captures at 1440 px, 768 px and 390 px, plus the
keyboard, reduced-motion, console, reconnect and resync review, were not archived before the user
requested the Sprint 1 delivery tag and Sprint 2 start. This tag therefore records a reproducible
technical checkpoint, not independent visual certification or production readiness.

## Tag decision

Create `v0.1.0-sprint-01-gridworld-agent-visualizer` on the scoped Sprint 1 commit before any Sprint 2
changes. Do not use the tag to claim DQN training, broad policy performance, cloud deployment or
visual certification.
