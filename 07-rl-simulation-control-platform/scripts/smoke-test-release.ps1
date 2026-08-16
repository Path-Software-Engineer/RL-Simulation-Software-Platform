[CmdletBinding()]
param(
    [Parameter(Mandatory)][ValidatePattern('^https://')][string]$ApiBaseUrl,
    [Parameter(Mandatory)][ValidatePattern('^https://')][string]$WebBaseUrl,
    [string]$OperatorToken = ""
)

$ErrorActionPreference = "Stop"
if (-not $OperatorToken) { $OperatorToken = $env:OPERATOR_TOKEN }
if (-not $OperatorToken) {
    throw "OPERATOR_TOKEN is required for release acceptance."
}

$ApiBaseUrl = $ApiBaseUrl.TrimEnd("/")
$WebBaseUrl = $WebBaseUrl.TrimEnd("/")

$Live = Invoke-RestMethod -Uri "$ApiBaseUrl/health/live" -TimeoutSec 20
$Ready = Invoke-RestMethod -Uri "$ApiBaseUrl/health/ready" -TimeoutSec 20
$OpenApi = Invoke-RestMethod -Uri "$ApiBaseUrl/openapi.json" -TimeoutSec 20
if ($Live.status -ne "alive" -or $Ready.status -ne "ready") {
    throw "Azure API health probes did not return live and ready."
}
if ($OpenApi.openapi -ne "3.1.0" -or $OpenApi.info.version -ne "0.3.0") {
    throw "The public Azure OpenAPI contract is not the expected 0.3.0 release."
}

$Web = Invoke-WebRequest -Uri $WebBaseUrl -UseBasicParsing -TimeoutSec 20
if ($Web.StatusCode -ne 200 -or $Web.Content -notmatch "World Model Rollout Viewer") {
    throw "The public Azure Nuxt release shell is not reachable."
}

& "$PSScriptRoot\smoke-test.ps1" `
    -ApiBaseUrl $ApiBaseUrl `
    -WebBaseUrl $WebBaseUrl `
    -OperatorToken $OperatorToken
& "$PSScriptRoot\smoke-test-sprint-02.ps1" `
    -ApiBaseUrl $ApiBaseUrl `
    -WebBaseUrl $WebBaseUrl `
    -OperatorToken $OperatorToken
& "$PSScriptRoot\smoke-test-sprint-03.ps1" `
    -ApiBaseUrl $ApiBaseUrl `
    -WebBaseUrl $WebBaseUrl `
    -OperatorToken $OperatorToken

Write-Host "OK - Azure + Neon release acceptance passed" -ForegroundColor Green
Write-Host "Evidence: HTTPS/WSS same-origin app | Neon persistence | transient Redis Streams | three real profiles"
