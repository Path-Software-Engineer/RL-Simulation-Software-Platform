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
if (-not $OperatorToken) {
    $OperatorToken = "local-operator-token-change-me"
}
$AuthHeaders = @{ "Authorization" = "Bearer $OperatorToken" }
$CommandHeaders = @{
    "Authorization" = "Bearer $OperatorToken"
    "Idempotency-Key" = [guid]::NewGuid().ToString()
}
$Request = @{
    environmentId = "11111111-1111-4111-8111-111111111101"
    policyId = "22222222-2222-4222-8222-222222222201"
    seed = 7
    maxSteps = 64
} | ConvertTo-Json

$Ready = Invoke-RestMethod -Uri "$ApiBaseUrl/health/ready" -TimeoutSec 10
if ($Ready.status -ne "ready") { throw "API dependencies are not ready." }

$Environment = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/environments/11111111-1111-4111-8111-111111111101" -Headers $AuthHeaders -TimeoutSec 10
$Policy = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/policies/22222222-2222-4222-8222-222222222201" -Headers $AuthHeaders -TimeoutSec 10
if ($Environment.version -ne "1.0.0" -or $Policy.sha256 -ne "22d5faf9a94fcd05fdf31d2a1429a1b8f02d4f394f6cb61b95cc26351060927f") {
    throw "Registered environment or policy identity differs from Sprint 1 evidence."
}

$Run = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/training-runs" -Method Post -Headers $CommandHeaders -ContentType "application/json" -Body $Request -TimeoutSec 10
$Terminal = @("succeeded", "failed", "cancelled")
for ($Attempt = 1; $Attempt -le 30; $Attempt++) {
    Start-Sleep -Milliseconds 500
    $Run = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/training-runs/$($Run.id)" -Headers $AuthHeaders -TimeoutSec 10
    if ($Terminal -contains $Run.status) { break }
}
if ($Run.status -ne "succeeded") { throw "Controlled run ended in unexpected state: $($Run.status)." }

$EpisodeResponse = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/training-runs/$($Run.id)/episodes?limit=20" -Headers $AuthHeaders -TimeoutSec 10
$Episodes = @($EpisodeResponse | Write-Output)
if ($Episodes.Count -ne 1) { throw "Expected exactly one persisted episode." }
$Episode = $Episodes[0]
$TransitionResponse = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/episodes/$($Episode.id)/transitions?limit=200" -Headers $AuthHeaders -TimeoutSec 10
$Transitions = @($TransitionResponse | Write-Output)
$MetricResponse = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/training-runs/$($Run.id)/metrics?limit=50" -Headers $AuthHeaders -TimeoutSec 10
$Metrics = @($MetricResponse | Write-Output)

if ($Episode.stepCount -ne 10 -or $Episode.collisions -ne 0 -or [math]::Round($Episode.totalReward, 2) -ne 9.64) {
    throw "Persisted episode evidence differs from the deterministic oracle."
}
$TransitionsComplete = $Transitions.Count -eq 10
for ($Index = 0; $TransitionsComplete -and $Index -lt $Transitions.Count; $Index++) {
    $Transition = $Transitions[$Index]
    $ExpectedTerminal = $Index -eq $Transitions.Count - 1
    $TransitionsComplete = (
        $Transition.stepIndex -eq $Index -and
        [bool]$Transition.terminated -eq $ExpectedTerminal -and
        -not [bool]$Transition.truncated
    )
}
if (-not $TransitionsComplete) {
    $LastTransition = $Transitions | Select-Object -Last 1
    throw "Transition evidence is incomplete: count=$($Transitions.Count), lastStep=$($LastTransition.stepIndex), terminated=$($LastTransition.terminated), truncated=$($LastTransition.truncated)."
}
if ($Metrics.Count -ne 3) { throw "Expected exactly three bounded metric samples." }

$Web = Invoke-WebRequest -Uri $WebBaseUrl -UseBasicParsing -TimeoutSec 10
if ($Web.StatusCode -ne 200 -or $Web.Content -notmatch "Gridworld Agent Visualizer") {
    throw "The Nuxt application shell is not reachable."
}

Write-Host "OK - Sprint 1 end-to-end smoke passed" -ForegroundColor Green
Write-Host "Evidence: 1 real episode | 10 transitions | 9.64 reward | 0 collisions | goal reached"
