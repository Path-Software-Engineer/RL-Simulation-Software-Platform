# Sprint 1 — Gridworld Agent Visualizer

## Objective

Register one controlled Gridworld, execute versioned tabular agents outside HTTP, persist real
episode evidence and make state/action/reward/terminality understandable in an accessible Nuxt UI.

## Scope and outputs

- Nuxt 4.5 application shell, Grid renderer, policy overlay and episode player.
- Go 1.26 control API with run state machine, problem details, idempotency and WebSocket resync.
- Python 3.12 runner for registered Q-Learning and SARSA policy artifacts.
- PostgreSQL 17 + TimescaleDB 2.29 durable model and Redis 8.8 message boundary.
- OpenAPI 3.1, AsyncAPI 3.0, JSON Schemas, manifests, stories, ADRs, threat model and runbook.

## Day traceability

| Days | Evidence |
|---|---|
| 1002–1004 | Charter, glossary, stack ADR and monorepo boundaries |
| 1005–1008 | HTTP contract, migration, Go vertical slice and Nuxt shell |
| 1009–1012 | Gridworld rules, environment/policy registry and deterministic runner |
| 1013–1015 | Redis Streams contracts, integrated runtime and accessible grid |
| 1016–1019 | Player contract, temporal persistence, bounded queries and WebSocket resync |
| 1020–1022 | Player/feedback, release gates, evidence and release-candidate review |

## Official controlled result

The registered Q-Learning policy with seed 7 reaches the goal in 10 steps, with 0 collisions and
9.64 total reward. This single deterministic result validates integration; it does not validate
generalization or production readiness.

## Acceptance status

Implementation and repository validators are present. The sprint remains a release candidate until
the complete containerized gate and real browser review are recorded in `review.md`. No Sprint 2
artifact or capability is active.

## Evidence index

- [Week 1 review](week-01/review.md)
- [Week 2 exploration](week-02/exploration.md) and [review](week-02/review.md)
- [Week 3 exploration](week-03/exploration.md) and [review](week-03/review.md)
- [Controlled episode summary](../../../reports/sprint-01/controlled-episode-summary.md)
- [Release acceptance](../../../reports/sprint-01/release-evidence.md)
