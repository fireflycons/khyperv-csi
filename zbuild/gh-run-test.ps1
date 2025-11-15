$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

function Write-Log {
    param (
        [string]$Message
    )

    Write-Host ([Datetime]::Now.ToString("yyyy/MM/dd HH:mm:ss")) "-" $Message
}

$repoDirectory = & git rev-parse --show-toplevel

try {
    Push-Location $repoDirectory

    Write-Log "PowerShell Version $($PSVersionTable.PSVersion)"
    $version = Get-Content '.\_VERSION .txt' | Select-Object -First 1

    # Build module
    Write-Log "Building module"
    $module = "cmd\khypervprovider\psmodule\$($version).nupkg"
    Remove-Item -Force cmd\khypervprovider\psmodule\*.nupkg
    powershell-modules\build-module.ps1 -Version $version $module
    go generate ./cmd/khypervprovider/psmodule

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


