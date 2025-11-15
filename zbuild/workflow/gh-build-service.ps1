$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

. (Join-Path $PSScriptRoot gh-funcs.ps1)

$repoDirectory = & git rev-parse --show-toplevel

try {
    Push-Location $repoDirectory
    Write-Log "Building Windows Service binary"

    # Build module
    zbuild\workflow\gh-build-module.ps1

    # Buid binary
    $goMod = (Get-Content .\go.mod | Select-String 'module') -split ' ' | Select-Object -Last 1
    $buildDate = (Get-Date).ToUniversalTime().ToString('ddd MMM dd HH:mm:ss UTC yyyy')
    $commitHash = & git rev-parse --short HEAD
    $version = Get-Version
    $binTarget =  "khypervprovider.exe"

    go build -o $binTarget -ldflags "-s -w -X $goMod/internal/common.Version=$version -X $goMod/internal/common.CommitHash=$commitHash -X '$goMod/internal/common.BuildDate=$buildDate'" ./cmd/khypervprovider
    Write-Output "ARTIFACT=$binTarget" >> $env:GITHUB_ENV
}
finally {
    Pop-Location
}

