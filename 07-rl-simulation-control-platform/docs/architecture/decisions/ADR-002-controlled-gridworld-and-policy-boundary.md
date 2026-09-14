# ADR-002: controlled Gridworld and policy boundary

- Status: Accepted
- Date: 2026-08-14

## Decision

Sprint 1 registers one immutable 6x6 Gridworld version and two tabular policies: Q-Learning and
SARSA. The runner resolves environments and policies from explicit registries. User-supplied module
paths, Python expressions, policy binaries and environment code are rejected.

## Evidence meaning

The policy artifacts are deterministic educational fixtures with version, SHA-256, observation
space, action space and provenance. They demonstrate the software workflow and reproducible episode
execution, not generalization or production readiness.

## Safety boundary

Reward indicates progress inside this controlled environment. It does not establish safety,
robustness or real-world suitability.
