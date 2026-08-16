# Changelog

## [Unreleased]

## [v0.3.0-sprint-03-world-model-rollout-viewer] - 2026-08-16

### Added

- Content-addressed empirical world-model artifact with 12 transition examples and one bounded
  11-action rollout plan.
- Real environment and autoregressive model execution with persisted predicted coordinates,
  Manhattan step error, accumulated error and rollout-risk metrics.
- Migration 0003 and OpenAPI/AsyncAPI 0.3 contracts that preserve nullable Sprint 1/2 transitions.
- Nuxt next-state comparison grids, horizontal rollout inspector, exact table, error chart and
  planning-risk cards.
- Direct validator, Python tests and a Sprint 3 cross-layer smoke while retaining Sprint 1/2 smokes.

### Evidence boundary

- The controlled direct profile reaches the real goal in 11 steps with one collision, first model
  divergence at step 2 and 9.00 cells of accumulated error.
- The model omits obstacle structure and is evaluated on one versioned plan; completion is not
  evidence of planning reliability, generalization or safety.
- The full eight-stage containerized gate and all three live cross-layer smoke flows passed on
  2026-08-16.
- Independent browser captures, Azure deployment and Neon connectivity are not evidence contained
  in this technical delivery tag.

## [v0.2.0-sprint-02-dqn-training-dashboard] - 2026-08-16

### Added

- Bounded seeded DQN with a one-hidden-layer Q-network, replay buffer, epsilon decay and target
  network synchronization.
- Versioned DQN training artifact plus migration 0002 for the registered profile and ten training
  metric families.
- Per-episode event projection for 40 episodes and 400 persisted metric samples.
- Nuxt DQN observatory with reward/moving-average, epsilon and loss charts, accessible tables,
  progress, summary cards and action distribution.
- Direct DQN validator, runner tests and a real Sprint 2 cross-layer smoke test.

### Evidence boundary

- The seed-11 direct profile records best reward 7.64, final moving average -33.13 and 37.5%
  success. This is evidence of a real bounded run, not convergence.
- The full eight-stage containerized gate and both live cross-layer smoke flows passed on
  2026-08-16: the tabular checkpoint plus 40 DQN episodes and 400 persisted metric samples.
- Independent browser captures and visual certification were not archived before this technical
  delivery tag.

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
