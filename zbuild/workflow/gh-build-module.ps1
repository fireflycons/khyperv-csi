$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

. (Join-Path $PSScriptRoot gh-funcs.ps1)

$repoDirectory = & git rev-parse --show-toplevel

try {
    Push-Location $repoDirectory
    Write-Log "Building module"
    $module = "cmd\khypervprovider\psmodule\$($version).nupkg"
    $version = Get-Version
    Remove-Item -Force cmd\khypervprovider\psmodule\*.nupkg
    powershell-modules\build-module.ps1 -Version $version $module
    go generate ./cmd/khypervprovider/psmodule
}
finally {
    Pop-Location
}

