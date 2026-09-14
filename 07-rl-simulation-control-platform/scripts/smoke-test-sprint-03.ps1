[CmdletBinding()]
param(
    [string]$ApiBaseUrl = "http://127.0.0.1:8080",
    [string]$WebBaseUrl = "http://127.0.0.1:3000",
    [string]$OperatorToken = ""
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
if (-not $OperatorToken) {
    if ($env:OPERATOR_TOKEN) {
        $OperatorToken = $env:OPERATOR_TOKEN
    } else {
        $EnvironmentFile = Join-Path $Root ".env"
        if (Test-Path -LiteralPath $EnvironmentFile) {
            $ConfiguredToken = Get-Content -LiteralPath $EnvironmentFile |
                Where-Object { $_ -match '^OPERATOR_TOKEN=' } |
                Select-Object -First 1
            if ($ConfiguredToken) {
                $OperatorToken = $ConfiguredToken.Substring("OPERATOR_TOKEN=".Length).Trim()
            }
        }
    }
}
if (-not $OperatorToken) { $OperatorToken = "local-operator-token-change-me" }

$AuthHeaders = @{ "Authorization" = "Bearer $OperatorToken" }
$CommandHeaders = @{
    "Authorization" = "Bearer $OperatorToken"
    "Idempotency-Key" = [guid]::NewGuid().ToString()
}
$PolicyId = "22222222-2222-4222-8222-222222222204"
$PolicyHash = "7b962a91ea1b9a833e1935413a9df2795b40c0533505a63d8487d16a193be8b2"
$Request = @{
    environmentId = "11111111-1111-4111-8111-111111111101"
    policyId = $PolicyId
    seed = 23
    maxSteps = 64
} | ConvertTo-Json

$Policy = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/policies/$PolicyId" -Headers $AuthHeaders -TimeoutSec 10
if ($Policy.algorithm -ne "world-model" -or $Policy.sha256 -ne $PolicyHash) {
    throw "Registered world-model identity differs from the Sprint 3 artifact."
}

$Run = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/training-runs" -Method Post -Headers $CommandHeaders -ContentType "application/json" -Body $Request -TimeoutSec 10
$Terminal = @("succeeded", "failed", "cancelled")
for ($Attempt = 1; $Attempt -le 120; $Attempt++) {
    Start-Sleep -Milliseconds 250
    $Run = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/training-runs/$($Run.id)" -Headers $AuthHeaders -TimeoutSec 10
    if ($Terminal -contains $Run.status) { break }
}
if ($Run.status -ne "succeeded" -or $Run.terminalReason -ne "rollout_budget_completed") {
    throw "World-model rollout ended unexpectedly: status=$($Run.status), reason=$($Run.terminalReason)."
}

$EpisodeResponse = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/training-runs/$($Run.id)/episodes?limit=100" -Headers $AuthHeaders -TimeoutSec 10
$Episodes = @($EpisodeResponse | Write-Output)
if ($Episodes.Count -ne 1 -or $Episodes[0].stepCount -ne 11 -or $Episodes[0].collisions -ne 1) {
    throw "Expected one 11-step rollout with one real collision."
}

$TransitionResponse = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/episodes/$($Episodes[0].id)/transitions?limit=200" -Headers $AuthHeaders -TimeoutSec 10
$Transitions = @($TransitionResponse | Write-Output)
if ($Transitions.Count -ne 11) { throw "Expected 11 persisted rollout transitions." }
if (@($Transitions | Where-Object { -not $_.predictedState -or -not $_.predictedNextState }).Count -ne 0) {
    throw "Predicted state evidence is incomplete."
}
$Errors = @($Transitions | ForEach-Object { [double]$_.stepError })
if ($Errors[0] -ne 0 -or $Errors[1] -ne 1 -or $Errors[-1] -ne 0) {
    throw "Per-step rollout error differs from the registered comparison."
}
if ([double]$Transitions[-1].accumulatedError -ne 9) {
    throw "Accumulated rollout error differs from the registered comparison."
}

$MetricCounts = @{
    training_examples = 1
    prediction_error = 11
    accumulated_error = 11
    rollout_risk = 11
}
foreach ($Entry in $MetricCounts.GetEnumerator()) {
    $Response = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/training-runs/$($Run.id)/metrics?limit=200&metric=$($Entry.Key)" -Headers $AuthHeaders -TimeoutSec 10
    $Items = @($Response | Write-Output)
    if ($Items.Count -ne $Entry.Value) {
        throw "Metric $($Entry.Key) has $($Items.Count) samples instead of $($Entry.Value)."
    }
}

$Web = Invoke-WebRequest -Uri $WebBaseUrl -UseBasicParsing -TimeoutSec 10
if ($Web.StatusCode -ne 200 -or $Web.Content -notmatch "World Model Rollout Viewer") {
    throw "The Sprint 3 Nuxt rollout viewer shell is not reachable."
}

Write-Host "OK - Sprint 3 world-model end-to-end smoke passed" -ForegroundColor Green
Write-Host "Evidence: 12 examples | 11 real/predicted transitions | divergence step 2 | 9.00 accumulated error"
