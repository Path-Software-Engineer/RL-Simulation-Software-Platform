# Sprint 3 — World Model Rollout Viewer

## Goal

Make model error inspectable. A reviewer can execute one registered transition model, compare its
autoregressive prediction with the real Gridworld and follow how error propagates through a plan.

## Delivered vertical

- SHA-256-verified empirical action-delta artifact with 12 versioned examples.
- Fixed 11-action rollout executed against both predicted and real states.
- Predicted-state, model-version and Manhattan-error projection through Redis and PostgreSQL.
- OpenAPI/AsyncAPI 0.3 contracts and migration 0003.
- Nuxt comparison grids, rollout sequence, exact table, error chart and risk cards.
- Direct validator, Python tests and real Compose smoke acceptance.

## Acceptance

Run:

```powershell
.\scripts\run-quality-gate.ps1 -KeepRunning
```

The full gate passed on 2026-08-16 with one 11-step rollout, 12 training examples, expected and
predicted states on every step, first divergence at step 2, final accumulated error 9.00 and the
real goal reached. Azure and Neon remain a separate release phase.

## Evidence boundary

The model learns average action displacement but no obstacle map. One controlled rollout exposes
that limitation; it does not prove useful planning, generalization, robustness or safety. Cloud
storage labs are documented but no provider resource or deployment is claimed.
