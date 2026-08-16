# Local operations runbook

## Start and verify

```powershell
Copy-Item .env.example .env
# Change OPERATOR_TOKEN in .env and use that same value in the browser unlock form.
.\scripts\run-platform.ps1
.\scripts\smoke-test.ps1
docker compose ps
```

Every service must report healthy; `migrate` must report a successful completion. A started
container is not sufficient acceptance.

Sprint 2 acceptance runs both real flows:

```powershell
.\scripts\smoke-test.ps1
.\scripts\smoke-test-sprint-02.ps1
```

The second flow may take about one minute with presentation delay enabled. It must finish with 40
episodes, 400 metric samples, epsilon `1.0 → 0.05`, a non-empty action distribution and a consistent
latest episode trace.

The OpenAPI document remains public locally, while every `/api/v1` route requires Bearer and the
run WebSocket requires `rl-run-v1` plus the same operator token as its subprotocols.

## Inspect failures

```powershell
docker compose logs --tail 200 control-api
docker compose logs --tail 200 rl-runner
docker compose logs --tail 100 timescaledb redis web
docker compose exec redis redis-cli XLEN rl.dead-letter.v1
```

Do not delete the DLQ to hide rejected messages. Record and fix the reason, then intentionally
replay a corrected versioned command.

## Reset local evidence

```powershell
.\scripts\stop-platform.ps1 -DeleteData
.\scripts\run-platform.ps1
```

This deletes only the Compose volumes for this project and must be used deliberately.

## Release acceptance

Run `.\scripts\run-quality-gate.ps1`. Then inspect desktop, tablet and mobile widths, keyboard
navigation, reduced motion, console and network requests in the real browser. Only after both
automated and visual evidence pass may the release branch and tag be created.

The Sprint 2 technical delivery tag is `v0.2.0-sprint-02-dqn-training-dashboard`. Its automated
acceptance passed on 2026-08-16. Independent browser captures were not archived with that tag and
must not be inferred from the containerized gate.
