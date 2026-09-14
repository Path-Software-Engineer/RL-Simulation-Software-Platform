[CmdletBinding()]
param([switch]$DeleteData)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

if ($DeleteData) {
    docker compose down --volumes --remove-orphans
} else {
    docker compose down --remove-orphans
}
if ($LASTEXITCODE -ne 0) { throw "The local platform could not be stopped cleanly." }
