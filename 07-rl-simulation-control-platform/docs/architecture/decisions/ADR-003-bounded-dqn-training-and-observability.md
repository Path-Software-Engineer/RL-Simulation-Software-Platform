# ADR-003 — Bounded DQN training and observability

## Decision

Sprint 2 executes only the repository-registered `dqn-gridworld-training-v1.json` profile. The
runner implements a 36-state, one-hidden-layer neural Q-network with replay sampling, a target
network and a linear epsilon schedule. It accepts no uploaded models, Python modules, environment
names or hyperparameters from the browser.

Training emits one versioned episode event and ten metric events per episode. PostgreSQL remains
the durable source of truth; Redis Streams transports commands and projections. The Nuxt dashboard
queries bounded metric series and never fabricates fallback points.

## Why

The product needs a real bridge from tabular control to neural approximation without opening an
arbitrary code-execution or artifact-loading boundary. A small pure-Python network keeps the
mechanics inspectable and reproducible while exposing authentic training instability.

## Consequences

- The registered profile runs exactly 40 episodes with at most 64 steps per episode in acceptance.
- Completion means the budget executed, not that reward converged or a benchmark was solved.
- Charts must expose exact data tables because color and line shape alone are insufficient evidence.
- Future Gymnasium or PyTorch workers require a new versioned profile and explicit runtime ADR.
- Cloud object-storage labs remain infrastructure preparation until provider evidence exists.
