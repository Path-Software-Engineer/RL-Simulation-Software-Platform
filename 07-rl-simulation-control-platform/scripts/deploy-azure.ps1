[CmdletBinding()]
param(
    [ValidateLength(2, 18)]
    [ValidatePattern('^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$')]
    [string]$NamePrefix = "p7rl",
    [string]$ResourceGroup = "rg-p7-rl-simulation-demo",
    [string]$Location = "centralus",
    [string]$ImageTag = "",
    [switch]$SkipFoundation,
    [switch]$SkipBuild,
    [switch]$SkipSmoke
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $Root

function Invoke-AzureCli {
    param([Parameter(Mandatory)][string[]]$Arguments)
    $Output = & az @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Azure CLI operation failed. Review the preceding Azure error."
    }
    return ($Output -join [Environment]::NewLine)
}

function Read-RequiredSecret {
    param([Parameter(Mandatory)][string]$Name)
    $Value = [Environment]::GetEnvironmentVariable($Name, "Process")
    if ([string]::IsNullOrWhiteSpace($Value)) {
        throw "$Name must be set in the current PowerShell process. Do not paste it into chat or commit it."
    }
    return $Value.Trim()
}

function Push-DockerImage {
    param(
        [Parameter(Mandatory)][string]$Image,
        [ValidateRange(1, 10)][int]$MaxAttempts = 4
    )
    for ($Attempt = 1; $Attempt -le $MaxAttempts; $Attempt++) {
        docker push $Image
        if ($LASTEXITCODE -eq 0) {
            return
        }
        if ($Attempt -eq $MaxAttempts) {
            throw "Docker push failed after $MaxAttempts attempts: $Image"
        }
        $DelaySeconds = [Math]::Min(60, 10 * [Math]::Pow(2, $Attempt - 1))
        Write-Warning (
            "Docker push attempt $Attempt failed; retrying resumable upload in " +
            "$DelaySeconds seconds: $Image"
        )
        Start-Sleep -Seconds $DelaySeconds
    }
}

function Assert-NeonUrl {
    param(
        [Parameter(Mandatory)][string]$Name,
        [Parameter(Mandatory)][string]$Value,
        [Parameter(Mandatory)][bool]$Pooled
    )
    $Uri = $null
    if (-not [Uri]::TryCreate($Value, [UriKind]::Absolute, [ref]$Uri)) {
        throw "$Name is not a valid absolute PostgreSQL URL."
    }
    if ($Uri.Scheme -notin @("postgres", "postgresql") -or $Uri.Host -notmatch '\.neon\.tech$') {
        throw "$Name must be a Neon PostgreSQL connection URL."
    }
    if ($Value -notmatch '(?i)[?&]sslmode=(require|verify-full)(&|$)') {
        throw "$Name must require TLS through sslmode=require or sslmode=verify-full."
    }
    if ($Value -notmatch '(?i)[?&]channel_binding=require(&|$)') {
        throw "$Name must include channel_binding=require."
    }
    $IsPooledHost = $Uri.Host -match '-pooler\.'
    if ($Pooled -ne $IsPooledHost) {
        $Expected = if ($Pooled) { "the pooled -pooler endpoint" } else { "the direct non-pooler endpoint" }
        throw "$Name must use $Expected."
    }
    return $Uri
}

if (-not (Get-Command az -ErrorAction SilentlyContinue)) {
    throw "Azure CLI is required. Install it, open a new PowerShell, run 'az login', and retry."
}
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    throw "Git is required to derive an immutable release image tag."
}

$VersionInfo = Invoke-AzureCli @("version", "--output", "json") | ConvertFrom-Json
if ([version]$VersionInfo.'azure-cli' -lt [version]"2.75.0") {
    throw "Azure CLI 2.75.0 or newer is required for this release."
}

$Account = Invoke-AzureCli @("account", "show", "--output", "json") | ConvertFrom-Json
if (-not $Account.id) { throw "No active Azure subscription was found. Run 'az login'." }

