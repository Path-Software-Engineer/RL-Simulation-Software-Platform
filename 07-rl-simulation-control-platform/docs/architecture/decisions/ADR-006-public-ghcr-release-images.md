# ADR-006: Public GHCR release images remove the fixed Azure registry cost

- **Status:** Accepted
- **Date:** 2026-09-24

## Context

ADR-005 selected Azure Container Registry Standard based on a time-limited new-account grant. The
free Azure credit later expired by calendar time, and ACR Standard remained a fixed daily-cost risk
if the subscription were reactivated. Container Apps can scale to zero, but ACR cannot.

## Decision

The five application images are published to public GitHub Container Registry packages:

```text
ghcr.io/path-software-engineer/rl-simulation-control-platform-control-api:<commit>
ghcr.io/path-software-engineer/rl-simulation-control-platform-rl-runner:<commit>
ghcr.io/path-software-engineer/rl-simulation-control-platform-web:<commit>
ghcr.io/path-software-engineer/rl-simulation-control-platform-migrate:<commit>
ghcr.io/path-software-engineer/rl-simulation-control-platform-gateway:<commit>
```

GitHub Actions publishes every image from the same commit with an immutable twelve-character SHA
tag, provenance and an OCI source label. The packages must be Public before deployment. Container
Apps then pulls anonymously, without an Azure registry resource or long-lived registry credential.

The migration script refuses to remove the legacy ACR until the live Container App and migration
Job reference all five expected GHCR images and the remote acceptance flow has passed.

## Cost boundary

This removes the fixed Azure registry resource. GitHub currently documents Container Registry
image storage and bandwidth as free, but that policy can change. Azure Container Apps Consumption
and Neon Free remain bounded by their respective grants, not guaranteed zero-bill services.

## Consequences

- The public Azure URL and same-origin routing do not change.
- Image packages expose application binaries publicly; source already remains in a public repository.
- New GHCR packages default to Private and require a one-time explicit visibility change.
- A clean Git commit remains the release identity across source, images, Bicep and acceptance.
- The old `p7rlqmfakszb6jgus` registry can be deleted only after successful GHCR cutover evidence.
