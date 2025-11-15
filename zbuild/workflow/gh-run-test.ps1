$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

. (Join-Path $PSScriptRoot gh-funcs.ps1)

$repoDirectory = & git rev-parse --show-toplevel

try {
    Push-Location $repoDirectory

    Write-Log "PowerShell Version $($PSVersionTable.PSVersion)"
    $version = Get-Version

    # Build module
    zbuild\workflow\gh-build-module.ps1

    # Install module
    Write-Log "Installing module"
    cmd\khypervprovider\psmodule\install-module.ps1 -Package "cmd\khypervprovider\psmodule\khyperv-csi.$($version).nupkg" -CurrentUser

    # Tests
    Write-Log "Running tests"
    go test -timeout 1m -v ./...
}
finally {
    Pop-Location
}


