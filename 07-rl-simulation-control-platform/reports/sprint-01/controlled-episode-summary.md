# Controlled episode summary

## Configuration

- Environment: `pathfinder-grid-6x6` v1.0.0, 6×6 discrete space.
- Policy: Q-Learning v1.0.0.
- Policy SHA-256: `58d3eac12251df3a3e708eeb6544a056921db9d6f4b2f4930db4d85ae32f7a08`.
- Seed: 7. Maximum steps: 64.

## Deterministic oracle

| Metric | Expected |
|---|---:|
| Steps | 10 |
| Total reward | 9.64 |
| Collisions | 0 |
| Terminal reason | goal reached |

The repository direct-runner check proves the environment/policy behavior. The live smoke additionally
requires the same values after Redis delivery and PostgreSQL projection.

## Limitations

This is one repository-authored policy in one controlled environment. There is no training,
cross-environment evaluation, stochastic robustness test, DQN, robotics or safety certification.
