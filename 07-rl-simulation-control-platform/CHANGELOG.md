# Changelog

## [Unreleased]

Sprint 2 and Sprint 3 remain outside this tagged checkpoint.

## [v0.1.0-sprint-01-gridworld-agent-visualizer] - 2026-08-16

### Added

- Versioned Gridworld environment and tabular-policy contracts.
- Go/Gin control API with durable run state, audit and idempotent commands.
- PostgreSQL/TimescaleDB migrations for environments, runs, episodes, transitions and metrics.
- Redis Streams command/event boundary with consumer groups, inbox deduplication and dead-letter handling.
- Python Gridworld runner with allowlisted Q-Learning and SARSA policies.
- Nuxt/Vue Gridworld visualizer, policy overlay, episode player and confirmed-state WebSocket resync.
- Reproducible manifests, reports, tests, Docker Compose topology and repository quality gate.

### Evidence boundary

- The bundled policies are small deterministic teaching artifacts, not evidence of broad RL performance.
- Reward is an environment signal, not a safety or robustness score.
- The browser and Go API do not execute training or arbitrary Python modules.
- DQN training, robotics navigation and world-model rollouts remain deferred.
- The full eight-stage containerized quality gate and live cross-layer smoke passed on 2026-08-15.
- Browser captures and independent visual certification were not archived before this technical
  delivery tag.
