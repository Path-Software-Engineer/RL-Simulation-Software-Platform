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
$Swagger = Invoke-WebRequest -Uri "$ApiBaseUrl/docs/" -UseBasicParsing -TimeoutSec 20
if ($Live.status -ne "alive" -or $Ready.status -ne "ready") {
    throw "Azure API health probes did not return live and ready."
}
if ($OpenApi.openapi -ne "3.1.0" -or $OpenApi.info.version -ne "0.3.0") {
    throw "The public Azure OpenAPI contract is not the expected 0.3.0 release."
}
if ($Swagger.StatusCode -ne 200 -or $Swagger.Content -notmatch "SwaggerUIBundle") {
    throw "The public Azure Swagger UI is not reachable."
}

$PublicEnvironments = @(
    Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/environments" -TimeoutSec 20 |
        Write-Output
)
$PublicRuns = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/training-runs?limit=1" -TimeoutSec 20
if ($PublicEnvironments.Count -lt 1 -or @($PublicRuns.items | Write-Output).Count -lt 1) {
    throw "The public read-only portfolio API returned no persisted evidence."
}

$OperatorSession = Invoke-RestMethod `
    -Uri "$ApiBaseUrl/api/v1/operator/session" `
    -Headers @{ Authorization = "Bearer $OperatorToken" } `
    -TimeoutSec 20
if ($OperatorSession.role -ne "operator") {
    throw "The protected operator session check did not accept the release credential."
}

$UnauthenticatedStatus = 0
try {
    Invoke-WebRequest `
        -Uri "$ApiBaseUrl/api/v1/training-runs" `
        -Method Post `
        -UseBasicParsing `
        -ContentType "application/json" `
        -Body "{}" `
        -TimeoutSec 20 | Out-Null
} catch {
    $UnauthenticatedStatus = [int]$_.Exception.Response.StatusCode
}
if ($UnauthenticatedStatus -ne 401) {
    throw "An unauthenticated training command was not rejected with HTTP 401."
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
    -OperatorToken $OperatorToken `
    -TrainingTimeoutSeconds 300
& "$PSScriptRoot\smoke-test-sprint-03.ps1" `
    -ApiBaseUrl $ApiBaseUrl `
    -WebBaseUrl $WebBaseUrl `
    -OperatorToken $OperatorToken

Write-Host "OK - Azure + Neon release acceptance passed" -ForegroundColor Green
Write-Host "Evidence: public UI/Swagger/reads | protected commands | Neon | WSS | three real profiles"
