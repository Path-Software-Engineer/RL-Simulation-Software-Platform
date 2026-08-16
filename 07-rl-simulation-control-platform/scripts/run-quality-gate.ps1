[CmdletBinding()]
param([switch]$KeepRunning)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

function Invoke-Gate([string]$Label, [scriptblock]$Action) {
    Write-Host "  -> $Label" -ForegroundColor Cyan
    & $Action
    if ($LASTEXITCODE -ne 0) { throw "$Label failed." }
}

function Resolve-Python {
    $Command = Get-Command python -ErrorAction SilentlyContinue
    if ($Command) { return $Command.Source }
    $Bundled = "C:\Users\Asus\.cache\codex-runtimes\codex-primary-runtime\dependencies\python\python.exe"
    if (Test-Path -LiteralPath $Bundled) { return $Bundled }
    throw "Python is required for repository validators."
}

function Assert-DockerReady {
    docker info --format '{{.ServerVersion}}' 2>$null | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Docker Desktop Linux engine is not available. Start Docker Desktop and wait until the engine is running, then retry the quality gate."
    }
}

$Python = Resolve-Python
$AsyncApiCliImage = "asyncapi/cli:6.0.2@sha256:8d74506cc69f650b7b165988221d8b54af5f8cc476a8fdafbde652816f7ace53"
$Started = $false
try {
    Assert-DockerReady
    Write-Host "[1/8] Validating versioned contracts and artifacts"
    Invoke-Gate "OpenAPI / AsyncAPI / JSON Schema" { & $Python scripts/validate-contracts.py }
    Invoke-Gate "Artifact identity and SHA-256" { & $Python scripts/verify-artifacts.py }
    Invoke-Gate "Direct deterministic runner" { & $Python scripts/direct-runner-check.py }
    Invoke-Gate "Direct bounded DQN trainer" { & $Python scripts/direct-dqn-check.py }
    Invoke-Gate "Direct bounded world-model rollout" { & $Python scripts/direct-world-model-check.py }
    Invoke-Gate "Repository UTF-8, whitespace and secret hygiene" { & $Python scripts/check-repository-hygiene.py }
    Invoke-Gate "Python syntax" { & $Python -m compileall -q workers/rl-runner/src scripts }

    Write-Host "[2/8] Running formal OpenAPI, AsyncAPI and JSON Schema validators"
    Invoke-Gate "OpenAPI 3.1 and JSON Schema 2020-12" {
        docker run --rm -v "${Root}:/workspace" -w /workspace python:3.12.13-slim-bookworm sh -c "python -m pip install -q -r scripts/requirements-contracts.lock && python scripts/validate-formal-contracts.py"
    }
    Invoke-Gate "AsyncAPI 3.0" {
        docker run --rm `
            --platform linux/amd64 `
            --user root `
            -v "${Root}:/workspace:ro" `
            $AsyncApiCliImage `
            validate /workspace/contracts/events/asyncapi.json --suppressAllWarnings
    }

    Write-Host "[3/8] Validating the Python runner in its pinned container"
    Invoke-Gate "Ruff and pytest" {
        docker run --rm -v "${Root}:/workspace" -w /workspace/workers/rl-runner python:3.12.13-slim-bookworm sh -c "python -m pip install -q -r requirements.lock -r requirements-dev.lock && ruff check src tests /workspace/scripts && pytest"
    }

    Write-Host "[4/8] Validating the Go control API"
    Invoke-Gate "gofmt" {
        $UnformattedFiles = @(
            docker run --rm `
                -v "${Root}:/workspace" `
                -w /workspace/services/control-api `
                golang:1.26.5-alpine `
                gofmt -l .
        )
        if ($LASTEXITCODE -ne 0) { throw "gofmt container execution failed." }
        if ($UnformattedFiles.Count -gt 0) {
            $UnformattedFiles | ForEach-Object { Write-Host "    $_" -ForegroundColor Yellow }
            throw "gofmt found unformatted Go files."
        }
    }
    Invoke-Gate "go test" {
        docker run --rm `
            -v "${Root}:/workspace" `
            -w /workspace/services/control-api `
            golang:1.26.5-alpine `
            go test ./...
    }

    Write-Host "[5/8] Typechecking, testing and building Nuxt"
    Invoke-Gate "Nuxt build image" { docker compose build web }

    Write-Host "[6/8] Validating Docker Compose and application images"
    Invoke-Gate "Docker Compose config" { docker compose config --quiet }
    Invoke-Gate "Docker Compose build" { docker compose build control-api rl-runner }

    Write-Host "[7/8] Running the real cross-layer acceptance flow"
    Invoke-Gate "Start healthy platform" { docker compose up --detach --wait --wait-timeout 240 }
    $Started = $true
    & "$PSScriptRoot\smoke-test.ps1"
    & "$PSScriptRoot\smoke-test-sprint-02.ps1"
    & "$PSScriptRoot\smoke-test-sprint-03.ps1"

    Write-Host "[8/8] Checking repository hygiene"
    Invoke-Gate "Git whitespace" {
        git diff --check
        if ($LASTEXITCODE -ne 0) { throw "Unstaged Git whitespace check failed." }
        git diff --cached --check
        if ($LASTEXITCODE -ne 0) { throw "Staged Git whitespace check failed." }
    }
    Invoke-Gate "Repository hygiene recheck" { & $Python scripts/check-repository-hygiene.py }
    Write-Host "OK - Project 07 Sprint 3 quality gate passed" -ForegroundColor Green
} finally {
    if ($Started -and -not $KeepRunning) { docker compose down --remove-orphans | Out-Null }
}
