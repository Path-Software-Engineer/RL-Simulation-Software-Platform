# Sprint 1 threat model

| Threat | Boundary | Control |
|---|---|---|
| Arbitrary code or environment execution | Runner | Hard-coded adapter allowlist; no upload or module name in HTTP |
| Path traversal to a policy | Artifact registry | Resolved path must stay under the policy root |
| Policy substitution | Runner / DB | UUID, version and SHA-256 must match the registered artifact |
| Duplicate command delivery | API / messaging | Idempotency record, transactional outbox and consumer inbox |
| Oversized request or telemetry | HTTP / streams | Body limits, field bounds, query limits and stream max lengths |
| Unauthenticated commands | HTTP | Public GET allowlist; mutations require constant-time Bearer comparison |
| Public evidence overexposure | HTTP | Only bounded registered runs, episodes, metrics and transitions are readable |
| Cross-origin WebSocket misuse | Experience | Exact origin allowlist, opaque run UUID and authenticated subprotocol |
| Misleading live status | Experience | Socket carries notification only; every update resyncs from REST |
| Secret disclosure | Runtime | Environment injection, structured logs and repository secret scan |
| Unbounded compute | Runner | One registered environment, max 200 steps, one consumer, CPU/memory cap |
| Model or rollout injection | Runner | Content-addressed examples and plan; browser supplies only registered UUID |
| Fabricated model evidence | Product / persistence | Prediction and real transition are computed together and projected atomically |
| Reward interpreted as safety | Product | Explicit evidence boundary in UI, docs and reports |
| Rollout interpreted as a safe plan | Product | Error, obstacle omission and open-loop drift remain visible |

Authentication is deliberately single-operator: it is not an identity provider, RBAC or production
session service. Portfolio evidence reads are public and bounded; creating or controlling runs and
recording feedback still requires the private token. The token must be changed outside source
control, remains in browser `sessionStorage`, is compared in constant time and is never emitted by
application logs.
