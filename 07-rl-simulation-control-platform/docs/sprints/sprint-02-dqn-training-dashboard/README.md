# Sprint 2 — DQN Training Dashboard

## Goal

Turn the Sprint 1 control plane into an observable neural-agent training product. A reviewer can
start one registered DQN profile, watch durable progress and inspect reward, moving average,
epsilon, loss, episode length, success rate and action distribution without trusting fabricated
frontend data.

## Delivered vertical

- `dqn-gridworld-training-v1.json` with immutable identity and SHA-256.
- One-hidden-layer Q-network, replay buffer, epsilon-greedy exploration and target synchronization.
- Forty bounded episodes and ten metric families emitted incrementally through Redis Streams.
- TimescaleDB constraint migration and bounded per-metric reads.
- Nuxt dashboard with six summary cards, three accessible charts, chart tables and action bars.
- Direct 40-episode validator, Python unit tests and real Compose smoke acceptance.

## Acceptance

Run:

```powershell
.\scripts\run-quality-gate.ps1 -KeepRunning
```

The eight-stage gate passed on 2026-08-16 with 40 ordered episodes, 400 metric samples, epsilon
`1.0 → 0.05`, a non-empty action distribution and a latest episode whose transition count matched
its durable summary. Independent desktop, tablet and mobile captures were not archived with the
technical delivery tag.

## Evidence boundary

The direct seed-11 profile produced best reward 7.64, final 10-episode moving average -33.13 and
37.5% success. Negative or oscillating curves are valid observations. This sprint demonstrates
training mechanics and observability, not convergence, robustness, safety or production readiness.
