# Sprint 3 controlled rollout summary

## Direct deterministic evidence

- Profile: `world-model-gridworld-v1.json`.
- Model: empirical action delta, version 1.0.0.
- Versioned transition examples: 12.
- Rollout steps: 11.
- Real result: goal reached, one collision, total reward 8.64.
- First predicted/expected divergence: step 2.
- Maximum per-step Manhattan error: 1 cell.
- Final accumulated Manhattan error: 9.00 cells.
- Persistible metric samples: 34.

## Interpretation

The result proves that a fitted transition model and the real environment execute the same actions
and produce durable comparison evidence. The obstacle collision is intentionally not hidden: the
model has no obstacle representation, so its imagined path drifts.

This is not evidence of reliable planning, generalization, robustness, safety or cloud deployment.
The Sprint 3 Compose smoke passed on 2026-08-16 and verified the live local PostgreSQL projection.
Azure deployment and Neon connectivity remain separately unverified.
