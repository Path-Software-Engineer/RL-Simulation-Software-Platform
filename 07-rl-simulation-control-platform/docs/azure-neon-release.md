# Zero-fixed-cost Azure + Neon release guide

## Status and evidence boundary

The repository contains a zero-fixed-cost portfolio topology and automated acceptance flow. It does not
claim a live deployment until `scripts/deploy-azure.ps1` reports the public Web, API and OpenAPI
URLs after all three remote smokes pass.

## Topology

```text
Browser --HTTPS/WSS--> Azure Container Apps Consumption (min 0, max 1)
                         Caddy same-origin gateway
                           |--> Nuxt
                           |--> Go API ----> Neon Free pooled endpoint
                           |      |
                           |      +--------> transient Redis sidecar
                           |                     ^
                           +--> Python runner ----+

Manual migration Job --------------------> Neon Free direct endpoint
Public GitHub Container Registry --------> immutable application images
```

All runtime containers share one replica and scale to zero together. Redis contains transient
transport state only; PostgreSQL remains durable truth. Azure Managed Redis and Log Analytics are
intentionally absent. The resource group defaults to `rg-p7-rl-simulation-demo` in `centralus`.

## Zero-cost constraints

- Container Apps uses Consumption with `minReplicas: 0` and `maxReplicas: 1`. Azure documents no
  usage charge while the app is at zero and includes a monthly consumption grant.
- Five public GHCR packages hold immutable commit-tagged images. Azure pulls them anonymously, so
  the deployment has no dedicated registry resource, registry secret or fixed registry charge.
- Application logs use the `azure-monitor` control-plane destination without any diagnostic
  setting, storage target or Log Analytics workspace, so no application logs are persisted.
- Neon must remain on its Free plan with autosuspend enabled.
- No paid Redis, NAT gateway, private endpoint, dedicated workload profile or always-on replica is
  permitted in this release profile.

These controls bound the demo to the published free grants; pricing and the subscription's spending
protection must still be checked before deployment. Budget alerts notify but do not stop resources.

## Prerequisites

1. Create a Neon Free project using a PostgreSQL version that exposes `timescaledb`.
2. Keep autosuspend enabled and confirm `CREATE EXTENSION IF NOT EXISTS timescaledb` is allowed.
3. Copy both connection strings: pooled for the API and direct for migrations. Both must use TLS
   and `channel_binding=require`.
4. Install Azure CLI 2.75.0 or newer and authenticate with `az login`.
5. Publish the five images through `.github/workflows/publish-p7-ghcr.yml` and change each package
   visibility to Public. GitHub does not make a newly published package public automatically.
6. Use a clean Git checkout on the release branch. Deployment refuses dirty worktrees so every
   image tag identifies an exact commit.

For Windows without `winget` or administrator rights:

```powershell
Set-Location "C:\JeanLoa\Path-Software-Engineer\RL-Simulation-Software-Platform\07-rl-simulation-control-platform"
.\scripts\install-azure-cli-current-user.ps1
```

Use hidden prompts so credentials are not written into PowerShell history:

```powershell
function Read-ReleaseSecret([string]$Prompt) {
    $SecureValue = Read-Host $Prompt -AsSecureString
    try {
        $Credential = New-Object System.Net.NetworkCredential -ArgumentList "", $SecureValue
        return $Credential.Password
    }
    finally {
        $SecureValue.Dispose()
    }
}

$env:NEON_DATABASE_URL = Read-ReleaseSecret "Neon pooled URL"
$env:NEON_DATABASE_URL_DIRECT = Read-ReleaseSecret "Neon direct URL"
$env:OPERATOR_TOKEN = [guid]::NewGuid().ToString("N") + [guid]::NewGuid().ToString("N")
```

Never paste these values into chat, commit them or store them in `.env`.

Publish the exact clean commit through GitHub Actions:

```powershell
gh workflow run publish-p7-ghcr.yml --ref <release-branch>
gh run watch --exit-status
```

New GHCR packages start as Private. In the GitHub organization, open each of the five package
settings and change **Package settings → Change visibility → Public**. This is irreversible; verify
the package names listed in ADR-006 before confirming.

## Deploy

```powershell
Set-Location "C:\JeanLoa\Path-Software-Engineer\RL-Simulation-Software-Platform\07-rl-simulation-control-platform"
az version
az account show --output table
.\scripts\deploy-azure.ps1
```

The script validates the CLI, subscription, clean Git state and secret shapes; compiles Bicep;
provisions the scale-to-zero foundation without ACR; verifies that all five immutable GHCR images
allow anonymous pulls; applies the
checksum-protected Neon migrations; and validates public HTTPS/WSS plus all three real product
profiles. The migration image uses the compact PostgreSQL 17 client; TimescaleDB runs in Neon, not
inside that release job. The schema deliberately uses only Neon's Timescale Apache feature set;
hourly aggregation and retention remain portable PostgreSQL objects.

The recommended migration from the legacy ACR reuses workflow-published images and removes the
exact old registry only after the public application and all three remote smokes pass:

```powershell
.\scripts\deploy-azure.ps1 -SkipBuild -RemoveLegacyAcr
```

For an emergency local publication instead of GitHub Actions, omit `-SkipBuild` and set
process-scoped `GHCR_USERNAME` plus a classic `GHCR_TOKEN` with `write:packages`. The script uses a
temporary Docker configuration and removes it after publication.

Successful output contains only public URLs and non-secret evidence:

```text
OK - Azure + Neon release deployment passed
Web:     https://...
API:     https://...
Swagger UI: https://.../docs/
OpenAPI: https://.../openapi.json
```

## Inspect a failed release

Stored application logs are disabled. While a replica or Job is active, inspect its live stream:

```powershell
az containerapp job execution list --name p7rl-migrate `
    --resource-group rg-p7-rl-simulation-demo --output table
az containerapp logs show --name p7rl-platform `
    --resource-group rg-p7-rl-simulation-demo --follow
```

Do not print Container Apps secrets while collecting evidence.

## Verify the cost guard

```powershell
az resource list --resource-group rg-p7-rl-simulation-demo `
    --query "[].{Type:type,Name:name}" --output table
az containerapp show --name p7rl-platform `
    --resource-group rg-p7-rl-simulation-demo `
    --query "properties.template.scale.{min:minReplicas,max:maxReplicas}" --output table
az acr list --resource-group rg-p7-rl-simulation-demo --output table
```

Acceptance requires no `Microsoft.Cache`, `Microsoft.OperationalInsights`, dedicated profile or
other unlisted resource, the scale output must be `0 / 1`, and the legacy ACR list must be empty.

## Cost shutdown

Deleting the resource group removes the Azure release but does not delete GHCR packages or the
Neon project. Run this only when the public demo is intentionally retired:

```powershell
az group delete --name rg-p7-rl-simulation-demo
```

Full resource-group deletion is deliberately not included in automation. The narrower legacy ACR
deletion is available only behind `-RemoveLegacyAcr`, after image-reference and remote-smoke checks.
