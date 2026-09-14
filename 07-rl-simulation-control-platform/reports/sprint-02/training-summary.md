# Sprint 2 controlled DQN summary

## Direct deterministic evidence

- Profile: `dqn-gridworld-training-v1.json`.
- Seed: 11.
- Episodes: 40.
- Maximum steps per episode: 64.
- Persistible metric samples: 400 across ten metric families.
- Best episode reward: 7.64.
- Final 10-episode moving average: -33.13.
- Final success rate: 37.5%.
- Epsilon schedule: 1.0 to 0.05.

## Interpretation

The run proves that the network, replay, target synchronization and metric pipeline execute on real
environment transitions. The poor final moving average is not hidden or relabeled: it demonstrates
the instability and hyperparameter sensitivity expected from a small educational DQN.

This is not evidence of benchmark performance, convergence, generalization, safety or cloud
deployment. The live TimescaleDB projection is claimed only after the Sprint 2 Compose smoke passes.