$Dirty = @(git status --porcelain)
if ($LASTEXITCODE -ne 0) { throw "Git status failed." }
if ($Dirty.Count -gt 0) {
    throw "The release checkout must be clean so the image tag maps to an exact Git commit."
}
if (-not $ImageTag) {
    $ImageTag = (git rev-parse --short=12 HEAD).Trim()
    if ($LASTEXITCODE -ne 0 -or -not $ImageTag) { throw "Could not derive the Git image tag." }
}
if ($ImageTag -notmatch '^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,127}$') {
    throw "ImageTag is not a valid OCI tag."
}

$DatabaseUrl = Read-RequiredSecret "NEON_DATABASE_URL"
$DatabaseUrlDirect = Read-RequiredSecret "NEON_DATABASE_URL_DIRECT"
$OperatorToken = Read-RequiredSecret "OPERATOR_TOKEN"
$PooledNeonUri = Assert-NeonUrl -Name "NEON_DATABASE_URL" -Value $DatabaseUrl -Pooled $true
$DirectNeonUri = Assert-NeonUrl -Name "NEON_DATABASE_URL_DIRECT" -Value $DatabaseUrlDirect -Pooled $false
$PooledHostWithoutMarker = $PooledNeonUri.Host -replace '-pooler\.', '.'
$PooledUser = ($PooledNeonUri.UserInfo -split ':', 2)[0]
$DirectUser = ($DirectNeonUri.UserInfo -split ':', 2)[0]
if (
    $PooledHostWithoutMarker -ne $DirectNeonUri.Host -or
    $PooledNeonUri.AbsolutePath -ne $DirectNeonUri.AbsolutePath -or
    $PooledUser -ne $DirectUser
) {
    throw "The pooled and direct Neon URLs must target the same endpoint, database and role."
}
if ($OperatorToken.Length -lt 16 -or $OperatorToken.Length -gt 128 -or $OperatorToken -match '[\s,]') {
    throw "OPERATOR_TOKEN must contain 16-128 protocol-safe characters without whitespace or commas."
}

Write-Host "Azure release target: $($Account.name) / $Location" -ForegroundColor Cyan
Write-Host "Git image tag: $ImageTag" -ForegroundColor Cyan

$FoundationSource = Get-Content -LiteralPath "infra/azure/foundation.bicep" -Raw
$WorkloadSource = Get-Content -LiteralPath "infra/azure/workloads.bicep" -Raw
if ($FoundationSource -match 'Microsoft\.(Cache|OperationalInsights)/') {
    throw "Zero-cost guard failed: managed Redis and stored Log Analytics are forbidden."
}
if (
    $FoundationSource -notmatch "destination:\s*'azure-monitor'" -or
    $FoundationSource -notmatch "name:\s*'Standard'"
) {
    throw "Zero-cost guard failed: logs must have no storage destination and ACR must use the documented Standard grant."
}
if ($WorkloadSource -notmatch 'minReplicas:\s*0' -or $WorkloadSource -notmatch 'maxReplicas:\s*1') {
    throw "Zero-cost guard failed: the public app must scale from zero to at most one replica."
}
if ($WorkloadSource -match 'minReplicas:\s*[1-9]') {
    throw "Zero-cost guard failed: an always-on Container Apps replica was declared."
}
Write-Host "OK - zero-cost source guard passed" -ForegroundColor Green

Invoke-AzureCli @("config", "set", "extension.use_dynamic_install=yes_without_prompt", "--only-show-errors") | Out-Null
foreach ($Namespace in @("Microsoft.App", "Microsoft.ContainerRegistry", "Microsoft.ManagedIdentity")) {
    Invoke-AzureCli @("provider", "register", "--namespace", $Namespace, "--wait", "--only-show-errors") | Out-Null
}
Invoke-AzureCli @("bicep", "install", "--only-show-errors") | Out-Null
Invoke-AzureCli @("bicep", "build", "--file", "infra/azure/foundation.bicep", "--stdout", "--only-show-errors") | Out-Null
Invoke-AzureCli @("bicep", "build", "--file", "infra/azure/workloads.bicep", "--stdout", "--only-show-errors") | Out-Null

