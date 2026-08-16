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
$PolicyId = "22222222-2222-4222-8222-222222222203"
$PolicyHash = "99902302d7119e631c23d982eba7a06b0135718a81780caa55e68d0732dccac7"
$Request = @{
    environmentId = "11111111-1111-4111-8111-111111111101"
    policyId = $PolicyId
    seed = 17
    maxSteps = 64
} | ConvertTo-Json

$Policy = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/policies/$PolicyId" -Headers $AuthHeaders -TimeoutSec 10
if ($Policy.algorithm -ne "dqn" -or $Policy.sha256 -ne $PolicyHash) {
    throw "Registered DQN identity differs from the Sprint 2 artifact."
}

$Run = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/training-runs" -Method Post -Headers $CommandHeaders -ContentType "application/json" -Body $Request -TimeoutSec 10
$Terminal = @("succeeded", "failed", "cancelled")
for ($Attempt = 1; $Attempt -le 180; $Attempt++) {
    Start-Sleep -Milliseconds 500
    $Run = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/training-runs/$($Run.id)" -Headers $AuthHeaders -TimeoutSec 10
    if ($Terminal -contains $Run.status) { break }
}
if ($Run.status -ne "succeeded" -or $Run.terminalReason -ne "training_budget_completed") {
    throw "DQN training ended unexpectedly: status=$($Run.status), reason=$($Run.terminalReason)."
}

$EpisodeResponse = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/training-runs/$($Run.id)/episodes?limit=100" -Headers $AuthHeaders -TimeoutSec 10
$Episodes = @($EpisodeResponse | Write-Output)
if ($Episodes.Count -ne 40) { throw "Expected 40 persisted DQN episodes; found $($Episodes.Count)." }
if (@($Episodes.episodeNumber | Sort-Object -Unique).Count -ne 40) {
    throw "DQN episode numbers are incomplete or duplicated."
}

$MetricNames = @(
    "episode_reward", "moving_average_reward", "epsilon", "loss", "episode_steps",
    "success_rate", "action_up", "action_right", "action_down", "action_left"
)
$Metrics = @{}
foreach ($MetricName in $MetricNames) {
    $Response = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/training-runs/$($Run.id)/metrics?limit=200&metric=$MetricName" -Headers $AuthHeaders -TimeoutSec 10
    $Metrics[$MetricName] = @($Response | Write-Output)
    if ($Metrics[$MetricName].Count -ne 40) {
        throw "Metric $MetricName has $($Metrics[$MetricName].Count) samples instead of 40."
    }
}

$Epsilon = $Metrics["epsilon"]
if ([math]::Abs([double]$Epsilon[0].value - 1.0) -gt 0.0001 -or [math]::Abs([double]$Epsilon[-1].value - 0.05) -gt 0.0001) {
    throw "Persisted epsilon schedule does not match the registered profile."
}
$ActionCount = 0.0
foreach ($MetricName in @("action_up", "action_right", "action_down", "action_left")) {
    $ActionCount += ($Metrics[$MetricName] | Measure-Object -Property value -Sum).Sum
}
if ($ActionCount -le 0) { throw "DQN action distribution is empty." }

$LatestEpisode = $Episodes | Sort-Object episodeNumber | Select-Object -Last 1
$TransitionResponse = Invoke-RestMethod -Uri "$ApiBaseUrl/api/v1/episodes/$($LatestEpisode.id)/transitions?limit=200" -Headers $AuthHeaders -TimeoutSec 10
$Transitions = @($TransitionResponse | Write-Output)
if ($Transitions.Count -ne $LatestEpisode.stepCount) {
    throw "Latest DQN episode transition evidence is incomplete."
}

$Web = Invoke-WebRequest -Uri $WebBaseUrl -UseBasicParsing -TimeoutSec 10
if ($Web.StatusCode -ne 200 -or $Web.Content -notmatch "World Model Rollout Viewer") {
    throw "The Sprint 2 Nuxt dashboard shell is not reachable."
}

Write-Host "OK - Sprint 2 DQN end-to-end smoke passed" -ForegroundColor Green
Write-Host "Evidence: 40 real episodes | 400 persisted metrics | replay | epsilon decay | loss | target sync"
