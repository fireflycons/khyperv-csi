Write-Output ((
    Get-ChildItem -Recurse -File -Filter *.ps* -Path powershell-modules |
        ForEach-Object {
            (Resolve-Path -Relative $_.FullName).Replace("\", "/")
        }
) -join " ")