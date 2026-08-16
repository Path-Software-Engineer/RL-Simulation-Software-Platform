# Sprint 1 labs

The Gridworld setup, state/action card, reward map, episode runner, trajectory viewer and policy-note
experiments were promoted into the production vertical slice. Cloud storage labs remain deferred:
Sprint 1 produces only small repository manifests and durable local database evidence, so adding a
cloud bucket would not improve the accepted product flow.

## Sprint 2 DQN labs

- `tec-dqn-training-log-schema-lab`: OpenAPI metric enum, migration 0002 and event projection.
- `tec-reward-chart-lab`: episode reward plus 10-episode moving average and exact table.
- `tec-epsilon-schedule-viewer-lab`: registered linear schedule from 1.0 to 0.05.
- `tec-loss-viewer-lab`: finite mean squared TD-error series without smoothing away spikes.
- `tec-action-distribution-lab`: accumulated action counts and percentages.
- `tec-training-summary-card-lab`: moving reward, best episode, epsilon, loss, success and dominant
  action derived from persisted data.
- `docs-dqn-training-storytelling-lab`: the report keeps the negative final moving average and
  separates execution from convergence.

The AWS S3, GCP Storage and Azure Blob labs remain connection-ready roadmap items. No cloud log
upload or provider deployment is claimed by this sprint.
