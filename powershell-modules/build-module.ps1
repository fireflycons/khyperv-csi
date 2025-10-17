param (
    [string]$Version = "1.0.0",
    [string]$Target
)

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

# Build the nuget package

New-Item -Path (Split-Path -Parent -Path $Target) -ItemType Directory -Force | Out-Null
Import-Module Microsoft.PowerShell.Utility -RequiredVersion 3.1.0.0 -Force -ErrorAction Stop
Get-Command Update-ModuleManifest | Out-Null
$repoName = [Guid]::NewGuid().ToString()
$artifactLocation = (Resolve-Path (Split-Path -Parent -Path $Target)).Path
$sourceLocation = Join-Path $PSScriptRoot 'khyperv-csi'
$manifest = Join-Path $sourceLocation 'khyperv-csi.psd1'
$moduleData = Import-PowerShellDataFile -Path $manifest
$moduleVersion = $moduleData['ModuleVersion']

if ($moduleVersion -ne $Version) {
    Update-ModuleManifest -Path $manifest -ModuleVersion $Version
    $moduleVersion = $Version
}

$nugetPackage = Join-Path $artifactLocation "$repoName.$moduleVersion.nupkg"

if (Test-Path $nugetPackage) {
    Remove-Item $nugetPackage -Force
}

Write-Host "Building $repoName.$moduleVersion.nupkg"


try {
    if (-not (Get-PSRepository -Name $repoName -ErrorAction SilentlyContinue)) {
        Write-Host "Creating temporary repository $repoName at $artifactLocation"
        Register-PSRepository -Name $repoName -SourceLocation $artifactLocation -PublishLocation $artifactLocation -InstallationPolicy Trusted  -ErrorAction Stop | Out-Null
    }

    Publish-Module -Path $sourceLocation -Repository $repoName -NuGetApiKey 'dummy' -Force -ErrorAction Stop
}
finally {
    if (Get-PSRepository -Name $repoName -ErrorAction SilentlyContinue) {
        Unregister-PSRepository -Name $repoName
        Write-Host "Removed temporary repository $repoName"
    }
}

# Emit PowerShell code to install/remove the module
