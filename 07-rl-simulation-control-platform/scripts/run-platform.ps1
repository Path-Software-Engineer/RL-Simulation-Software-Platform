[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

docker compose up --build --detach --wait --wait-timeout 240
if ($LASTEXITCODE -ne 0) {
    docker compose ps
    docker compose logs --tail 120
    throw "The RL Simulation Control Platform did not become healthy."
}

docker compose ps
Write-Host "Web:     http://127.0.0.1:3000" -ForegroundColor Cyan
Write-Host "API:     http://127.0.0.1:8080" -ForegroundColor Cyan
Write-Host "OpenAPI: http://127.0.0.1:8080/openapi.json" -ForegroundColor Cyan
