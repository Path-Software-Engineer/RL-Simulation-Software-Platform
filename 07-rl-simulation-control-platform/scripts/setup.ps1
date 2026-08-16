[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

docker version *> $null
if ($LASTEXITCODE -ne 0) { throw "Docker Desktop with the Linux engine is required." }

docker compose config --quiet
if ($LASTEXITCODE -ne 0) { throw "Docker Compose configuration is invalid." }

docker compose pull timescaledb redis migrate
if ($LASTEXITCODE -ne 0) { throw "Pinned infrastructure images could not be pulled." }

docker compose build control-api rl-runner web
if ($LASTEXITCODE -ne 0) { throw "Sprint 1 application images could not be built." }

Write-Host "OK - Project 07 Sprint 1 dependencies and images are ready." -ForegroundColor Green
