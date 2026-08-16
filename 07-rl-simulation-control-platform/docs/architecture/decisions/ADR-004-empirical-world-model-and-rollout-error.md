# ADR-004 — Empirical world model and rollout error

## Status

Accepted for Sprint 3.

## Context

The product needs an inspectable world-model vertical without accepting arbitrary model files,
inventing frontend predictions or implying that a small educational model can plan safely.

## Decision

Register one content-addressed profile containing 12 real transition examples and one 11-action
rollout. The Python boundary fits an average row/column displacement for each action. It predicts
autoregressively while the registered Gridworld executes the same actions. The event projection
stores predicted and expected coordinates, model version, Manhattan error and accumulated error on
the existing transition record.

## Consequences

- The computation is real, deterministic, dependency-light and independently testable.
- The missing obstacle representation creates visible drift instead of a fabricated perfect model.
- Nullable fields preserve Sprint 1/2 transition compatibility.
- The profile cannot accept browser-supplied datasets, plans, code or model binaries.
- The result does not establish planning performance, generalization, robustness or safety.