Invoke-AzureCli @(
    "group", "create",
    "--name", $ResourceGroup,
    "--location", $Location,
    "--tags", "project=rl-simulation-control-platform", "release=v1.0.0", "environment=demo", "managedBy=bicep",
    "--only-show-errors",
    "--output", "none"
) | Out-Null

$FoundationDeployment = "$NamePrefix-foundation"
if (-not $SkipFoundation) {
    Write-Host "Provisioning the free-grant ACR and scale-to-zero Container Apps environment" -ForegroundColor Cyan
    $FoundationJson = Invoke-AzureCli @(
        "deployment", "group", "create",
        "--name", $FoundationDeployment,
        "--resource-group", $ResourceGroup,
        "--template-file", "infra/azure/foundation.bicep",
        "--parameters", "location=$Location", "namePrefix=$NamePrefix",
        "--only-show-errors",
        "--query", "properties.outputs",
        "--output", "json"
    )
} else {
    $FoundationJson = Invoke-AzureCli @(
        "deployment", "group", "show",
        "--name", $FoundationDeployment,
        "--resource-group", $ResourceGroup,
        "--query", "properties.outputs",
        "--output", "json"
    )
}
$Foundation = $FoundationJson | ConvertFrom-Json
$RegistryName = $Foundation.registryName.value
$RegistryServer = $Foundation.registryLoginServer.value
$EnvironmentName = $Foundation.environmentName.value
$PullIdentityName = $Foundation.pullIdentityName.value
if (-not $RegistryName -or -not $RegistryServer -or -not $EnvironmentName -or -not $PullIdentityName) {
    throw "The foundation deployment did not return every required resource identity."
}

if (-not $SkipBuild) {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        throw "Docker is required to build release images locally."
    }
    docker info --format '{{.ServerVersion}}' 2>$null | Out-Null
    if ($LASTEXITCODE -ne 0) {
        throw "Docker Desktop Linux engine is not ready."
    }
    Invoke-AzureCli @(
        "acr", "login",
        "--name", $RegistryName,
        "--only-show-errors"
    ) | Out-Host

    $Images = @(
        @{ Name = "control-api"; Dockerfile = "infra/docker/control-api.Dockerfile" },
        @{ Name = "rl-runner"; Dockerfile = "infra/docker/rl-runner.Dockerfile" },
        @{ Name = "web"; Dockerfile = "infra/docker/web.Dockerfile" },
        @{ Name = "migrate"; Dockerfile = "infra/docker/migrate.Dockerfile" },
        @{ Name = "gateway"; Dockerfile = "infra/docker/gateway.Dockerfile" }
    )
    foreach ($Image in $Images) {
        $FullImage = "$RegistryServer/p7/$($Image.Name):$ImageTag"
        Write-Host "Building local image: $FullImage" -ForegroundColor Cyan
        docker build `
            --platform linux/amd64 `
            --file $Image.Dockerfile `
            --tag $FullImage `
            .
        if ($LASTEXITCODE -ne 0) {
            throw "Local Docker build failed for p7/$($Image.Name)."
        }
        Write-Host "Pushing immutable image: $FullImage" -ForegroundColor Cyan
        Push-DockerImage -Image $FullImage
    }
}

