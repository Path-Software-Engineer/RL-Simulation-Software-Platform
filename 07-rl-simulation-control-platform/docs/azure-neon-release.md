# Zero-cost Azure + Neon release guide

## Status and evidence boundary

The repository contains a zero-cost portfolio topology and automated acceptance flow. It does not
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
Azure Container Registry --identity------> immutable application images
```

All runtime containers share one replica and scale to zero together. Redis contains transient
transport state only; PostgreSQL remains durable truth. Azure Managed Redis and Log Analytics are
intentionally absent. The resource group defaults to `rg-p7-rl-simulation-demo` in `centralus`.

## Zero-cost constraints

- Container Apps uses Consumption with `minReplicas: 0` and `maxReplicas: 1`. Azure documents no
  usage charge while the app is at zero and includes a monthly consumption grant.
- ACR uses Standard because new Azure accounts currently include one Standard registry for 12
  months. Delete the resource group before that grant expires.
- Application-log storage is disabled (`destination: none`), avoiding a Log Analytics resource.
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
5. Use a clean Git checkout on the release branch. Deployment refuses dirty worktrees so every
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

## Deploy

```powershell
Set-Location "C:\JeanLoa\Path-Software-Engineer\RL-Simulation-Software-Platform\07-rl-simulation-control-platform"
az version
az account show --output table
.\scripts\deploy-azure.ps1
```

The script validates the CLI, subscription, clean Git state and secret shapes; compiles Bicep;
provisions the zero-cost foundation; builds five immutable images; applies checksum-protected Neon
migrations; and validates public HTTPS/WSS plus all three real product profiles.

Successful output contains only public URLs and non-secret evidence:

```text
OK - Azure + Neon release deployment passed
Web:     https://...
API:     https://...
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
```

Acceptance requires no `Microsoft.Cache`, `Microsoft.OperationalInsights`, dedicated profile or
other unlisted resource, and the scale output must be `0 / 1`.

## Cost shutdown

Deleting the resource group removes the Azure release and registry images. It does not delete the
Neon project. Run this only when the public demo is intentionally retired:

```powershell
az group delete --name rg-p7-rl-simulation-demo
```

The destructive command is deliberately not included in automation.
