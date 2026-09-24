# ADR-005: Zero-cost Azure Container Apps and Neon release boundary

- **Status:** Superseded by ADR-006 for registry hosting
- **Date:** 2026-08-16

## Context

The three sprint checkpoints run as a local Docker Compose topology. The public portfolio release
must preserve the real Go, Python and Nuxt boundaries without creating an always-on cloud bill.
PostgreSQL remains durable truth; Redis Streams is transient command/event transport.

Azure Managed Redis and permanently warm Container Apps both create ongoing cost, so they violate
the explicit zero-spend release constraint even when they are technically convenient.

## Decision

The zero-cost demo topology uses:

- Azure Container Registry Standard, covered for 12 months by the documented new-account grant,
  with admin credentials disabled and a user-assigned `AcrPull` identity.
- One Azure Container App on the Consumption plan containing a Caddy gateway, Nuxt, Go, Python and
  an internal Redis sidecar in the same replica.
- `minReplicas: 0` and `maxReplicas: 1`; the complete replica activates only on public HTTP traffic
  and returns to zero when idle.
- No Azure Managed Redis and no Log Analytics workspace. Real-time logs remain available while a
  replica is active, but Azure stores no application logs.
- One manual Container Apps Job for ordered, checksum-protected schema migrations.
- Neon Free with PostgreSQL and `timescaledb`; the API receives the pooled URL and migrations
  receive a separate direct URL.

Caddy owns the single public HTTPS/WSS ingress and proxies control-plane routes to Go and all other
routes to Nuxt. Redis is deliberately ephemeral: durable runs, outbox rows, projections, metrics
and audit stay in Neon. A scale-to-zero event may discard transport state, so this topology is a
bounded portfolio demo, not a production availability claim.

## Cost boundary

The topology is designed to remain inside the currently documented Azure free grants. It removes
every permanently allocated replica and every service without a free grant. `maxReplicas: 1`
bounds consumption, and the deployment must be retired before the 12-month registry grant ends.
Pricing and subscription spending protection must be rechecked at deployment time; a budget alert
is informational and never treated as a hard spending cap.

## Secret boundary

Neon URLs and the operator token enter Bicep only as secure parameters and become Container Apps
secrets. They are never Bicep outputs, repository files or console messages.

## Migration boundary

The migration Job uses the direct Neon connection. Each migration owns its transaction and updates
`schema_migrations`. The Job also verifies a `release_migration_checksums` ledger and fails closed
on drift. The Neon Free migration stays inside the Timescale Apache feature boundary: the metric
hypertable is paired with a live PostgreSQL aggregate view and a portable retention trigger instead
of licensed continuous-aggregate or background-retention policies. Application acceptance starts
only after the Job succeeds.

## Consequences

- Azure and Neon deployment has its own evidence boundary; local Sprint tags do not prove it.
- Cold starts are expected after scale-to-zero and are acceptable for a portfolio demonstration.
- No stored Azure application logs or Redis durability are claimed for this zero-cost profile.
- A clean Git checkout and an authenticated Azure CLI are required for a traceable release.

The Container Apps, Neon and transient-Redis decisions remain active. ADR-006 replaces only the
ACR Standard and managed-identity image-pull decision with public GHCR images.
