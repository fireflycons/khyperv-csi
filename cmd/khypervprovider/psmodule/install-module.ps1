param (
    [string]$Package,
    [switch]$Remove,
    [switch]$CurrentUser
)

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

function Write-Log {
    param (
        [string]$Message
    )

    Write-Host ([Datetime]::Now.ToString("yyyy/MM/dd HH:mm:ss")) "-" $Message
}

if (-not ($Package -or $Remove.IsPresent)) {
    Write-Log "Nothing to do"
    return
}

$moduleName = "khyperv-csi"

Write-Log "Checking for existing installation..."
$installedModules = (Get-Module -ListAvailable | Where-Object { $_.Name -eq $moduleName })

if ($Remove.IsPresent) {
    if (-not $installedModules) {
        Write-Log "Module is not installed"
    } else {
        Write-Log "Uninstalling $moduleName..."
        Uninstall-Module -Name $moduleName -AllVersions -ErrorAction SilentlyContinue | Out-Null
    }

    return
}

if (-not (Test-Path -Path $Package)) {
	throw "Path $Package does not exist"
}

$repoName = [Guid]::NewGuid().ToString()
$publishDir = (Resolve-Path (Split-Path -Parent -Path $Package)).Path

if ($installedModules) {
    Write-Log "Uninstalling existing $moduleName module"
    Remove-Module  -Name $moduleName -ErrorAction SilentlyContinue | Out-Null
    Uninstall-Module -Name $moduleName -AllVersions -ErrorAction SilentlyContinue | Out-Null
}

try {
    if (-not (Test-Path -PathType Container $publishDir)) {
        New-Item -Path $publishDir -ItemType Directory -ErrorAction Stop | Out-Null
    }
    Write-Log "Creating temporary repo $repoName at $publishDir"
	Register-PSRepository -Name $repoName -SourceLocation $publishDir -InstallationPolicy Trusted -ErrorAction Stop | Out-Null
    Write-Log "Installing package $(Split-Path -Leaf $Package) from repo $repoName"
    if ($CurrentUser.IsPresent) {
        Install-Module -Name ((Split-Path -Leaf $Package) -split '\.' | Select-Object -First 1) -Repository $repoName -Force -ErrorAction Stop
    } else {
	    Install-Module -Scope AllUsers -Name ((Split-Path -Leaf $Package) -split '\.' | Select-Object -First 1) -Repository $repoName -Force -ErrorAction Stop
    }
} finally {
	if (Get-PSRepository -Name $repoName -ErrorAction SilentlyContinue) {
        Write-Log "Removing temporary repo $repoName"
		Unregister-PSRepository -Name $repoName -ErrorAction SilentlyContinue | Out-Null
	}
}