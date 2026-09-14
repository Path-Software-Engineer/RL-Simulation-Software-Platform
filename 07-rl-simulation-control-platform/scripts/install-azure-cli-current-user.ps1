[CmdletBinding()]
param(
    [ValidatePattern('^\d+\.\d+\.\d+$')]
    [string]$Version = "2.89.1"
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

$installRoot = Join-Path $env:LOCALAPPDATA "Programs\AzureCLI\$Version"
$archivePath = Join-Path $env:TEMP "azure-cli-$Version-x64.zip"
$downloadUrl = "https://azcliprod.blob.core.windows.net/zip/azure-cli-$Version-x64.zip"
$expectedCommand = Join-Path $installRoot "bin\az.cmd"

if (-not (Test-Path -LiteralPath $expectedCommand -PathType Leaf)) {
    Write-Host "Downloading the official Azure CLI $Version ZIP..."
    New-Item -ItemType Directory -Path $installRoot -Force | Out-Null
    Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath

    Write-Host "Installing Azure CLI for the current user..."
    Expand-Archive -LiteralPath $archivePath -DestinationPath $installRoot -Force
}

$azureCommand = Get-ChildItem -LiteralPath $installRoot -Filter "az.cmd" -File -Recurse |
    Select-Object -First 1
if ($null -eq $azureCommand) {
    throw "Azure CLI installation is incomplete: az.cmd was not found under $installRoot"
}

$azureBin = $azureCommand.Directory.FullName
$processEntries = @($env:Path -split ";" | Where-Object { $_ })
if ($processEntries -notcontains $azureBin) {
    $env:Path = (@($azureBin) + $processEntries) -join ";"
}

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$userEntries = @($userPath -split ";" | Where-Object { $_ })
if ($userEntries -notcontains $azureBin) {
    [Environment]::SetEnvironmentVariable(
        "Path",
        (@($userEntries) + $azureBin) -join ";",
        "User"
    )
}

& $azureCommand.FullName version --output json | Out-Null
if ($LASTEXITCODE -ne 0) {
    throw "Azure CLI was installed but its version check failed."
}

Write-Host "OK - Azure CLI $Version is installed for the current user"
Write-Host "Command: $($azureCommand.FullName)"
Write-Host "Open a new PowerShell window before running az from PATH."
