Write-Output ((
    Invoke-Command -ScriptBlock {
        Get-ChildItem -File -Path .\internal\windows -Recurse -Filter *.go
        Get-ChildItem -File -Path .\cmd\khypervprovider -Recurse -Filter *.go
    } | ForEach-Object {
        (Resolve-Path -Relative $_.FullName).Replace("\", "/")
    }
) -join " ")