Write-Host "Deploying Container Apps workloads" -ForegroundColor Cyan
$env:AZURE_RELEASE_LOCATION = $Location
$env:AZURE_RELEASE_NAME_PREFIX = $NamePrefix
$env:AZURE_RELEASE_REGISTRY_NAME = $RegistryName
$env:AZURE_RELEASE_ENVIRONMENT_NAME = $EnvironmentName
$env:AZURE_RELEASE_PULL_IDENTITY_NAME = $PullIdentityName
$env:AZURE_RELEASE_IMAGE_TAG = $ImageTag
try {
    $WorkloadsJson = Invoke-AzureCli @(
        "deployment", "group", "create",
        "--name", "$NamePrefix-workloads",
        "--resource-group", $ResourceGroup,
        "--parameters", "infra/azure/workloads.bicepparam",
        "--only-show-errors",
        "--query", "properties.outputs",
        "--output", "json"
    )
}
finally {
    Remove-Item Env:AZURE_RELEASE_LOCATION -ErrorAction SilentlyContinue
    Remove-Item Env:AZURE_RELEASE_NAME_PREFIX -ErrorAction SilentlyContinue
    Remove-Item Env:AZURE_RELEASE_REGISTRY_NAME -ErrorAction SilentlyContinue
    Remove-Item Env:AZURE_RELEASE_ENVIRONMENT_NAME -ErrorAction SilentlyContinue
    Remove-Item Env:AZURE_RELEASE_PULL_IDENTITY_NAME -ErrorAction SilentlyContinue
    Remove-Item Env:AZURE_RELEASE_IMAGE_TAG -ErrorAction SilentlyContinue
}
$Workloads = $WorkloadsJson | ConvertFrom-Json
$MigrationJobName = $Workloads.migrationJobName.value
$AppName = $Workloads.appName.value
$ApiUrl = $Workloads.apiUrl.value
$WebUrl = $Workloads.webUrl.value
if (-not $MigrationJobName -or -not $AppName -or -not $ApiUrl -or -not $WebUrl) {
    throw "The workload deployment did not return its release endpoints."
}

$ScaleJson = Invoke-AzureCli @(
    "containerapp", "show",
    "--name", $AppName,
    "--resource-group", $ResourceGroup,
    "--query", "properties.template.scale.{min:minReplicas,max:maxReplicas}",
    "--output", "json",
    "--only-show-errors"
)
$Scale = $ScaleJson | ConvertFrom-Json
if ($Scale.min -ne 0 -or $Scale.max -ne 1) {
    throw "Zero-cost guard failed after deployment: Container Apps scale is not 0..1."
}
$ForbiddenResources = Invoke-AzureCli @(
    "resource", "list",
    "--resource-group", $ResourceGroup,
    "--query", "[?starts_with(type, 'Microsoft.Cache/') || starts_with(type, 'Microsoft.OperationalInsights/')].type",
    "--output", "tsv",
    "--only-show-errors"
)
if (-not [string]::IsNullOrWhiteSpace($ForbiddenResources)) {
    throw "Zero-cost guard failed after deployment: a forbidden billable resource exists."
}
Write-Host "OK - deployed resource guard passed (scale 0..1, no Managed Redis or Log Analytics)" -ForegroundColor Green

Write-Host "Applying versioned migrations to Neon" -ForegroundColor Cyan
$ExecutionName = (
    Invoke-AzureCli @(
        "containerapp", "job", "start",
        "--name", $MigrationJobName,
        "--resource-group", $ResourceGroup,
        "--query", "name",
        "--output", "tsv",
        "--only-show-errors"
    )
).Trim()
if (-not $ExecutionName) { throw "The migration job did not return an execution name." }

$MigrationStatus = ""
for ($Attempt = 1; $Attempt -le 90; $Attempt++) {
    $MigrationStatus = (
        Invoke-AzureCli @(
            "containerapp", "job", "execution", "show",
            "--name", $MigrationJobName,
            "--resource-group", $ResourceGroup,
            "--job-execution-name", $ExecutionName,
            "--query", "properties.status",
            "--output", "tsv",
            "--only-show-errors"
        )
    ).Trim()
    if ($MigrationStatus -in @("Succeeded", "Failed")) { break }
    Start-Sleep -Seconds 10
}
if ($MigrationStatus -ne "Succeeded") {
    throw "Neon migration execution ended with status '$MigrationStatus'. Inspect the Container Apps Job logs."
}

if (-not $SkipSmoke) {
    & "$PSScriptRoot\smoke-test-release.ps1" `
        -ApiBaseUrl $ApiUrl `
        -WebBaseUrl $WebUrl `
        -OperatorToken $OperatorToken
}

Write-Host "OK - Azure + Neon release deployment passed" -ForegroundColor Green
Write-Host "Web:     $WebUrl"
Write-Host "API:     $ApiUrl"
Write-Host "OpenAPI: $ApiUrl/openapi.json"

$DatabaseUrl = $null
$DatabaseUrlDirect = $null
$OperatorToken = $null
